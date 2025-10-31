package xtproxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMountPoints(t *testing.T) {
	mountPoints, err := ParseMountPoints([]string{
		"file:/// /tmp01",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, mountPoints)
	assert.Equal(t, len(mountPoints), 1)
	assert.Equal(t, mountPoints[0].URL.String(), "file:///")
	assert.Equal(t, mountPoints[0].Path, "/tmp01")

	mountPoints, err = ParseMountPoints([]string{
		"file:///", "/tmp01",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, mountPoints)
	assert.Equal(t, len(mountPoints), 1)
	assert.Equal(t, mountPoints[0].URL.String(), "file:///")
	assert.Equal(t, mountPoints[0].Path, "/tmp01")

	mountPoints, err = ParseMountPoints([]string{
		"file:///aaa/ /tmp01",
		"file:///bbb/", "/tmp02",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, mountPoints)
	assert.Equal(t, len(mountPoints), 2)
	assert.Equal(t, mountPoints[0].URL.String(), "file:///aaa/")
	assert.Equal(t, mountPoints[0].Path, "/tmp01")
	assert.Equal(t, mountPoints[1].URL.String(), "file:///bbb/")
	assert.Equal(t, mountPoints[1].Path, "/tmp02")

	mountPoints, err = ParseMountPoints([]string{
		"file:///aaa/",
	})
	assert.ErrorContains(t, err, "odd count in mount point list")

	mountPoints, err = ParseMountPoints([]string{
		":: /tmp/b",
	})
	assert.ErrorContains(t, err, "invalid url", mountPoints)
}
