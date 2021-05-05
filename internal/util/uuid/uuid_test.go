package uuid

import (
	"testing"
)

func TestUUID(t *testing.T) {
	m := map[string]string{}
	for i := 0; i < 1000; i++ {
		id, err := New()
		if err != nil {
			t.Errorf("Unexpected error when creating an uuid: %s", err)
		}
		if _, ok := m[id]; ok {
			t.Errorf("Unexpected duplicate uuid: %s", id)
		}
		m[id] = id
	}
}
