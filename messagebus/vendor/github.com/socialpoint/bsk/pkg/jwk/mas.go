package jwk

import (
	"net/http"
	"time"

	"github.com/socialpoint/bsk/pkg/httpc"
)

// MasterAuthorityKeysReader returns a Reader that reads keys from the Master Authority public API
// It's configured with appropriate default values to ease construction
func MasterAuthorityKeysReader() Reader {
	client := httpc.Decorate(&http.Client{Timeout: 30 * time.Second}, httpc.FaultTolerance(5, time.Second))
	return HTTPReader(client, "http://mas.bs.laicosp.net/certs")
}
