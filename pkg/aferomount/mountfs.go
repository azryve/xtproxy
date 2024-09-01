package aferomount

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// MountFs is a convenience afero.Fs to map prefix to another prefixs
type MountFs struct {
	base       afero.Fs              // base fs if no mountpoint found
	mounts     map[string][]afero.Fs // map from path to mount
	paths      []string              // reverse sorted by lenght
	pathIndex  map[string]int        // reverse indexes of paths
	dirCreated map[string]struct{}   // marks that we created a dir at mount
}

// MountFile represents a file inside a mountfs with original name intact
type MountFile struct {
	afero.File
	name string
}

// NewStripPrefixFs is an internal FS implementation to remove a path prefix
// of a path when calling the underlying FS implementation.
func NewMountFS(base afero.Fs) *MountFs {
	return &MountFs{
		base:       base,
		mounts:     make(map[string][]afero.Fs),
		paths:      make([]string, 0),
		pathIndex:  make(map[string]int),
		dirCreated: make(map[string]struct{}),
	}
}

func (m *MountFs) Name() string {
	return "MountFs"
}

func (m *MountFs) Mount(mountfs afero.Fs, path string) error {
	apath := absPath(path)
	afs := afero.Afero{m.base}
	dirExist, err := afs.DirExists(apath)
	if err != nil {
		return err
	}
	if !dirExist {
		if err := afs.MkdirAll(apath, 0777); err != nil {
			return err
		}
		m.dirCreated[apath] = struct{}{}
	}
	mounts, ok := m.mounts[apath]
	if !ok {
		mounts = make([]afero.Fs, 0, 1)
		paths, idx := insertSorted(m.paths, apath, byReverseLen)
		m.paths = paths
		m.pathIndex[apath] = idx
	}
	mounts = append(mounts, mountfs)
	m.mounts[apath] = mounts
	return nil
}

func (m *MountFs) Umount(path string) error {
	apath := absPath(path)
	mounts, ok := m.mounts[apath]
	if !ok {
		return errors.New("not mounted")
	}
	if len(mounts) > 0 {
		mounts = mounts[:len(mounts)-1]
		m.mounts[apath] = mounts
	}
	if len(mounts) > 0 {
		return nil
	}
	delete(m.mounts, apath)
	idx := m.pathIndex[apath]
	delete(m.pathIndex, apath)
	copy(m.paths[idx:], m.paths[idx+1:])
	m.paths = m.paths[:len(m.paths)-1]
	if _, ok := m.dirCreated[apath]; ok {
		delete(m.dirCreated, apath)
		if err := m.base.Remove(apath); err != nil {
			return err
		}
	}
	return nil
}

func (m *MountFs) findMount(name string) (string, afero.Fs) {
	aname := absPath(name)
	for _, mpath := range m.paths {
		mounts := m.mounts[mpath]
		if len(mounts) > 0 && strings.HasPrefix(aname, mpath) {
			mname := absPath(aname[len(mpath):])
			mount := mounts[len(mounts)-1]
			return mname, mount
		}
	}
	return name, m.base
}

func (m *MountFs) Chtimes(mname string, atime, mtime time.Time) (err error) {
	mname, mount := m.findMount(mname)
	return mount.Chtimes(mname, atime, mtime)
}

func (m *MountFs) Chmod(mname string, mode os.FileMode) (err error) {
	mname, mount := m.findMount(mname)
	return mount.Chmod(mname, mode)
}

func (m *MountFs) Chown(mname string, uid int, gid int) (err error) {
	mname, mount := m.findMount(mname)
	return mount.Chown(mname, uid, gid)
}

func (m *MountFs) Stat(mname string) (fi os.FileInfo, err error) {
	mname, mount := m.findMount(mname)
	return mount.Stat(mname)
}

func (m *MountFs) Rename(oldname, newname string) (err error) {
	mname, mount := m.findMount(oldname)
	newname, newmount := m.findMount(newname)
	if mount != newmount {
		return &os.PathError{Op: "rename", Path: oldname, Err: errors.New("EXDEV")}
	}
	return mount.Rename(mname, newname)
}

func (m *MountFs) RemoveAll(name string) (err error) {
	name, mount := m.findMount(name)
	return mount.RemoveAll(name)
}

func (m *MountFs) Remove(name string) (err error) {
	name, mount := m.findMount(name)
	return mount.Remove(name)
}

func (m *MountFs) OpenFile(name string, flag int, mode os.FileMode) (f afero.File, err error) {
	mname, mount := m.findMount(name)
	fh, err := mount.OpenFile(mname, flag, mode)
	if err != nil {
		return nil, err
	}
	return &MountFile{File: fh, name: name}, nil
}

func (m *MountFs) Open(name string) (f afero.File, err error) {
	mname, mount := m.findMount(name)
	fh, err := mount.Open(mname)
	if err != nil {
		return nil, err
	}
	return &MountFile{File: fh, name: name}, nil
}

func (m *MountFs) Mkdir(name string, mode os.FileMode) (err error) {
	mname, mount := m.findMount(name)
	return mount.Mkdir(mname, mode)
}

func (m *MountFs) MkdirAll(name string, mode os.FileMode) (err error) {
	mname, mount := m.findMount(name)
	return mount.MkdirAll(mname, mode)
}

func (m *MountFs) Create(name string) (f afero.File, err error) {
	mname, mount := m.findMount(name)
	fh, err := mount.Create(mname)
	if err != nil {
		return nil, err
	}
	return &MountFile{File: fh, name: name}, nil
}

func (m *MountFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	mname, mount := m.findMount(name)
	if lstater, ok := mount.(afero.Lstater); ok {
		return lstater.LstatIfPossible(mname)
	}
	fi, err := m.base.Stat(name)
	return fi, false, err
}

func (m *MountFile) Name() string {
	return m.name
}

func absPath(name string) string {
	if !filepath.IsAbs(name) {
		return filepath.Join(string(filepath.Separator), name)
	}
	return name
}

type keyFunc func(string) int

var byLen keyFunc = func(x string) int { return len(x) }
var byReverseLen keyFunc = func(x string) int { return -len(x) }

func insertSorted(slice []string, s string, key func(string) int) ([]string, int) {
	index := sort.Search(len(slice), func(i int) bool {
		return key(s) <= key(slice[i])
	})
	slice = append(slice, "")
	copy(slice[index+1:], slice[index:])
	slice[index] = s
	return slice, index
}
