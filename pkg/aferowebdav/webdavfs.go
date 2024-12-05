package aferowebdav

import (
	"context"
	"os"

	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

var _ webdav.FileSystem = &WebdavFs{}

// WebdavFs is adapter implementing webdav.FileSystem over afero.Fs
type WebdavFs struct {
	fs afero.Fs
}

func NewWebdavFs(fs afero.Fs) *WebdavFs {
	return &WebdavFs{fs: fs}
}

func (m *WebdavFs) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	var f webdav.File
	var err error
	done := make(chan struct{})
	go func() {
		f, err = m.fs.OpenFile(name, flag, perm)
		close(done)
	}()
	select {
	case <-done:
		return f, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (m *WebdavFs) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	var fi os.FileInfo
	var err error
	done := make(chan struct{})
	go func() {
		fi, err = m.fs.Stat(name)
		close(done)
	}()
	select {
	case <-done:
		return fi, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (m *WebdavFs) RemoveAll(ctx context.Context, name string) error {
	return ErrNotImplemented
}

func (m *WebdavFs) Rename(ctx context.Context, oldName, newName string) error {
	return ErrNotImplemented
}
func (m *WebdavFs) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return ErrNotImplemented
}
