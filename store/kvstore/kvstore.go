package kvstore

import (
	"github.com/pandoprojects/pando/common"
	"github.com/pandoprojects/pando/rlp"
	"github.com/pandoprojects/pando/store"
	"github.com/pandoprojects/pando/store/database"
)

// NewKVStore create a new instance of KVStore.
func NewKVStore(db database.Database) store.Store {
	return &KVStore{db}
}

// KVStore a Database wrapped object.
type KVStore struct {
	db database.Database
}

// Put upserts key/value into DB
func (store *KVStore) Put(key common.Bytes, value interface{}) error {
	encodedValue, err := rlp.EncodeToBytes(value)
	if err != nil {
		return err
	}
	return store.db.Put(key, encodedValue)
}

// Delete deletes key entry from DB
func (store *KVStore) Delete(key common.Bytes) error {
	return store.db.Delete(key)
}

// Get looks up DB with key and returns result into value (passed by reference)
func (store *KVStore) Get(key common.Bytes, value interface{}) error {
	encodedValue, err := store.db.Get(key)
	if err != nil {
		return err
	}
	return rlp.DecodeBytes(encodedValue, value)
}
