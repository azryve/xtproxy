package aferomount

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestMountfsRootMount(t *testing.T) {
	var afs afero.Afero

	memfs1 := afero.NewMemMapFs()
	afs = afero.Afero{memfs1}
	assert.NoError(t, afs.MkdirAll("/a/b", 0755))
	assert.NoError(t, afs.WriteFile("/a/file.txt", []byte("/a/file.txt: memfs1"), 0644))

	memfs2 := afero.NewMemMapFs()
	afs = afero.Afero{memfs2}
	assert.NoError(t, afs.MkdirAll("/a/b", 0755))
	assert.NoError(t, afs.WriteFile("/a/file.txt", []byte("/a/file.txt: memfs2"), 0644))

	var err error
	var expected string
	mountfs := NewMountFS(afero.NewReadOnlyFs(afero.NewMemMapFs()))
	afs = afero.Afero{mountfs}
	_, err = afs.ReadFile("/a/file.txt")
	assert.Error(t, err, "/a/file.txt should not exist")

	assert.NoError(t, mountfs.Mount(memfs1, "/"))
	expected = "/a/file.txt: memfs1"
	got, err := afs.ReadFile("/a/file.txt")
	assert.NoError(t, err)
	assert.Equal(t, expected, string(got))

	assert.NoError(t, mountfs.Mount(memfs2, "/"))
	expected = "/a/file.txt: memfs2"
	got, err = afs.ReadFile("/a/file.txt")
	assert.NoError(t, err)
	assert.Equal(t, expected, string(got))

	assert.NoError(t, mountfs.Umount("/"))
	expected = "/a/file.txt: memfs1"
	got, err = afs.ReadFile("/a/file.txt")
	assert.NoError(t, err)
	assert.Equal(t, expected, string(got))

	assert.NoError(t, mountfs.Umount("/"))
	_, err = afs.ReadFile("/a/file.txt")
	assert.Error(t, err)
	assert.Error(t, mountfs.Umount("/"), "/ should not be mounted")
}

func TestMountfsOverlappingMounts(t *testing.T) {
	var afs afero.Afero

	memfs1 := afero.NewMemMapFs()
	afs = afero.Afero{memfs1}
	assert.NoError(t, afs.WriteFile("/a/file.txt", []byte("/a/file.txt: memfs1"), 0644))

	memfs2 := afero.NewMemMapFs()
	afs = afero.Afero{memfs2}
	assert.NoError(t, afs.WriteFile("/a/file.txt", []byte("/a/file.txt: memfs2"), 0644))

	mountfs := NewMountFS(afero.NewMemMapFs())
	afs = afero.Afero{mountfs}

	var err error
	var got []byte
	var expected string
	assert.NoError(t, mountfs.Mount(memfs1, "/"))
	assert.NoError(t, mountfs.Mount(memfs2, "/a"))
	expected = "/a/file.txt: memfs2"
	got, err = afs.ReadFile("/a/a/file.txt")
	assert.NoError(t, err)
	assert.Equal(t, string(got), expected)
}

func TestMountDirCreated(t *testing.T) {
	var afs afero.Afero

	memfs1 := afero.NewMemMapFs()
	afs = afero.Afero{memfs1}
	assert.NoError(t, afs.WriteFile("/file.txt", []byte("/file.txt: memfs1"), 0644))
	mountfs := NewMountFS(afero.NewMemMapFs())
	afs = afero.Afero{mountfs}

	exists, err := afs.DirExists("/a")
	assert.NoError(t, err)
	assert.True(t, !exists, "expected /a dir should not exist")

	assert.NoError(t, mountfs.Mount(memfs1, "/a"))
	exists, err = afs.DirExists("/a")
	assert.NoError(t, err)
	assert.True(t, exists, "expected /a dir to exist")

	stat, err := afs.Stat("/a")
	assert.NoError(t, err)
	assert.True(t, stat.IsDir(), "expected /a should be a dir")

	mountfs.Umount("/a")
	exists, err = afs.DirExists("/a")
	assert.NoError(t, err)
	assert.True(t, !exists, "expected /a dir should be removed after unmount")
}

func TestMountDirExisted(t *testing.T) {
	var afs afero.Afero

	memfs1 := afero.NewMemMapFs()
	afs = afero.Afero{memfs1}
	assert.NoError(t, afs.WriteFile("/file.txt", []byte("/file.txt: memfs1"), 0644))
	mountfs := NewMountFS(afero.NewMemMapFs())
	afs = afero.Afero{mountfs}
	afs.MkdirAll("/a", 0777)

	exists, err := afs.DirExists("/a")
	assert.NoError(t, err)
	assert.True(t, exists, "expected /a dir should exist")

	assert.NoError(t, mountfs.Mount(memfs1, "/a"))
	exists, err = afs.DirExists("/a")
	assert.NoError(t, err)
	assert.True(t, exists, "expected /a dir continue to exist")

	stat, err := afs.Stat("/a")
	assert.NoError(t, err)
	assert.True(t, stat.IsDir(), "expected /a should be a dir")

	mountfs.Umount("/a")
	exists, err = afs.DirExists("/a")
	assert.NoError(t, err)
	assert.True(t, exists, "expected /a dir should continue to exist")
}

func TestMountDirCreatedOverlap(t *testing.T) {
	memfs1 := afero.NewMemMapFs()
	memfs2 := afero.NewMemMapFs()

	mountfs := NewMountFS(afero.NewMemMapFs())
	afs := afero.Afero{mountfs}

	assert.NoError(t, mountfs.Mount(memfs1, "/"))
	assert.NoError(t, mountfs.Mount(memfs2, "/a"))

	stat, err := afs.Stat("/a")
	assert.NoError(t, err)
	assert.True(t, stat.IsDir(), "expected /a should be a dir")
}

func TestInsertSorted(t *testing.T) {
	var slice []string
	var expected string
	slice = make([]string, 0)
	for _, s := range []string{"bb", "ccc", "a"} {
		slice, _ = insertSorted(slice, s, byLen)
	}
	expected = "a,bb,ccc"
	if got := strings.Join(slice, ","); got != expected {
		t.Fatalf("expected '%s', got '%s'", expected, got)
	}

	slice = make([]string, 0)
	for _, s := range []string{"bb", "ccc", "a"} {
		slice, _ = insertSorted(slice, s, byReverseLen)
	}
	expected = "ccc,bb,a"
	if got := strings.Join(slice, ","); got != expected {
		t.Fatalf("expected '%s', got '%s'", expected, got)
	}
}
