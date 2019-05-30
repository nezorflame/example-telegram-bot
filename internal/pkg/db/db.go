package db

import (
	"os"
	"sort"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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
	timeout time.Duration
}

// New creates new instance of bolt DB
func New(path string, timeout time.Duration) (*DB, error) {
	// open connection to the DB
	log.WithField("path", path).WithField("timeout", timeout).Debug("Creating DB connection")
	opts := bbolt.DefaultOptions
	if timeout > 0 {
		opts.Timeout = timeout
	}
	b, err := bbolt.Open(path, 0755, opts)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open DB")
	}

	// create global bucket if it doesn't exist yet
	log.WithField("bucket", string(bucketName)).Debug("Setting the default bucket")
	err = b.Update(func(tx *bbolt.Tx) error {
		_, bErr := tx.CreateBucketIfNotExists(bucketName)
		return bErr
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create global bucket")
	}

	// return the DB
	db := &DB{b: b, timeout: timeout}
	log.Debug("DB initiated")
	return db, nil
}

// Close closes the DB
func (db *DB) Close(delete bool) error {
	log.Debug("Closing the DB")
	path := db.b.Path()
	done := make(chan error)
	go func() {
		done <- db.b.Close()
		log.Debug("DB closed OK")
		close(done)
	}()
	timer := time.NewTimer(db.timeout)
	if delete {
		defer os.Remove(path)
	}

	select {
	case err := <-done:
		if err != nil {
			return errors.Wrap(err, "unable to close DB")
		}
		return nil
	case <-timer.C:
		return errors.Wrap(bbolt.ErrTimeout, "unable to close DB")
	}
}

// Keys returns a list of available keys in the global bucket, sorted alphabetically
func (db *DB) Keys() ([]string, error) {
	var keys []string
	log.Debug("Getting the list of DB current keys")
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
		return nil, errors.Wrap(err, "unable to get the list of keys from DB")
	}
	sort.Strings(keys)
	return keys, nil
}

// Get acquires value from DB by provided key
func (db *DB) Get(key string) ([]byte, error) {
	var value []byte
	log.WithField("key", key).Debug("Getting value from DB")
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
		return nil, errors.Wrap(err, "unable to get value for key '%s' from DB")
	}
	log.WithField("key", key).Debug("Got the value")
	return value, nil
}

// Put sets/updates the value in DB by provided bucket and key
func (db *DB) Put(key string, val []byte) error {
	log.WithField("key", key).Debug("Saving the value to DB")
	err := db.b.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		return b.Put([]byte(key), val)
	})
	if err != nil {
		return errors.Wrapf(err, "unable to put value for key '%s' to DB", key)
	}
	return nil
}

// Delete removes the value from DB by provided bucket and key
func (db *DB) Delete(key string) error {
	log.WithField("key", key).Debug("Deleting from DB")
	err := db.b.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		return b.Delete([]byte(key))
	})
	if err != nil {
		return errors.Wrapf(err, "unable to delete value for key '%s' from DB", key)
	}
	return nil
}

// Purge removes the bucket from DB
func (db *DB) Purge() error {
	err := db.b.Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket(bucketName)
	})
	if err != nil {
		return errors.Wrap(err, "unable to purge global bucket from DB")
	}
	return nil
}
