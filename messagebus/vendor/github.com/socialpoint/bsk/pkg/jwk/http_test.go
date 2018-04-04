package jwk

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"net/http/httptest"

	"github.com/stretchr/testify/assert"
)

// TODO: This test has been copied from the MAS repo.
// TODO: Refactor, the approach is not correct, we should use a round tripper and not a real HTTP server
func TestHTTPReader(t *testing.T) {
	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/ok":
			_, _ = w.Write([]byte(`{
		 "keys": [
		  {
		   "kty": "RSA",
		   "alg": "RS256",
		   "use": "sig",
		   "kid": "8132127249b217deaa73cef36809f5d4ba9a6683",
		   "n": "q4k8Zv2cg2NdBvABHNgC0XiICO2p5UmIzJg5qN0Lg1O7mLxdVJSvpsqjys814yBIVvlNFZnipbHsStM8A9Pd5bvPL2MSKO1dO3_W02BwYTcMvYXnKlHAuF4jLa2TIqIK6s2Nv3iOrIOVguQUlPcV5mQ-PBrBCtMNIQxWPLdOKFIRb3JJa-UR2i5MbYg45j4LSLH4pucyERF7-BSajsPyFVLccTPkQMdaIztN_6k3GDA1HwHk1yCIweQT3T_YEF7S1tF0SL70UKfeTSNTLTn1pAHCIV_AAoXnCPSqey_b4NdZsJ72dy2W9uhz2FV4q5T3tpbTOsCXE-6viSI-Sd0kzQ",
		   "e": "AQAB"
		  }
		 ]
		}`))
		case "/ko":
			w.WriteHeader(http.StatusServiceUnavailable)
		case "/invalid":
			_, _ = w.Write([]byte("{invalid[JSON"))
		}
	}))
	defer ts.Close()

	for _, test := range []struct {
		url    string
		expErr bool
	}{
		{ts.URL + "/ok", false},
		{ts.URL + "/ko", true},
		{ts.URL + "/invalid", true},
		{"http://localhost:-9999", true},
		{":invalid", true},
	} {
		reader := HTTPReader(http.DefaultClient, test.url)
		ks, err := reader.Read()

		if test.expErr {
			assert.Error(err)
			assert.Nil(ks)
		} else {
			assert.NoError(err)
			assert.NotNil(ks)
			assert.True(len(ks) > 0)
			assert.Equal(test.url, ks[0].Svc)
		}
	}
}

type httpSourceRoundTripper struct{}

func (g *httpSourceRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	res := &http.Response{
		Body:       ioutil.NopCloser(strings.NewReader(sampleKeySetResponse)),
		StatusCode: http.StatusOK,
		Header:     http.Header{},
	}

	return res, nil
}

var sampleKeySetResponse = `
{
 "keys": [
  {
   "kty": "RSA",
   "alg": "RS256",
   "use": "sig",
   "kid": "1560bb54b33c1d6856c3c595957ed073c534366a",
   "n": "1KwxcPeBf1iY91n12cMzGNWgm_BhZL2KNcDS46hMFaloi56uXF8X-NApNDYm-wTav8K_TPuaNAlfk1XL72-RvejA2kSsCknnGNREp23E7RFk9bswnDhkZv9AKi9mBxPYJ5NrNC4j4Ytx0CLWmOgm87zHDQIZ58Qv6vM6wUMM2FrlwXs3qi4zJg5TVUSDpaIGmL1bY5n851WCPF4yt748DWQCXhvPgxXrvTewz_vOjIJH_bp6Iwzz6GuJ7ADlvkWt4aWS0igtMxrtjoZ-GRbhUn_jAwCTvVeIWyYKUQU-oSfgcy3NEtQYkokcc9E47bKk8JLUzN5y4Ue9oLtDJhDG6Q",
   "e": "AQAB"
  },
  {
   "kty": "RSA",
   "alg": "RS256",
   "use": "sig",
   "kid": "8e80f677f599809b50ba1012222fc92fd6671a2f",
   "n": "wGm0QK0HdO-WAaoxSLklYA_kNt2mQ6S6VKPIk-XmNNy2oA7pW-aVI3Xk7lKK3pGUCzDxcVTccyhwbQSkSYVfUPUROYH9GUBZGzriqNnofrmId-T42Eva1El3EsZpldv5vyM53V5Sp-Ef9nYTDXowYArR-83rJTeKlfFyNzWSwHoG2wkmMyubFZGE8_nXWBpXDZq5MnlgaSzY7CztEst2KdocFpRbeMjeH1Yjsaur6pVEsxVBjPLcPLCKq-x7VUe3CVOhKsRHAjeZj3_mwRNQcfAvyIZZEWqO31bPwf_CYSo0dJMYyJ2TSXZQ3y0p703bZTYkrXwmsTw_YOFLKXQecQ",
   "e": "AQAB"
  },
  {
   "kty": "RSA",
   "alg": "RS256",
   "use": "sig",
   "kid": "fa98a858b6a7f96de15c1b158843241b073006d5",
   "n": "4-PXmfuWS-Lan2-WTJbmJMA6ChFul6ngzTr9qnU9rc1gozNxsyVxE6tPPeyYlQvP50If0XZPEiHFEcIlHNoA3NJhkw81em4QEMoRcn9oPur6gbmyXd0abKTBR44BQg0mXyK711JbBWjpcrpyKUs-CRp2jEMl09xQ2tTQ18SxiVRu7TAF5zZLWvrGErGJyWFAvO0YyPj9kD5YO_r7d-U4iO5-O5PjpZVNNGIxaZ-zxHeCww3vEUlbCFOjBhAnsj02naJv-4ICSKgxE1E_FzYlb4saZhx05Ek5yzdwH88JZZ9yxedVAiFOwthRSPPDPIywFpH6nUNFcYeZLyJojNGF1Q",
   "e": "AQAB"
  },
  {
   "kty": "RSA",
   "alg": "RS256",
   "use": "sig",
   "kid": "ea78209870244be0fdabeda6d821fb20d7a83bcb",
   "n": "r_klHhZTMDl4YdsTAiWmmc5tDH9GUGfF3MEhQ5XH8Y_Q8SgF-EVxgjtKhlgw_-VGpuWIC-m6bAp7Z5anbFSbvYeXIGAh18SDc1dJJKtT4TpLRFJNWe4EKJNRfNCAtIp0fb5TYDrfgchXF6iM52IB2abj10k56RrvMK5wOJa-DwWweO7Wv7CnKq3Vf_ZT_XThIGfJrkazbk5oXwtw7lcpg3bg6Y-8Q6iNAdduCNV4SW5NYdaDSRYqknIOyoAhl7Foi1IXnTm-OHNpFDLoxpWfGxZPoOfUugasEb2eC522wQOo2RtKdT0ihhmI2kVudZG5O8GGnj87qB8nuTkbwn2WNQ",
   "e": "AQAB"
  }
 ]
}
`
