package jwk

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

// NewFile returns a ReadWriter implementation that reads and writes keys from a file
func NewFile(path string) ReadWriter {
	return &file{path: path}
}

type file struct {
	path string
}

func (fs *file) Read() ([]*Key, error) {
	f, err := os.Open(fs.path)
	if err != nil {
		return nil, err
	}

	ks := NewKeySet()
	err = json.NewDecoder(f).Decode(&ks)

	keys := ks.Keys()
	sort.Sort(sort.Reverse(byNotBefore(keys)))

	return keys, err
}

func (fs *file) Write(keys ...*Key) error {
	f, err := os.OpenFile(fs.path, os.O_RDONLY|os.O_CREATE, 0666) // Read-only
	if err != nil {
		return err
	}
	defer f.Close()

	ks := NewKeySet()

	// Ignoring decoding error is intentional
	// If the file is empty or corrupt we will just create a new valid one
	// from an empty key set.
	_ = json.NewDecoder(f).Decode(&ks)

	err = ks.Write(keys...)
	if err != nil {
		return err
	}

	return atomicFilesystemWrite(fs.path, ks)
}

// Write keys to filesystem in an atomic way
//
// Basically a temp file is created, keys are dumped to it and then,
// if everything has gone right, the temp file overrides the original one.
//
// This strategy has been inspired an adapted by:
// - https://github.com/youtube/vitess/blob/master/go/ioutil2/ioutil.go#L15
// - https://github.com/dchest/safefile
func atomicFilesystemWrite(path string, ks *KeySet) error {
	tmp, err := ioutil.TempFile(filepath.Dir(path), "tmp")
	if err != nil {
		return err
	}
	defer tmp.Close()

	err = json.NewEncoder(tmp).Encode(&ks)
	if err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}

	err = tmp.Sync()
	if err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}

	return os.Rename(tmp.Name(), path)
}
