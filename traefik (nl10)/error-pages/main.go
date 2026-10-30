package main

import (
	"bytes"
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//go:embed error.html
var pageHTML string

// Render once at startup: requests only write an immutable, pre-rendered page.
func newHandler() http.Handler {
	tpl := template.Must(template.New("error").Parse(pageHTML))
	pages := make(map[int][]byte)
	for code := 400; code <= 599; code++ {
		title := http.StatusText(code)
		if title == "" {
			title = "Request failed"
		}
		message := "We couldn’t complete this request. Please try again later."
		switch code {
		case 401:
			message = "Please sign in to access this service."
		case 403:
			message = "You don’t have permission to access this service."
		case 404:
			message = "We couldn’t find this page. Please check the address."
		case 429:
			message = "Too many requests. Please wait a moment and try again."
		case 502, 503, 504:
			message = "This service is temporarily unavailable. Please try again later."
		}
		var body bytes.Buffer
		if err := tpl.Execute(&body, struct {
			Code           int
			Title, Message string
		}{code, title, message}); err != nil {
			panic(err)
		}
		pages[code] = body.Bytes()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'")
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// Only this exact internal endpoint is healthy; all other paths default to 404.
		if r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		code := http.StatusNotFound
		if value, ok := strings.CutPrefix(r.URL.Path, "/errors/"); ok && len(value) == 3 {
			if parsed, err := strconv.Atoi(value); err == nil && parsed >= 400 && parsed <= 599 {
				code = parsed
			}
		}
		body := pages[code]
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeader(code)
		if r.Method != http.MethodHead {
			_, _ = w.Write(body)
		}
	})
}

func main() {
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	server := &http.Server{
		Addr: addr, Handler: newHandler(),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    8192,
	}
	// No access logs: paths, cookies and other request data are never recorded.
	log.Fatal(server.ListenAndServe())
}
