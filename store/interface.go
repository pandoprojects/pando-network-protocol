package store

import (
	"github.com/pandoprojects/pando/common"
)

// Store is the interface for key/value storages.
type Store interface {
	Put(key common.Bytes, value interface{}) error
	Delete(key common.Bytes) error
	Get(key common.Bytes, value interface{}) error
}
