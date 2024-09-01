package xtproxy

import (
	"errors"
	"net/url"

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
	case "http":
		fs = httpURL{URL}
	case "https":
		fs = httpURL{URL}
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
