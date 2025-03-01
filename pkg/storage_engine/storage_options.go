package storage_engine

import "github.com/cockroachdb/pebble"

type StorageEngineOptions struct {
	Path string
}

func DefaultStorageEngineOptions(path string) (*StorageEngineOptions, error) {
	return &StorageEngineOptions{
		Path: path,
	}, nil
}

func (opts *StorageEngineOptions) GetPebbleOpts() *pebble.Options {
	pebbleOpts := &pebble.Options{}
	pebbleOpts.EnsureDefaults()
	pebbleOpts.ErrorIfNotExists = true // 默认打开数据库时，如果数据库不存在，会返回错误
	return pebbleOpts
}
