package xtproxy

import (
	"fmt"
	"net/url"
	"strings"

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
	return afero_s3.NewFs(params.Bucket, sess), nil
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
