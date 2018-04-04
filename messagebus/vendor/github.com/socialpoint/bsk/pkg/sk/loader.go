package sk

import (
	"os"
	"sync"
	"time"

	"context"
)

const refreshInterval = time.Minute

// Loader loads catalogs from a source
type Loader interface {
	Load() *Catalog
}

// NewFilesystemLoader returns a loader that loads from filesystem
func NewFilesystemLoader(path string) (*FilesystemLoader, error) {
	l := &FilesystemLoader{path: path}
	err := l.refresh()

	return l, err
}

// FilesystemLoader loads catalogs from filesystem
type FilesystemLoader struct {
	path  string
	cache *Catalog
	mutex sync.RWMutex // protects cache
}

// Load loads a catalog from filesystem
func (l *FilesystemLoader) Load() *Catalog {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return l.cache
}

func (l *FilesystemLoader) refresh() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	f, err := os.Open(l.path)
	if err != nil {
		return err
	}

	l.cache, err = ParseFromHCL(f)

	return err
}

// Run converts this loader into a server.Runner
func (l *FilesystemLoader) Run(ctx context.Context) {
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := l.refresh(); err != nil {
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// NewStaticLoader creates a new catalog loader that always load the given catalog
func NewStaticLoader(c *Catalog) Loader {
	return &staticLoader{catalog: c}
}

type staticLoader struct {
	catalog *Catalog
}

// Load return a reference to the statically stored catalog
func (sl *staticLoader) Load() *Catalog {
	return sl.catalog
}
