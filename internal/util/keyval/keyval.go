package keyval

import (
	"errors"
	"fmt"
	"time"

	"github.com/joyrex2001/kubedock/internal/util/uuid"
	cache "github.com/patrickmn/go-cache"
)

// ErrNotFound is raised when retrieval of the given key fails.
var ErrNotFound = errors.New("not found")

// Database defines the public interface.
type Database interface {
	Create(interface{}) (string, error)
	Read(string) (interface{}, error)
	Update(string, interface{}) error
	Delete(string) error
}

// instance is the internal object representation for Database.
type instance struct {
	db *cache.Cache
}

// New will return a Database object.
func New() (Database, error) {
	return &instance{
		db: cache.New(cache.NoExpiration, 10*time.Minute),
	}, nil
}

// Create will store given value at given and return an unique
// key representing the resource.
func (kv *instance) Create(val interface{}) (string, error) {
	key, err := uuid.New()
	if err != nil {
		return "", err
	}
	kv.db.Set(key, val, cache.NoExpiration)
	return key, nil
}

// Read will return the value for given key. If the key does not
// exist, it will return an error.
func (kv *instance) Read(key string) (interface{}, error) {
	x, ok := kv.db.Get(key)
	if !ok {
		return nil, fmt.Errorf("%s: %w", key, ErrNotFound)
	}
	return x, nil
}

// Update will store given value at given key. If the key does not
// exist, it will return an error.
func (kv *instance) Update(key string, val interface{}) error {
	if err := kv.db.Replace(key, val, cache.NoExpiration); err != nil {
		return fmt.Errorf("%s: %w", key, ErrNotFound)
	}
	return nil
}

// Delete will delete the value for given key. If the key does
// not exist, it will return an error.
func (kv *instance) Delete(key string) error {
	if _, err := kv.Read(key); err != nil {
		return err
	}
	kv.db.Delete(key)
	return nil
}
