package aferowebdav

import (
	"os"
	"time"

	"github.com/spf13/afero"
	"github.com/studio-b12/gowebdav"
)

// Fs implements afero.Fs interface over webdav client
type Fs struct {
	c *gowebdav.Client
}

func NewFs(client *gowebdav.Client) afero.Fs {
	return &Fs{client}
}

func (m *Fs) Name() string {
	return "aferowebdav"
}

func (m *Fs) Stat(name string) (fi os.FileInfo, err error) {
	return m.c.Stat(name)
}

func (m *Fs) Rename(oldname, newname string) (err error) {
	return m.c.Rename(oldname, newname, true)
}

func (m *Fs) RemoveAll(name string) (err error) {
	return m.c.RemoveAll(name)
}

func (m *Fs) Remove(name string) (err error) {
	return m.c.Remove(name)
}

func (m *Fs) OpenFile(name string, _ int, _ os.FileMode) (f afero.File, err error) {
	return NewFile(m.c, name), nil
}

func (m *Fs) Open(name string) (f afero.File, err error) {
	return NewFile(m.c, name), nil
}

func (m *Fs) Mkdir(name string, mode os.FileMode) (err error) {
	return m.c.Mkdir(name, mode)
}

func (m *Fs) MkdirAll(name string, mode os.FileMode) (err error) {
	return m.c.MkdirAll(name, mode)
}

func (m *Fs) Chtimes(name string, atime, mtime time.Time) (err error) {
	return ErrNotImplemented
}

func (m *Fs) Chmod(name string, mode os.FileMode) (err error) {
	return ErrNotImplemented
}

func (m *Fs) Chown(name string, uid int, gid int) (err error) {
	return ErrNotImplemented
}

func (m *Fs) Create(name string) (f afero.File, err error) {
	return nil, ErrNotImplemented
}

func (m *Fs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	return nil, false, ErrNotImplemented
}
