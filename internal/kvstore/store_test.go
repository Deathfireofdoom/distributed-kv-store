package kvstore

import (
	"testing"
)

func TestStore(t *testing.T) {
	store := NewStore()

	tests := []struct {
		name      string
		operation func()
		key       string
		want      string
		wantFound bool
	}{
		{
			name: "Put and Get existing key",
			operation: func() {
				store.Put("key1", "value1")
			},
			key:       "key1",
			want:      "value1",
			wantFound: true,
		},
		{
			name:      "Get non-existing key",
			operation: func() {},
			key:       "key2",
			want:      "",
			wantFound: false,
		},
		{
			name: "Delete key",
			operation: func() {
				store.Put("key1", "value1")
				store.Delete("key1")
			},
			key:       "key1",
			want:      "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.operation()
			got, found := store.Get(tt.key)
			if found != tt.wantFound {
				t.Fatalf("expected found to be %v, got %v", tt.wantFound, found)
			}
			if got != tt.want {
				t.Fatalf("expected value %v, got %v", tt.want, got)
			}
		})
	}
}

func TestNewStore_Singleton(t *testing.T) {
	store1 := NewStore()
	store2 := NewStore()

	if store1 != store2 {
		t.Fatalf("expected store1 and store2 to be the same instance")
	}

	store1.Put("key1", "value1")
	value, ok := store2.Get("key1")
	if !ok {
		t.Fatalf("expected to find key1 in store2")
	}
	if value != "value1" {
		t.Fatalf("expected value1, got %s", value)
	}

	store2.Delete("key1")
	_, ok = store1.Get("key1")
	if ok {
		t.Fatalf("did not expect to find key1 after deletion in store2")
	}

	store1.mu.Lock()
	store1.data = make(map[string]string)
	store1.mu.Unlock()
}
