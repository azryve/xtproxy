package xtproxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestFileToHTTP(t *testing.T) {
	fs := afero.NewMemMapFs()
	afs := afero.Afero{Fs: fs}
	assert.NoError(t, afs.WriteFile("/file.txt", []byte("file contents"), 0644))

	xhttp := xtproxyHttpProxyForTest(t, fs)
	go xhttp.Wait()

	// ok file
	url := fmt.Sprintf("http://%s/file.txt", xhttp.Listener.Addr().String())
	httpc := &http.Client{}
	r, err := httpc.Get(url)
	assert.NoError(t, err)
	buf := strings.Builder{}
	io.Copy(&buf, r.Body)
	assert.Equal(t, "file contents", buf.String())

	// ok missing file
	url = fmt.Sprintf("http://%s/missing.txt", xhttp.Listener.Addr().String())
	r, err = httpc.Get(url)
	assert.NoError(t, err)
	assert.Equal(t, 404, r.StatusCode)
}

func TestHTTPToHTTP(t *testing.T) {
	basefs := afero.NewMemMapFs()
	afs := afero.Afero{Fs: basefs}
	assert.NoError(t, afs.MkdirAll("/a/", 0755))
	assert.NoError(t, afs.WriteFile("/a/file.txt", []byte("file contents"), 0644))

	xhttpbase := xtproxyHttpProxyForTest(t, basefs)
	go xhttpbase.Wait()

	fs, err := FsByURL(fmt.Sprintf("http://%s", xhttpbase.Listener.Addr().String()))
	assert.NoError(t, err)
	xhttp := xtproxyHttpProxyForTest(t, fs)
	go xhttp.Wait()

	// ok file
	url := fmt.Sprintf("http://%s/a/file.txt", xhttp.Listener.Addr().String())
	httpc := &http.Client{}
	r, err := httpc.Get(url)
	assert.NoError(t, err)
	buf := strings.Builder{}
	io.Copy(&buf, r.Body)
	assert.Equal(t, "file contents", buf.String())

	// missing file - will return 500 for now
	url = fmt.Sprintf("http://%s/missing.txt", xhttp.Listener.Addr().String())
	r, err = httpc.Get(url)
	assert.NoError(t, err)
	assert.Equal(t, 500, r.StatusCode)
}

func xtproxyHttpProxyForTest(t *testing.T, fs afero.Fs) *XTProxyHTTP {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	assert.NoError(t, err)
	lsn, err := net.ListenTCP("tcp", addr)
	assert.NoError(t, err)

	return &XTProxyHTTP{Fs: fs, Listener: lsn}
}
