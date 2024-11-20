package ports

import (
	"context"
	"errors"
)

var (
	ErrNoKVEntryForKey = errors.New("no entry found")
)

type KeyValueReader interface {
	Get(ctx context.Context, key []byte) ([]byte, error)
}

type KeyValueWriter interface {
	Put(ctx context.Context, key, value []byte) error
}

type KeyValueStore interface {
	KeyValueReader
	KeyValueWriter
	Close() error
}
