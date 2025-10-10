package xtproxy

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/afero"
)

type Fs interface {
	Fs() (afero.Fs, error)
}

// FsByURL generates fs from url determining it by scheme
func FsByURL(rawURL string) (afero.Fs, error) {
	var fs Fs
	URL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	switch URL.Scheme {
	case "file":
		fs = fileURL{URL}
	case "s3":
		fs = s3URL{URL}
	case "http", "https":
		fs = httpURL{URL}
	case "webdav", "webdavs":
		fs = webdavURL{URL}
	default:
		return nil, errors.New("unknown scheme")
	}
	return fs.Fs()
}

// file://<path>
type fileURL struct {
	URL *url.URL
}

func (m fileURL) Fs() (afero.Fs, error) {
	if m.URL.Scheme != "file" {
		return nil, ErrInvalidURL
	}
	path := m.URL.Host + m.URL.Path
	fs := afero.NewOsFs()
	fs = afero.NewBasePathFs(fs, path)
	return fs, nil
}

// MountPoint combines URL for remote side
// and path to which this URL will be mounted to
type MountPoint struct {
	URL  *url.URL
	Path string
}

// ParseMountPoints returns a list mount points based on cli args
// Assumes that str urls and paths do not contain whitespaces
//
// Its needed because depending on the context mount pairs are passed as single string or separately:
// xtproxy "<url1> <path1>" <url2> <path2>        -> ["<url1> <path1>", "<url2>", "<path2>"]
// XTPROXY_MOUNTS="<url1> <path1> <url2> <path2>" -> ["<url1>", "<path1>", "<url2>", "<path2>"]
func ParseMountPoints(mountArgs []string) ([]MountPoint, error) {
	mountPoints := make([]MountPoint, 0, len(mountArgs))

	// resplit everything by whitespace
	splitted := make([]string, 0, len(mountArgs))
	for _, m := range mountArgs {
		splitted = append(splitted, strings.Split(m, " ")...)
	}
	if len(splitted)%2 == 1 {
		return nil, fmt.Errorf("odd count in mount point list: '%s'", strings.Join(splitted, ", "))
	}

	for len(splitted) >= 2 {
		urlStr, mountPath := splitted[0], splitted[1]
		splitted = splitted[2:]

		URL, err := url.Parse(urlStr)
		if err != nil {
			return nil, fmt.Errorf("invalid url '%s': %w", urlStr, err)
		}
		mountPoints = append(mountPoints, MountPoint{
			URL:  URL,
			Path: mountPath,
		})
	}
	return mountPoints, nil
}
