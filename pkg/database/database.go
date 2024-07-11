package database

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

const (
	BucketHistory  = "history"
	BucketMappings = "mappings"
	DbFile         = "tapto.db"
)

// Check if the db exists on disk.
func DbExists() bool {
	_, err := os.Stat(DbFile)
	return err == nil
}

// Open the db with the given options. If the database does not exist it
// will be created and the buckets will be initialized.
func open(options *bolt.Options) (*bolt.DB, error) {
	err := os.MkdirAll(filepath.Dir(DbFile), 0755)
	if err != nil {
		return nil, err
	}

	db, err := bolt.Open(DbFile, 0600, options)
	if err != nil {
		return nil, err
	}

	db.Update(func(txn *bolt.Tx) error {
		for _, bucket := range []string{BucketHistory, BucketMappings} {
			_, err := txn.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return err
			}
		}

		return nil
	})

	return db, nil
}

type Database struct {
	bdb *bolt.DB
}

func Open() (*Database, error) {
	db, err := open(&bolt.Options{})
	if err != nil {
		return nil, err
	}

	return &Database{bdb: db}, nil
}

func (d *Database) Close() error {
	return d.bdb.Close()
}

// TODO: reader source (physical reader vs web)
// TODO: metadata
type HistoryEntry struct {
	Time    time.Time `json:"time"`
	Type    string    `json:"type"`
	UID     string    `json:"uid"`
	Text    string    `json:"text"`
	Data    string    `json:"data"`
	Success bool      `json:"success"`
}

func HistoryKey(entry HistoryEntry) string {
	// TODO: web has no uid, this could collide
	return entry.Time.Format(time.RFC3339) + "-" + entry.UID
}

func (d *Database) AddHistory(entry HistoryEntry) error {
	return d.bdb.Update(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(BucketHistory))

		data, err := json.Marshal(entry)
		if err != nil {
			return err
		}

		return b.Put([]byte(HistoryKey(entry)), data)
	})
}

func (d *Database) GetHistory() ([]HistoryEntry, error) {
	var entries []HistoryEntry
	i := 0
	max := 25

	err := d.bdb.View(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(BucketHistory))

		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			if i >= max {
				break
			}

			var entry HistoryEntry
			err := json.Unmarshal(v, &entry)
			if err != nil {
				return err
			}

			entries = append(entries, entry)

			i++
		}

		return nil
	})

	return entries, err
}
