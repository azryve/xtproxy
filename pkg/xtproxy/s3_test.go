package xtproxy

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestS3Params(t *testing.T) {
	s3URL := "s3://access:secret@s3-api.example.com/region-name/bucket-name"
	parsedURL, err := url.Parse(s3URL)
	assert.NoError(t, err)
	expected := s3Params{
		Endpoint:   "s3-api.example.com",
		Bucket:     "bucket-name",
		AccessKey:  "access",
		Secret:     "secret",
		Region:     "region-name",
		DisableSSL: false,
	}
	got, err := fsSchemeS3Params(parsedURL)
	assert.NoError(t, err)
	if expected != got {
		t.Fatalf("expected '%v', got '%v'", expected, got)
	}
}
