package iox_test

import (
	"io"
	"testing"

	"github.com/socialpoint/bsk/pkg/awsx/iox"
	"github.com/stretchr/testify/assert"
)

func TestSpyReadCloserWithPayload_sunny(t *testing.T) {
	a := assert.New(t)
	payload := "1234"

	type args struct {
		p []byte
	}
	tests := []struct {
		name  string
		args  args
		wantN int
		wantP []byte
	}{
		{
			name:  "arg and const payload same length",
			args:  args{p: make([]byte, 4)},
			wantN: 4,
			wantP: []byte("1234"),
		},
		{
			name:  "arg longer",
			args:  args{p: make([]byte, 5)},
			wantN: 4,
			wantP: append([]byte("1234"), 0),
		},
		{
			name:  "arg shorter",
			args:  args{p: make([]byte, 3)},
			wantN: 3,
			wantP: []byte("123"),
		},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		r := iox.DummyReadCloser([]byte(payload))
		t.Run(tt.name, func(t *testing.T) {
			gotN, err := r.Read(tt.args.p)
			a.NoError(err)
			a.Equal(tt.wantN, gotN)
			a.Equal(tt.wantP, tt.args.p)
		})
	}
}

func TestSpyReadCloserWithPayload_closed(t *testing.T) {
	a := assert.New(t)
	r := iox.DummyReadCloser([]byte("1234"))
	p, err := r.Read(make([]byte, 5))
	a.NoError(err)
	a.Equal(4, p)

	a.NoError(r.Close())
	p, err = r.Read(make([]byte, 5))
	a.Error(err)
	a.Equal(0, p)
}

func TestSpyReadCloserWithPayload_EOF(t *testing.T) {
	a := assert.New(t)
	r := iox.DummyReadCloser([]byte("1234"))

	p, err := r.Read(make([]byte, 5))
	a.NoError(err)
	a.Equal(4, p)

	p, err = r.Read(make([]byte, 5))
	a.Equal(io.EOF, err)
	a.Equal(0, p)
}

func TestSpyReadCloserWithError(t *testing.T) {
	a := assert.New(t)
	r := &iox.ErrReadCloser{}
	p, err := r.Read(make([]byte, 5))
	a.Error(err)
	a.Equal(0, p)
}
