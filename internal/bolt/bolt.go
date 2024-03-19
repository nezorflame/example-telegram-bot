package bolt

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"time"

	"go.etcd.io/bbolt"
)

// Package vars
var (
	ErrNotFound = errors.New("key not found")
	ErrNilValue = errors.New("value is nil")

	bucketName = []byte("global")
)

// DB describes local BoltDB database
type DB struct {
	b       *bbolt.DB
	log     *slog.Logger
	timeout time.Duration
}

// New creates new instance of bolt DB
func New(path string, timeout time.Duration, log *slog.Logger) (*DB, error) {
	if log == nil {
		return nil, errors.New("empty logger")
	}

	// open connection to the DB
	log.With("path", path).With("timeout", timeout).Debug("Creating DB connection")
	opts := bbolt.DefaultOptions
	if timeout > 0 {
		opts.Timeout = timeout
	}
	b, err := bbolt.Open(path, 0o755, opts)
	if err != nil {
		return nil, fmt.Errorf("unable to open DB: %w", err)
	}

	// create global bucket if it doesn't exist yet
	log.With("bucket", string(bucketName)).Debug("Setting the default bucket")
	err = b.Update(func(tx *bbolt.Tx) error {
		_, bErr := tx.CreateBucketIfNotExists(bucketName)
		return bErr
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create global bucket: %w", err)
	}

	// return the DB
	db := &DB{b: b, log: log, timeout: timeout}
	db.log.Debug("DB initiated")
	return db, nil
}

// Close closes the DB
func (db *DB) Close(delete bool) error {
	db.log.Debug("Closing the DB")
	path := db.b.Path()
	done := make(chan error)
	go func() {
		done <- db.b.Close()
		db.log.Debug("DB closed OK")
		close(done)
	}()
	timer := time.NewTimer(db.timeout)
	if delete {
		defer os.Remove(path)
	}

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("unable to close DB: %w", err)
		}
		return nil
	case <-timer.C:
		return fmt.Errorf("unable to close DB: %w", bbolt.ErrTimeout)
	}
}

// Keys returns a list of available keys in the global bucket, sorted alphabetically
func (db *DB) Keys() ([]string, error) {
	var keys []string
	db.log.Debug("Getting the list of DB current keys")
	err := db.b.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		return b.ForEach(func(k, v []byte) error {
			if v != nil {
				keys = append(keys, string(k))
			}
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get the list of keys from DB: %w", err)
	}
	sort.Strings(keys)
	return keys, nil
}

// Get acquires value from DB by provided key
func (db *DB) Get(key string) ([]byte, error) {
	var value []byte
	db.log.With("key", key).Debug("Getting value from DB")
	err := db.b.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		k, v := b.Cursor().Seek([]byte(key))
		if k == nil || string(k) != key {
			return ErrNotFound
		} else if v == nil {
			return ErrNilValue
		}
		value = make([]byte, len(v))
		copy(value, v)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get value for key '%s' from DB: %w", key, err)
	}
	db.log.With("key", key).Debug("Got the value")
	return value, nil
}

// Put sets/updates the value in DB by provided bucket and key
func (db *DB) Put(key string, val []byte) error {
	db.log.With("key", key).Debug("Saving the value to DB")
	err := db.b.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		return b.Put([]byte(key), val)
	})
	if err != nil {
		return fmt.Errorf("unable to put value for key '%s' to DB: %w", key, err)
	}
	return nil
}

// Delete removes the value from DB by provided bucket and key
func (db *DB) Delete(key string) error {
	db.log.With("key", key).Debug("Deleting from DB")
	err := db.b.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		return b.Delete([]byte(key))
	})
	if err != nil {
		return fmt.Errorf("unable to delete value for key '%s' from DB: %w", key, err)
	}
	return nil
}

// Purge removes the bucket from DB
func (db *DB) Purge() error {
	err := db.b.Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket(bucketName)
	})
	if err != nil {
		return fmt.Errorf("unable to purge global bucket from DB: %w", err)
	}
	return nil
}
