package main

import (
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestErrorPages(t *testing.T) {
	handler := newHandler()
	for _, code := range []int{400, 401, 403, 404, 429, 500, 502, 503, 504, 599} {
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest("GET", "/errors/"+strconv.Itoa(code), nil))
			if w.Code != code || !strings.Contains(w.Body.String(), ">"+strconv.Itoa(code)+"</p>") {
				t.Fatalf("status/body mismatch: %d", w.Code)
			}
			if w.Header().Get("Content-Type") != "text/html; charset=utf-8" || w.Header().Get("Cache-Control") != "no-store" || !strings.Contains(w.Header().Get("Content-Security-Policy"), "default-src 'none'") {
				t.Fatal("missing response protections")
			}
			if w.Header().Get("Content-Length") != strconv.Itoa(w.Body.Len()) {
				t.Fatal("incorrect content length")
			}
		})
	}
}

func TestUntrustedInputs(t *testing.T) {
	handler := newHandler()
	for _, path := range []string{"/", "/missing", "/errors/200", "/errors/999", "/errors/403/extra", "/errors/+403", "/errors/0403", "/?code=403&message=%3Cscript%3E"} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", path, nil)
		r.Host = "attacker.example"
		r.Header.Set("X-Code", "403")
		r.Header.Set("Cookie", "sensitive-value")
		handler.ServeHTTP(w, r)
		if w.Code != 404 || strings.Contains(w.Body.String(), "attacker") || strings.Contains(w.Body.String(), "sensitive-value") || strings.Contains(w.Body.String(), "<script>") {
			t.Fatalf("unexpected handling of %s", path)
		}
	}
}

func TestMethodsAndHealth(t *testing.T) {
	handler := newHandler()
	for _, tc := range []struct {
		method, path string
		status       int
	}{
		{"HEAD", "/errors/403", 403},
		{"GET", "/healthz", 204},
		{"POST", "/errors/403", 405},
	} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code != tc.status || w.Body.Len() != 0 {
			t.Fatalf("%s %s: unexpected status/body", tc.method, tc.path)
		}
	}
}
