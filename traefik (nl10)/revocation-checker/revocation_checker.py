from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
import os
import time
from urllib.error import HTTPError, URLError
from urllib.parse import quote, unquote
from urllib.request import Request, urlopen


OPENBAO_CERT_BASE_URL = os.environ.get(
    "OPENBAO_CERT_BASE_URL",
    "https://secrets.cloud.northlake.dev/v1/pki_user_mtls/cert",
).rstrip("/")
CACHE_TTL_SECONDS = int(os.environ.get("CACHE_TTL_SECONDS", "30"))
FAIL_CLOSED = os.environ.get("FAIL_CLOSED", "true").lower() in {"1", "true", "yes"}
LISTEN_ADDR = os.environ.get("LISTEN_ADDR", "0.0.0.0")
LISTEN_PORT = int(os.environ.get("LISTEN_PORT", "8080"))

_cache = {}


def _normalize_serial(value):
    serial = value.strip().strip('"').replace("-", ":").lower()
    if ":" in serial:
        parts = [part.zfill(2) for part in serial.split(":") if part]
    else:
        if len(serial) % 2:
            serial = "0" + serial
        parts = [serial[index:index + 2] for index in range(0, len(serial), 2)]
    return ":".join(parts)


def _extract_quoted_field(header_value, field_name):
    marker = f'{field_name}="'
    start = header_value.find(marker)
    if start == -1:
        return ""

    start += len(marker)
    end = header_value.find('"', start)
    if end == -1:
        return ""

    return header_value[start:end]


def _extract_serial(headers):
    raw_info = headers.get("X-Forwarded-Tls-Client-Cert-Info", "")
    decoded_info = unquote(raw_info)
    serial = _extract_quoted_field(decoded_info, "SerialNumber")
    if serial:
        return _normalize_serial(serial)

    raw_pem = headers.get("X-Forwarded-Tls-Client-Cert", "")
    if raw_pem:
        raise ValueError("client cert PEM was present, but serial info was missing")

    raise ValueError("missing X-Forwarded-Tls-Client-Cert-Info serial")


def _read_openbao_cert(serial):
    now = time.time()
    cached = _cache.get(serial)
    if cached and cached["expires_at"] > now:
        return cached["result"]

    url = f"{OPENBAO_CERT_BASE_URL}/{quote(serial, safe=':')}"
    request = Request(url, headers={"Accept": "application/json"})

    with urlopen(request, timeout=5) as response:
        payload = json.loads(response.read().decode("utf-8"))

    data = payload.get("data") or {}
    result = {
        "known": bool(data.get("certificate")),
        "revocation_time": int(data.get("revocation_time") or 0),
        "revocation_time_rfc3339": data.get("revocation_time_rfc3339") or "",
        "issuer_id": data.get("issuer_id") or "",
    }
    _cache[serial] = {
        "expires_at": now + CACHE_TTL_SECONDS,
        "result": result,
    }
    return result


class RevocationHandler(BaseHTTPRequestHandler):
    server_version = "northlake-mtls-revocation-checker/1.0"

    def log_message(self, fmt, *args):
        print(f"{self.address_string()} - {fmt % args}", flush=True)

    def _send(self, status, body):
        payload = json.dumps(body, sort_keys=True).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)

    def do_GET(self):
        if self.path.startswith("/health"):
            self._send(200, {"ok": True})
            return

        if not self.path.startswith("/auth"):
            self._send(404, {"ok": False, "error": "not found"})
            return

        try:
            serial = _extract_serial(self.headers)
            cert = _read_openbao_cert(serial)

            if not cert["known"]:
                print(f"denied unknown client certificate serial={serial}", flush=True)
                self._send(403, {"ok": False, "serial": serial, "error": "unknown certificate"})
                return

            if cert["revocation_time"] > 0:
                print(
                    "denied revoked client certificate "
                    f"serial={serial} revocation_time={cert['revocation_time']}",
                    flush=True,
                )
                self._send(
                    403,
                    {
                        "ok": False,
                        "serial": serial,
                        "error": "certificate revoked",
                        "revocation_time": cert["revocation_time"],
                        "revocation_time_rfc3339": cert["revocation_time_rfc3339"],
                    },
                )
                return

            self._send(200, {"ok": True, "serial": serial})

        except HTTPError as exc:
            status = 403 if FAIL_CLOSED else 200
            print(f"denied OpenBao lookup failure status={exc.code}", flush=True)
            self._send(status, {"ok": not FAIL_CLOSED, "error": f"openbao http {exc.code}"})
        except (URLError, TimeoutError, ValueError, json.JSONDecodeError) as exc:
            status = 403 if FAIL_CLOSED else 200
            print(f"denied revocation check error={exc}", flush=True)
            self._send(status, {"ok": not FAIL_CLOSED, "error": str(exc)})


if __name__ == "__main__":
    print(
        f"listening on {LISTEN_ADDR}:{LISTEN_PORT}; OpenBao cert endpoint: {OPENBAO_CERT_BASE_URL}",
        flush=True,
    )
    ThreadingHTTPServer((LISTEN_ADDR, LISTEN_PORT), RevocationHandler).serve_forever()
