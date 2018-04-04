package game

import "fmt"

// URI is the game's URI format
const URI = "%s://%s.socialpointgames.com"

var endpoints = map[string]struct {
	protocol  string
	subdomain string
}{
	"dc/amazon":  {"http", "dcaz"},
	"dc/android": {"http", "dca"},
	"dc/canvas":  {"http", "dc-canvas"},
	"dc/ios":     {"http", "dynamicdc"},
	"dl/android": {"http", "dla"},
	"dl/ios":     {"http", "dli"},
	"ds/android": {"http", "dsa"},
	"ds/ios":     {"http", "dsi"},
	"mc/android": {"http", "mca"},
	"mc/canvas":  {"http", "mc"},
	"mc/ios":     {"http", "mci"},
	"rc/android": {"https", "rca"},
	"rc/ios":     {"https", "rci"},
}

// EndpointResolverFunc is a function that returns the URL where a game
// a platform must resolve to.
type EndpointResolverFunc func(game string, platform string) (string, error)

// DefaultEndpointResolver returns the game/platform HTTP endpoint and errors
// if the game and platform are not found.
func DefaultEndpointResolver(game string, platform string) (string, error) {
	key := fmt.Sprintf("%s/%s", game, platform)
	endpoint, ok := endpoints[key]
	if !ok {
		return "", fmt.Errorf("`%s` not found in game/platform endpoints", key)
	}
	return fmt.Sprintf(URI, endpoint.protocol, endpoint.subdomain), nil
}
