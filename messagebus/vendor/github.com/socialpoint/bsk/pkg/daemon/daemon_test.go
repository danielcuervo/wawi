package daemon_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	"github.com/socialpoint/bsk/pkg/daemon"
	"github.com/socialpoint/bsk/pkg/httpx"
	"github.com/socialpoint/bsk/pkg/netutil"
	"github.com/stretchr/testify/assert"
)

func TestDaemon(t *testing.T) {
	assert := assert.New(t)

	port := netutil.FreeTCPAddr().Port
	daemon := daemon.New(port, &HelloApp{})

	daemon.Run(context.Background())

	req, err := http.NewRequest("GET", "http://localhost:"+strconv.Itoa(port)+"/hello", nil)
	assert.NoError(err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(err)

	content, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	assert.NoError(err)

	assert.Equal("hello", string(content))
}

type HelloApp struct{}

func (a *HelloApp) RegisterHTTP(r *httpx.Router) {
	r.Route("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	}))
}

func (a *HelloApp) Run(ctx context.Context) {
}
