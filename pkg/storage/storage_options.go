package storage

import "github.com/cockroachdb/pebble"

type StorageEngineOptions struct {
	*pebble.Options
}

func DefaultStorageEngineOptions() *StorageEngineOptions {
	pebbleOpts := pebble.Options{}
	pebbleOpts.EnsureDefaults()
	return &StorageEngineOptions{
		Options: &pebbleOpts,
	}
}
