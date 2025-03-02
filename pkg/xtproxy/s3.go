package xtproxy

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	afero_s3 "github.com/fclairamb/afero-s3"
	"github.com/spf13/afero"
)

// s3URL s3://<access_key>:<secret>@endpoint/bucket
type s3URL struct {
	URL *url.URL
}

type s3Params struct {
	Endpoint   string
	Region     string
	Bucket     string
	AccessKey  string
	Secret     string
	DisableSSL bool
}

func (m s3URL) Fs() (afero.Fs, error) {
	if m.URL.Scheme != "s3" {
		return nil, ErrInvalidURL
	}
	params, err := fsSchemeS3Params(m.URL)
	if err != nil {
		return nil, err
	}
	sess, err := session.NewSession(&aws.Config{
		Endpoint: &params.Endpoint,
		Region:   &params.Region,
		Credentials: credentials.NewStaticCredentials(
			params.AccessKey,
			params.Secret,
			"",
		),
	})
	if err != nil {
		return nil, err
	}
	fs := afero_s3.NewFs(params.Bucket, sess)
	return &s3FsRootDirHack{fs}, nil
}

func fsSchemeS3Params(u *url.URL) (s3Params, error) {
	secret, _ := u.User.Password()
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) != 2 {
		return s3Params{}, fmt.Errorf("url: %s: %w", u.Path, ErrInvalidURL)
	}
	return s3Params{
		Endpoint:  u.Host,
		Region:    parts[0],
		Bucket:    parts[1],
		AccessKey: u.User.Username(),
		Secret:    secret,
	}, nil
}

// s3FsRootDirHack is nessesary to fix listing of top-level dir
// https://github.com/fclairamb/afero-s3/issues/599
// https://github.com/fclairamb/afero-s3/pull/393
type s3FsRootDirHack struct {
	afero.Fs
}

type s3FileRootDirHack struct {
	afero.File
	fs *s3FsRootDirHack
}

func (m *s3FsRootDirHack) Stat(name string) (os.FileInfo, error) {
	if path.Clean(name) == "/" {
		// taken from afero_s3.Fs.statDirectory()
		// https://github.com/fclairamb/afero-s3/commit/f84ed0c82b7245f182c3526acfd091725268ccae#diff-b4c63e1a6441951419600996e574b54527abe6f23c7b90bfa68e8ac1348ffc8eR270
		return afero_s3.NewFileInfo(path.Base(name), true, 0, time.Unix(0, 0)), nil
	}
	return m.Fs.Stat(name)
}

func (m *s3FsRootDirHack) Open(name string) (afero.File, error) {
	f, err := m.Fs.Open(name)
	return &s3FileRootDirHack{File: f, fs: m}, err
}

func (m *s3FileRootDirHack) Stat() (os.FileInfo, error) {
	return m.fs.Stat(m.File.Name())
}
