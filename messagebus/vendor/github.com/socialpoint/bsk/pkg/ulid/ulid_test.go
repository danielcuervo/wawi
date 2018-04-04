package ulid

import (
	"testing"
)

func TestTimeOrderedUuid(t *testing.T) {
	ulid := New()
	if len(ulid) != 26 {
		t.Fatalf("bad ULID size: %s. Must be 26 and is %d", ulid, len(ulid))
	}
}
