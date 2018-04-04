package ulid

import (
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid"
)

var rndMutex = &sync.Mutex{}

// New creates a new ULID and hides the initialization details from the caller
func New() string {
	// unprotected access to rand.Rand is not thread safe
	rndMutex.Lock()
	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	rndMutex.Unlock()

	return ulid.MustNew(ulid.Now(), entropy).String()
}
