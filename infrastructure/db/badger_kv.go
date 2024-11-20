package db

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/prskr/aucs/core/ports"
)

var _ ports.KeyValueStore = (*BadgerKVStore)(nil)

func NewBadgerKVStore(dbPath string, ttl time.Duration) (*BadgerKVStore, error) {
	opts := badger.DefaultOptions(dbPath)

	opts.Logger = NewBadgerSlogLogger(nil)

	badgerDB, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &BadgerKVStore{DB: badgerDB, TTL: ttl}, nil
}

type BadgerKVStore struct {
	TTL time.Duration
	DB  *badger.DB
}

// Get implements ports.KeyValueStore.
func (b *BadgerKVStore) Get(ctx context.Context, key []byte) ([]byte, error) {
	var value []byte
	err := b.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			value = bytes.Clone(val)
			return nil
		})
	})

	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, ports.ErrNoKVEntryForKey
	}

	return value, err
}

// Put implements ports.KeyValueStore.
func (b *BadgerKVStore) Put(ctx context.Context, key []byte, value []byte) error {
	return b.DB.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(key, value).WithTTL(b.TTL))
	})
}

// Close implements ports.KeyValueStore.
func (b *BadgerKVStore) Close() error {
	return b.DB.Close()
}
