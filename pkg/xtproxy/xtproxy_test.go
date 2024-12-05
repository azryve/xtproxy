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

	xhttp, addr := xtproxyHttpProxyForTest(t, fs, "")
	go xhttp.Wait()

	// ok file
	url := fmt.Sprintf("http://%s/file.txt", addr.String())
	httpc := &http.Client{}
	r, err := httpc.Get(url)
	assert.NoError(t, err)
	buf := strings.Builder{}
	io.Copy(&buf, r.Body)
	assert.Equal(t, "file contents", buf.String())

	// ok missing file
	url = fmt.Sprintf("http://%s/missing.txt", addr.String())
	r, err = httpc.Get(url)
	assert.NoError(t, err)
	assert.Equal(t, 404, r.StatusCode)
}

func TestHTTPToHTTP(t *testing.T) {
	basefs := afero.NewMemMapFs()
	afs := afero.Afero{Fs: basefs}
	assert.NoError(t, afs.MkdirAll("/a/", 0755))
	assert.NoError(t, afs.WriteFile("/a/file.txt", []byte("file contents"), 0644))

	xhttpbase, addr := xtproxyHttpProxyForTest(t, basefs, "")
	go xhttpbase.Wait()

	fs, err := FsByURL(fmt.Sprintf("http://%s", addr.String()))
	assert.NoError(t, err)

	xhttp, addr := xtproxyHttpProxyForTest(t, fs, "")
	go xhttp.Wait()

	// ok file
	url := fmt.Sprintf("http://%s/a/file.txt", addr.String())
	httpc := &http.Client{}
	r, err := httpc.Get(url)
	assert.NoError(t, err)
	buf := strings.Builder{}
	io.Copy(&buf, r.Body)
	assert.Equal(t, "file contents", buf.String())

	// missing file - will return 500 for now
	url = fmt.Sprintf("http://%s/missing.txt", addr.String())
	r, err = httpc.Get(url)
	assert.NoError(t, err)
	assert.Equal(t, 500, r.StatusCode)
}

func TestHTTPToWebdav(t *testing.T) {
	basefs := afero.NewMemMapFs()
	afs := afero.Afero{Fs: basefs}
	assert.NoError(t, afs.MkdirAll("/a/", 0755))
	assert.NoError(t, afs.WriteFile("/a/file.txt", []byte("file contents"), 0644))

	xhttpbase, addr := xtproxyHttpProxyForTest(t, basefs, "/.webdav")
	go xhttpbase.Wait()

	fsurl := fmt.Sprintf("webdav://%s/.webdav", addr.String())
	fs, err := FsByURL(fsurl)
	assert.NoError(t, err)
	xhttp, addr := xtproxyHttpProxyForTest(t, fs, "")
	go xhttp.Wait()

	// ok file
	url := fmt.Sprintf("http://%s/a/file.txt", addr.String())
	httpc := &http.Client{}
	r, err := httpc.Get(url)
	assert.NoError(t, err)
	buf := strings.Builder{}
	io.Copy(&buf, r.Body)
	assert.Equal(t, "file contents", buf.String())

	// missing file - will return 500 for now
	url = fmt.Sprintf("http://%s/missing.txt", addr.String())
	r, err = httpc.Get(url)
	assert.NoError(t, err)
	assert.Equal(t, 500, r.StatusCode)
}

func xtproxyHttpProxyForTest(t *testing.T, fs afero.Fs, webdavHandle string) (*XTProxy, net.Addr) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	assert.NoError(t, err)

	xt, err := NewXTProxy(
		WithHTTPAddr(addr),
		WithWebdavHandle(webdavHandle),
		WithMount(fs, "/"),
	)
	assert.NoError(t, err)
	return xt, xt.http.Listener.Addr()
}
