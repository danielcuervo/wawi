package jwk

import (
	"net/http"
	"time"

	"github.com/socialpoint/bsk/pkg/httpc"
)

// GoogleCertsReader returns a Reader that reads keys from the Google OAuth2 public API
// It's configured with appropriate default values to ease construction
func GoogleCertsReader() Reader {
	client := httpc.Decorate(&http.Client{Timeout: 30 * time.Second}, httpc.FaultTolerance(5, time.Second))
	return HTTPReader(client, "https://www.googleapis.com/oauth2/v3/certs")
}
