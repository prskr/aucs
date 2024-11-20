package cli

import (
	"time"

	"github.com/prskr/aucs/core/ports"
	"github.com/prskr/aucs/infrastructure/db"
)

type DBFlag struct {
	Path string        `name:"path" help:"Path to the database data directory" default:"${XDG_CACHE_HOME}/aucs/db"`
	TTL  time.Duration `name:"ttl" help:"Time to live for dependency look entries" default:"6h"`
}

func (f DBFlag) Open() (ports.KeyValueStore, error) {
	return db.NewBadgerKVStore(f.Path, f.TTL)
}
