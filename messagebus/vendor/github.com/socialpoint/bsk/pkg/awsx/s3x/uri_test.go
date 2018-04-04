package s3x_test

import (
	"fmt"
	"testing"

	"github.com/socialpoint/bsk/pkg/awsx/s3x"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS3URI(t *testing.T) {
	a := assert.New(t)

	u := s3x.NewURI("bucket", "key")
	a.Equal("bucket", u.Bucket())
	a.Equal("key", u.Key())
}

func TestParseURI(t *testing.T) {
	testCases := []struct {
		name           string
		raw            string
		expectedBucket string
		expectedKey    string
	}{
		{
			name:           "final /",
			raw:            "s3://bucket/key/",
			expectedBucket: "bucket",
			expectedKey:    "key/",
		},
		{
			name:           "no final /",
			raw:            "s3://bucket/key",
			expectedBucket: "bucket",
			expectedKey:    "key",
		},
		{
			name:           "empty key",
			raw:            "s3://bucket/",
			expectedBucket: "bucket",
			expectedKey:    "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)
			t.Parallel()

			u, err := s3x.ParseURI(tc.raw)
			require.NoError(t, err)
			a.Equal(tc.expectedBucket, u.Bucket())
			a.Equal(tc.expectedKey, u.Key())

			u, err = s3x.ParseURI("s3://bucket/")
			require.NoError(t, err)
			a.Equal("bucket", u.Bucket())
			a.Equal("", u.Key())
		})
	}
}

func TestParseURI_error(t *testing.T) {
	tests := []struct {
		url string
	}{
		{""},
		{"no_slashes"},
		{"only_final_slash/"},
		{"/only_initial_slash"},
		{"/only_initial_final_slash/"},
		{"/no_scheme/key/"},
		{"/"},
		{"//"},
		{"//key/"},
		{"bad_scheme://bucket/key/"},
		{"s3://"},
		{"s3://no_key"},
	}
	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			a := assert.New(t)
			t.Parallel()

			u, err := s3x.ParseURI("s3://bucket")
			a.Error(err)
			a.Nil(u)

			u, err = s3x.ParseURI(test.url)
			a.Error(err)
			a.Nil(u, test.url)
		})
	}
}

func TestURI_String(t *testing.T) {
	var u fmt.Stringer = s3x.NewURI("bucket", "key")
	assert.Equal(t, "s3://bucket/key", u.String())
}

func TestURI_Concat(t *testing.T) {
	a := assert.New(t)

	noFinalSlash := *s3x.NewURI("bucket", "key")
	finalSlash := s3x.ConcatURI(noFinalSlash, "subkey/")
	a.Equal(s3x.NewURI("bucket", "key/subkey/"), finalSlash)

	a.Equal(s3x.NewURI("bucket", "key/subkey/subkey2"), s3x.ConcatURI(*finalSlash, "subkey2"))
}
