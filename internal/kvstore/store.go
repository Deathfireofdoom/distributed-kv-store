package kvstore

import (
	"sync"

	"github.com/Deathfireofdoom/distributed-kv-store/internal/models"
)

type Store struct {
	mu   sync.Mutex
	data map[string]string
}

var instance *Store
var once sync.Once

func NewStore() *Store {
	once.Do(func() {
		instance = &Store{
			data: make(map[string]string),
		}
	})
	return instance
}

func (s *Store) Put(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.data[key]
	return value, ok
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *Store) Apply(command models.Command) {
	switch command.Type {
	case models.PutCommand:
		s.Put(command.Key, command.Value)
	case models.DeleteCommand:
		s.Delete(command.Key)
	}
}
