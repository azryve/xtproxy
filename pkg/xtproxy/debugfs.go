package xtproxy

import (
	"log"
	"os"
	"time"

	"github.com/spf13/afero"
)

type DebugFs struct {
	afero.Fs
}

type DebugFile struct {
	afero.File
}

func (m *DebugFs) Chtimes(mname string, atime, mtime time.Time) (err error) {
	err = m.Fs.Chtimes(mname, atime, mtime)
	log.Printf("debugfs Chtimes(%s, %s, %s) -> %s\n", mname, atime, mtime, err)
	return err
}

func (m *DebugFs) Chmod(mname string, mode os.FileMode) (err error) {
	err = m.Fs.Chmod(mname, mode)
	log.Printf("debugfs Chmod(%s, %s) -> %s\n", mname, mode, err)
	return err
}

func (m *DebugFs) Chown(mname string, uid int, gid int) (err error) {
	err = m.Fs.Chown(mname, uid, gid)
	log.Printf("debugfs Chown(%s, %d, %d) -> %s\n", mname, uid, gid, err)
	return err
}

func (m *DebugFs) Stat(mname string) (fi os.FileInfo, err error) {
	fi, err = m.Fs.Stat(mname)
	log.Printf("debugfs Stat(%s) -> (%s, %s)\n", mname, fi, err)
	return fi, err
}

func (m *DebugFs) Rename(oldname, newname string) (err error) {
	err = m.Fs.Rename(oldname, newname)
	log.Printf("debugfs Rename(%s, %s) -> %s\n", oldname, newname, err)
	return err
}

func (m *DebugFs) RemoveAll(name string) (err error) {
	err = m.Fs.RemoveAll(name)
	log.Printf("debugfs RemoveAll(%s) -> %s\n", name, err)
	return err
}

func (m *DebugFs) Remove(name string) (err error) {
	err = m.Fs.Remove(name)
	log.Printf("debugfs Remove(%s) -> %s\n", name, err)
	return err
}

func (m *DebugFs) OpenFile(name string, flag int, mode os.FileMode) (f afero.File, err error) {
	f, err = m.Fs.OpenFile(name, flag, mode)
	log.Printf("debugfs OpenFile(%s, %d, %s) -> (%s, %s)\n", name, flag, mode, f, err)
	f = &DebugFile{f}
	return f, err
}

func (m *DebugFs) Open(name string) (f afero.File, err error) {
	f, err = m.Fs.Open(name)
	log.Printf("debugfs Open(%s) -> (%s, %s)\n", name, f, err)
	f = &DebugFile{f}
	return f, err
}

func (m *DebugFs) Mkdir(name string, mode os.FileMode) (err error) {
	err = m.Fs.Mkdir(name, mode)
	log.Printf("debugfs Mkdir(%s, %s) -> %s\n", name, mode, err)
	return err
}

func (m *DebugFs) MkdirAll(name string, mode os.FileMode) (err error) {
	err = m.Fs.Mkdir(name, mode)
	log.Printf("debugfs MkdirAll(%s, %s) -> %s\n", name, mode, err)
	return err
}

func (m *DebugFs) Create(name string) (f afero.File, err error) {
	f, err = m.Fs.Create(name)
	log.Printf("debugfs Create(%s) -> (%s, %s)\n", name, f, err)
	f = &DebugFile{f}
	return f, err
}

func (m *DebugFile) Read(p []byte) (n int, err error) {
	n, err = m.File.Read(p)
	log.Printf("debugfs Open(%s).Read(p) -> (%d, %s)\n", m.Name(), n, err)
	return n, err
}

func (m *DebugFile) ReadAt(p []byte, off int64) (n int, err error) {
	n, err = m.File.ReadAt(p, off)
	log.Printf("debugfs Open(%s).ReadAt(p, %d) -> (%d, %s)\n", m.Name(), off, n, err)
	return n, err
}

func (m *DebugFile) Write(p []byte) (n int, err error) {
	n, err = m.File.Write(p)
	log.Printf("debugfs Open(%s).Write(p) -> (%d, %s)\n", m.Name(), n, err)
	return n, err
}

func (m *DebugFile) WriteAt(p []byte, off int64) (n int, err error) {
	n, err = m.File.WriteAt(p, off)
	log.Printf("debugfs Open(%s).WriteAt(p, %d) -> (%d, %s)\n", m.Name(), off, n, err)
	return n, err
}
