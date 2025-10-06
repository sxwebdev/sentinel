package badgerdb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

// Get retrieves a value by key from BadgerDB
func (d *DB) Get(ctx context.Context, key []byte) ([]byte, error) {
	var value []byte
	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(nil)
		return err
	})
	if err == badger.ErrKeyNotFound {
		return nil, storecmn.ErrNotFound
	}
	return value, err
}

// Set stores a key-value pair with optional expiration
func (d *DB) Set(ctx context.Context, key []byte, value []byte, expiration time.Duration) error {
	return d.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, value)
		if expiration > 0 {
			entry = entry.WithTTL(expiration)
		}
		return txn.SetEntry(entry)
	})
}

// Delete removes a key from BadgerDB
func (d *DB) Delete(ctx context.Context, key []byte) error {
	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Keys returns all keys with the given prefix
func (d *DB) Keys(ctx context.Context, prefix []byte) ([]string, error) {
	var keys []string
	err := d.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			keys = append(keys, string(item.KeyCopy(nil)))
		}
		return nil
	})
	return keys, err
}

// KeysAndValues returns all keys and values with the given prefix
func (d *DB) KeysAndValues(ctx context.Context, prefix []byte) (map[string][]byte, error) {
	result := make(map[string][]byte)
	err := d.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := string(item.KeyCopy(nil))
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			result[key] = value
		}
		return nil
	})
	return result, err
}

// GetFromJSON retrieves a value and unmarshals it into dst
func (d *DB) GetFromJSON(ctx context.Context, key []byte, dst any) error {
	value, err := d.Get(ctx, key)
	if err != nil {
		return err
	}
	if value == nil {
		return storecmn.ErrNotFound
	}
	return json.Unmarshal(value, dst)
}

// SetJSON marshals value to JSON and stores it
func (d *DB) SetJSON(ctx context.Context, key []byte, value any, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return d.Set(ctx, key, data, expiration)
}

// Exists checks if a key exists in BadgerDB
func (d *DB) Exists(ctx context.Context, key []byte) (bool, error) {
	err := d.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		return err
	})
	if err == badger.ErrKeyNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
