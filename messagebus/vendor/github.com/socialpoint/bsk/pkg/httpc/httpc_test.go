package httpc_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"time"

	"bytes"

	"fmt"

	"sync"

	"github.com/socialpoint-labs/bsk/metrics"
	"github.com/socialpoint/bsk/pkg/httpc"
	"github.com/stretchr/testify/assert"
)

const errorMsg = "randomError"

type NoopClient struct{}

func (c *NoopClient) Do(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
	}, nil
}

type FailingClient struct {
	attempts int
}

func (c *FailingClient) Do(*http.Request) (*http.Response, error) {
	defer func() {
		c.attempts++
	}()

	// This makes the client fails in the first 2 attempts
	if c.attempts < 2 {
		return nil, errors.New(errorMsg)
	}

	return &http.Response{
		StatusCode: http.StatusOK,
	}, nil
}

func TestHeaderDecorator(t *testing.T) {
	assert := assert.New(t)

	client := httpc.Decorate(&NoopClient{}, httpc.Header("test", "123"))

	req, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(err)

	resp, err := client.Do(req)
	assert.NoError(err)
	assert.Equal(200, resp.StatusCode)

	assert.Equal("123", req.Header.Get("test"))
}

func TestFaultTolerance(t *testing.T) {
	assert := assert.New(t)

	client := httpc.Decorate(&FailingClient{}, httpc.FaultTolerance(5, time.Millisecond))

	req, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(err)

	resp, err := client.Do(req)
	assert.NoError(err)

	assert.Equal(200, resp.StatusCode)
}

func TestLogger(t *testing.T) {
	assert := assert.New(t)

	recorder := &bytes.Buffer{}

	client := httpc.Decorate(&NoopClient{}, httpc.Logger(recorder))

	req, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(err)

	resp, err := client.Do(req)
	assert.NoError(err)
	assert.Equal(200, resp.StatusCode)

	assert.Equal("GET http://example.com\n", recorder.String())
}

func TestLoggerf(t *testing.T) {
	assert := assert.New(t)

	recorder := &bytes.Buffer{}

	formatter := func(r *http.Request) string { return fmt.Sprintf("[%s][%s]", r.Method, r.URL.String()) }
	client := httpc.Decorate(&NoopClient{}, httpc.Loggerf(recorder, formatter))

	req, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(err)

	resp, err := client.Do(req)
	assert.NoError(err)
	assert.Equal(200, resp.StatusCode)

	assert.Equal("[GET][http://example.com]", recorder.String())
}

func TestFake(t *testing.T) {
	assert := assert.New(t)

	for _, fakes := range [][]httpc.FakeResponse{
		{httpc.NewFake("foo", http.StatusOK)},
		{httpc.NewFake("teapot", http.StatusTeapot)},
		// multiple/successive
		{httpc.NewFake("foo", http.StatusOK), httpc.NewFake("teapot", http.StatusTeapot)},
	} {
		client := httpc.Decorate(http.DefaultClient, httpc.Fake(fakes...))
		assert.NotNil(client)

		for _, fake := range fakes {
			resp, err := client.Do(&http.Request{})
			assert.NoError(err)
			assert.NotNil(resp)

			body, err := ioutil.ReadAll(resp.Body)
			assert.NoError(err)
			assert.Equal(fake.Content, string(body))
			assert.Equal(fake.StatusCode, resp.StatusCode)
		}
	}
}

func TestConcurrentFake(t *testing.T) {
	assert := assert.New(t)

	r := httpc.NewFake("teapot", http.StatusTeapot)
	client := httpc.Decorate(http.DefaultClient, httpc.Fake(r, r))
	assert.NotNil(client)

	wg := &sync.WaitGroup{}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			resp, err := client.Do(&http.Request{})
			assert.NoError(err)
			assert.NotNil(resp)

			body, err := ioutil.ReadAll(resp.Body)
			assert.NoError(err)
			assert.Equal("teapot", string(body))
			assert.Equal(http.StatusTeapot, resp.StatusCode)
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestQueryDecorator(t *testing.T) {
	assert := assert.New(t)

	client := httpc.Decorate(&NoopClient{}, httpc.Query("test", "123"))

	req, err := http.NewRequest("GET", "http://example.com", nil)
	assert.NoError(err)

	resp, err := client.Do(req)
	assert.NoError(err)
	assert.Equal(200, resp.StatusCode)

	assert.Equal("123", req.URL.Query().Get("test"))
}

func TestInstrument(t *testing.T) {
	assert := assert.New(t)

	recorder := metrics.NewRecorder()

	req, err := http.NewRequest("GET", "http://example.com/test", nil)
	assert.NoError(err)

	for _, test := range []struct {
		client             httpc.Client
		expectedStatusCode int
	}{
		{&NoopClient{}, http.StatusOK},
		{&FailingClient{}, http.StatusInternalServerError},
	} {
		client := httpc.Decorate(test.client, httpc.Instrument(recorder, metrics.NewTag("foo", "bar")))
		resp, err := client.Do(req)

		timer := recorder.Get("httpc.request_duration")
		tm, ok := timer.(*metrics.RecorderTimer)
		assert.True(ok)
		assert.WithinDuration(tm.StoppedTime, tm.StartedTime, time.Millisecond)
		assert.True(metrics.HasTag(timer, "method", "get"))
		assert.True(metrics.HasTag(timer, "foo", "bar"))

		if _, failCase := test.client.(*FailingClient); failCase {
			assert.Nil(resp)
			assert.Error(err)
			continue
		}

		assert.Equal(test.expectedStatusCode, resp.StatusCode)
		assert.NoError(err)
		assert.True(metrics.HasTag(timer, "code", test.expectedStatusCode))
	}
}
