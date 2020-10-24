// Package store is an interface for distributed data storage.
// The design document is located at https://github.com/micro/development/blob/master/design/framework/store.md
package store

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrNotFound is returned when a key doesn't exist
	ErrNotFound = errors.New("not found")
	// StoreMaps
	storeMaps    = map[string]Store{}
	mux          sync.Mutex
	defaultStore Store
)

func Register(st Store, isDefault bool) {
	mux.Lock()
	defer mux.Unlock()

	storeMaps[st.String()] = st
	if isDefault {
		defaultStore = st
	}

	return
}

// Store is a data storage interface
type Store interface {
	// Init initialises the store. It must perform any required setup on the backing storage implementation and check that it is ready for use, returning any errors.
	Init(...Option) error
	// Options allows you to view the current options.
	Options() Options
	// Read takes a single key name and optional ReadOptions. It returns matching []*Record or an error.
	Read(key string, opts ...ReadOption) ([]*Record, error)
	// Write() writes a record to the store, and returns an error if the record was not written.
	Write(r *Record, opts ...WriteOption) error
	// Delete removes the record with the corresponding key from the store.
	Delete(key string, opts ...DeleteOption) error
	// List returns any keys that match, or an empty list with no error if none matched.
	List(opts ...ListOption) ([]*Record, error)
	// Close the store
	Close() error
	// String returns the name of the implementation.
	String() string

	Tx(funcs ...func(tx Tx) error) error
}

type Tx interface {
	Rollback() error
}

// Record is an item stored or retrieved from a Store
type Record struct {
	// The key to store the record
	Key string `json:"-"`
	// The value within the record
	Value []byte `json:"value"`
	// Any associated metadata for indexing
	Metadata map[string]interface{} `json:"metadata"`
	// Time to expire a record: TODO: change to timestamp
	Expiry time.Duration `json:"expiry,omitempty"`
}

func Read(key string, opts ...ReadOption) ([]*Record, error) {
	return defaultStore.Read(key, opts...)
}

func Write(r *Record, opts ...WriteOption) error {
	return defaultStore.Write(r, opts...)
}

func Delete(key string, opts ...DeleteOption) error {
	return defaultStore.Delete(key, opts...)
}

func List(opts ...ListOption) ([]*Record, error) {
	return defaultStore.List(opts...)
}
