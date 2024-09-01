package xtproxy

import (
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hairyhenderson/go-fsimpl/httpfs"
	"github.com/spf13/afero"
)

type XTProxyHTTP struct {
	Fs       afero.Fs
	Listener *net.TCPListener
	server   *http.Server
}

func (m *XTProxyHTTP) Wait() error {
	if err := m.init(); err != nil {
		return err
	}
	return m.server.Serve(m.Listener)
}

func (m *XTProxyHTTP) init() error {
	if m.server != nil {
		return nil
	}
	httpFs := afero.NewHttpFs(m.Fs)
	fileServer := http.FileServer(httpFs)
	mux := http.NewServeMux()
	mux.Handle("/", fileServer)
	m.server = &http.Server{
		Handler:     mux,
		ReadTimeout: 3 * time.Second,
		IdleTimeout: 10 * time.Second,
	}
	return nil
}

// LoggingMiddleware is a middleware that logs the request details
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Record the start time
		start := time.Now()

		// Create a wrapper for the ResponseWriter to capture the status code
		rw := &responseWriter{w, http.StatusOK}

		// Call the next handler in the chain
		next.ServeHTTP(rw, r)

		// Log the request details
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.statusCode, time.Since(start))
	})
}

// responseWriter is a wrapper around http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

type httpURL struct {
	URL *url.URL
}

func (m httpURL) Fs() (afero.Fs, error) {
	fs, err := httpfs.New(m.URL)
	if err != nil {
		return nil, err
	}
	fs = &fsTrimPrefix{fs}
	afs := &afero.FromIOFS{fs}
	return afs, nil
}

// by default httpfs expects non-absolute path
// and http.FileServer acually adds leading / to path
// to compensate for it lets trim it
type fsTrimPrefix struct {
	fs.FS
}

func (m *fsTrimPrefix) Open(name string) (fs.File, error) {
	return m.FS.Open(strings.TrimPrefix(name, "/"))
}
