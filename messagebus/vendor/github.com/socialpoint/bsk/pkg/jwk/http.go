package jwk

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/socialpoint/bsk/pkg/httpc"
)

// HTTPReader returns a reader that read keys from a HTTP endpoint using the provided client
func HTTPReader(client httpc.Client, url string) Reader {
	return ReaderFunc(func() ([]*Key, error) {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		res, err := client.Do(req)
		if err != nil {
			return nil, errors.New("unreachable endpoint")
		}

		ks := NewKeySet()

		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return nil, errors.New("unexpected endpoint error")
		}
		err = json.NewDecoder(res.Body).Decode(ks)
		if err != nil {
			return nil, err
		}

		for _, k := range ks.Keys() {
			k.Svc = url
		}

		return ks.Keys(), nil
	})
}
