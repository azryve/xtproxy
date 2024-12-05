package aferowebdav

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/studio-b12/gowebdav"
)

var _ afero.File = &File{}
var emptyMode fs.FileMode = 0

// File implements afero.File interface over webdav client
type File struct {
	c        *gowebdav.Client
	name     string
	position int64
}

func NewFile(client *gowebdav.Client, name string) *File {
	return &File{
		c:        client,
		name:     name,
		position: 0,
	}
}

func (m *File) Close() error {
	return nil
}

func (m *File) Name() string {
	return m.name
}

func (m *File) Stat() (os.FileInfo, error) {
	return m.c.Stat(m.name)
}

func (m *File) Read(p []byte) (int, error) {
	r, err := m.c.ReadStreamRange(m.name, m.position, int64(len(p)))
	if err != nil {
		return 0, err
	}
	defer r.Close()
	n, err := r.Read(p)
	m.position += int64(n)
	return n, err
}

func (m *File) ReadAt(p []byte, off int64) (int, error) {
	r, err := m.c.ReadStreamRange(m.name, int64(m.position)+off, int64(len(p)))
	if err != nil {
		return 0, err
	}
	defer r.Close()
	return r.Read(p)
}

func (m *File) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		m.position = offset
	case 1:
		m.position = int64(m.position) + offset
	case 2:
		info, err := m.c.Stat(m.name)
		if err != nil {
			return 0, err
		}
		m.position = info.Size() + offset
	}
	return int64(m.position), nil
}

func (m *File) Write(p []byte) (int, error) {
	// not implemented yet
	n, err := m.WriteAt(p, m.position)
	m.position += int64(n)
	return n, err
}

func (m *File) Readdir(count int) ([]os.FileInfo, error) {
	files, err := m.c.ReadDir(m.name)
	if err != nil {
		return nil, err
	}
	// taken from afero/mem/file.go
	if count > 0 {
		if count < len(files) {
			files = files[:count]
		}
		if len(files) == 0 {
			err = io.EOF
		}
	}
	return files, err
}

func (m *File) Readdirnames(count int) ([]string, error) {
	fi, err := m.Readdir(count)
	names := make([]string, len(fi))
	for i, f := range fi {
		_, names[i] = filepath.Split(f.Name())
	}
	return names, err
}

func (m *File) Sync() error {
	return nil
}

func (m *File) Truncate(size int64) error {
	return ErrNotImplemented
}

func (m *File) WriteString(s string) (ret int, err error) {
	return 0, ErrNotImplemented
}

func (m *File) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, ErrNotImplemented
}
