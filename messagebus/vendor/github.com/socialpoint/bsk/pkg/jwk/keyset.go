package jwk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"sort"

	"github.com/socialpoint/bsk/pkg/httpx"
)

// KeySet is a thread safe representation of a set of JWKs, as per the RFC
type KeySet struct {
	keys map[string]*Key
	mu   sync.RWMutex // guards keys
}

// NewKeySet creates a new JWK keys set containing the given keys
func NewKeySet(keys ...*Key) *KeySet {
	ks := &KeySet{
		keys: make(map[string]*Key),
	}

	for _, k := range keys {
		ks.keys[k.Kid] = k
	}

	return ks
}

func (ks *KeySet) Read() ([]*Key, error) {
	return ks.Keys(), nil
}

func (ks *KeySet) Write(keys ...*Key) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	for _, k := range keys {
		ks.keys[k.Kid] = k
	}

	return nil
}

// ReadFrom populates the key set with the aggregation of all the keys read from the given readers
func (ks *KeySet) ReadFrom(readers ...Reader) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	var keys []*Key

	for _, r := range readers {
		rks, err := r.Read()
		if err != nil {
			return err
		}

		keys = append(keys, rks...)
	}

	for _, k := range keys {
		ks.keys[k.Kid] = k
	}

	return nil
}

// Key returns the Key with the give kid
// Implements jwk.Source
func (ks *KeySet) Key(kid string) (*Key, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	if key, ok := ks.keys[kid]; ok {
		return key, nil
	}

	return nil, fmt.Errorf("key with kid %s not found", kid)
}

// Keys returns an slice with all the keys in the key set, sorted by activation date
func (ks *KeySet) Keys() []*Key {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	keys := make([]*Key, len(ks.keys))
	idx := 0
	for _, value := range ks.keys {
		keys[idx] = value
		idx++
	}

	return keys
}

// PublicKeys return a new key set containing the public parts of the keys contained in the key set
func (ks *KeySet) PublicKeys() *KeySet {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	pks := NewKeySet()

	for _, key := range ks.Keys() {
		pk, ok := key.Public().(*Key)
		if !ok {
			continue
		}
		pks.keys[pk.Kid] = pk
	}

	return pks
}

// Reset clears the KeySet
func (ks *KeySet) Reset() {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	ks.keys = make(map[string]*Key)
}

// Expire expires keys in the key set not valid at the given time
func (ks *KeySet) Expire(t time.Time) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	for kid := range ks.keys {
		if ks.keys[kid].NotAfter.Before(t) {
			delete(ks.keys, kid)
		}
	}
}

// ActiveKey returns a valid key from the key set at the given time
func (ks *KeySet) ActiveKey(t time.Time) *Key {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	for _, key := range ks.keys {
		if key.IsActive(t) {
			return key
		}
	}

	return nil
}

type jsonKeySet struct {
	Keys []*Key `json:"keys"`
}

// MarshalJSON marshals the key set into valid JSON, according to the RFC
// Implements json.Marshaler
func (ks *KeySet) MarshalJSON() ([]byte, error) {
	res := jsonKeySet{Keys: make([]*Key, len(ks.keys))}

	idx := 0
	for _, value := range ks.keys {
		res.Keys[idx] = value
		idx++
	}

	return json.Marshal(&res)
}

// UnmarshalJSON unmarshals they data into a valid key set
// Implements json.Unmarshaler
func (ks *KeySet) UnmarshalJSON(data []byte) error {
	src := jsonKeySet{}

	err := json.Unmarshal(data, &src)
	if err != nil {
		return err
	}

	ks.keys = make(map[string]*Key)

	return ks.Write(src.Keys...)
}

// ServeHTTP exposes the keyset through HTTP
// Implements http.Handler
func (ks *KeySet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httpx.Respond(w, r, http.StatusOK, ks.PublicKeys())
}

// SortKeysByNotBefore sort keys by the not before field
func SortKeysByNotBefore(keys []*Key) {
	sort.Sort(byNotBefore(keys))
}

type byNotBefore []*Key

func (k byNotBefore) Len() int {
	return len(k)
}

func (k byNotBefore) Swap(i, j int) {
	k[i], k[j] = k[j], k[i]
}

func (k byNotBefore) Less(i, j int) bool {
	return k[i].NotBefore.Before(k[j].NotBefore)
}
