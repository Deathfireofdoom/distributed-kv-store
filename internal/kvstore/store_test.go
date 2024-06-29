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
