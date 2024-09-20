package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	bolt "go.etcd.io/bbolt"
)

const (
	BucketHistory  = "history"
	BucketMappings = "mappings"
	BucketClients  = "clients"
)

func dbFile(pl platforms.Platform) string {
	return filepath.Join(pl.ConfigFolder(), config.TapToDbFilename)
}

// Check if the db exists on disk.
func DbExists(pl platforms.Platform) bool {
	_, err := os.Stat(dbFile(pl))
	return err == nil
}

// Open the db with the given options. If the database does not exist it
// will be created and the buckets will be initialized.
func open(pl platforms.Platform, options *bolt.Options) (*bolt.DB, error) {
	err := os.MkdirAll(filepath.Dir(dbFile(pl)), 0755)
	if err != nil {
		return nil, err
	}

	db, err := bolt.Open(dbFile(pl), 0600, options)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(txn *bolt.Tx) error {
		for _, bucket := range []string{
			BucketHistory,
			BucketMappings,
			BucketClients,
		} {
			_, err := txn.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

type Database struct {
	bdb *bolt.DB
}

func Open(pl platforms.Platform) (*Database, error) {
	db, err := open(pl, &bolt.Options{})
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
	// TODO: web has no uid, this could collide, use autoincrement instead
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
	maxResults := 25

	err := d.bdb.View(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(BucketHistory))

		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			if i >= maxResults {
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

type Client struct {
	Id      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Address string    `json:"address"`
	Secret  string    `json:"secret"`
}

func clientKey(id uuid.UUID) []byte {
	return []byte(id.String())
}

func (d *Database) GetClient(id uuid.UUID) (Client, error) {
	var c Client

	err := d.bdb.View(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(BucketClients))

		v := b.Get(clientKey(id))
		if v == nil {
			return fmt.Errorf("client not found: %s", id)
		}

		return json.Unmarshal(v, &c)
	})

	return c, err
}

func (d *Database) AddClient(c Client) error {
	if c.Id == uuid.Nil {
		return errors.New("client id is missing")
	} else if c.Secret == "" {
		return errors.New("client secret is missing")
	} else if strings.Contains(c.Id.String(), ":") {
		return errors.New("client id cannot contain ':'")
	}

	return d.bdb.Update(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(BucketClients))

		data, err := json.Marshal(c)
		if err != nil {
			return err
		}

		return b.Put(clientKey(c.Id), data)
	})
}

func (d *Database) RemoveClient(id uuid.UUID) error {
	return d.bdb.Update(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(BucketClients))
		return b.Delete(clientKey(id))
	})
}

func (d *Database) GetAllClients() ([]Client, error) {
	var clients []Client
	err := d.bdb.View(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(BucketClients))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var c Client
			err := json.Unmarshal(v, &c)
			if err != nil {
				return err
			}

			clients = append(clients, c)
		}

		return nil
	})

	return clients, err
}
