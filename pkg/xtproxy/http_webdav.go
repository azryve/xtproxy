package xtproxy

import (
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/azryve/xtproxy/pkg/aferowebdav"
	"github.com/hairyhenderson/go-fsimpl/httpfs"
	"github.com/spf13/afero"
	"github.com/studio-b12/gowebdav"
	"golang.org/x/net/webdav"
)

type XTProxyHTTPWebdav struct {
	Fs           afero.Fs
	Listener     *net.TCPListener
	WebdavHandle string
	server       *http.Server
}

func (m *XTProxyHTTPWebdav) Wait() error {
	if err := m.init(); err != nil {
		return err
	}
	return m.server.Serve(m.Listener)
}

func (m *XTProxyHTTPWebdav) init() error {
	if m.server != nil {
		return nil
	}
	mux := http.NewServeMux()

	// setup basic http file server for /
	httpFs := afero.NewHttpFs(m.Fs)
	httpHandler := http.FileServer(httpFs)
	httpHandler = ContentTypeMiddleware(httpHandler)
	httpHandler = LoggingMiddleware(httpHandler)
	mux.Handle("/", httpHandler)

	// setup webdav file server for /<webdavHandle>
	if m.WebdavHandle != "" {
		webdavHandle := m.WebdavHandle
		if !strings.HasPrefix(webdavHandle, "/") {
			webdavHandle = "/" + webdavHandle
		}
		if !strings.HasSuffix(webdavHandle, "/") {
			webdavHandle += "/"
		}
		webdavFs := aferowebdav.NewWebdavFs(m.Fs)
		var webdavHandler http.Handler = &webdav.Handler{
			Prefix:     webdavHandle,
			FileSystem: webdavFs,
			LockSystem: webdav.NewMemLS(),
		}
		webdavHandler = LoggingMiddleware(webdavHandler)
		mux.Handle(webdavHandle, webdavHandler)
	}

	m.server = &http.Server{
		Handler:     mux,
		ReadTimeout: 3 * time.Second,
		IdleTimeout: 10 * time.Second,
	}
	return nil
}

// ContentTypeMiddleware sets a content type manually
// preventing trying to seek in case its not supported by underlying fs (another http for example)
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileExt := filepath.Ext(r.URL.Path)
		mimeType := mime.TypeByExtension(fileExt)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		w.Header().Set("Content-Type", mimeType)
		next.ServeHTTP(w, r)
	})
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
	fs = &fsTrimPrefix{FS: fs}
	afs := &afero.FromIOFS{FS: fs}
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

type webdavURL struct {
	URL *url.URL
}

func (m webdavURL) Fs() (afero.Fs, error) {
	if m.URL == nil {
		return nil, ErrInvalidURL
	}
	httpUrl := *m.URL
	switch m.URL.Scheme {
	case "webdav":
		httpUrl.Scheme = "http"
	case "webdavs":
		httpUrl.Scheme = "https"
	default:
		return nil, ErrInvalidURL
	}
	user := m.URL.User.Username()
	pass, _ := m.URL.User.Password()
	client := gowebdav.NewClient(httpUrl.String(), user, pass)
	return aferowebdav.NewFs(client), nil
}
