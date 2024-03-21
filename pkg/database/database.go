package database

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
	bolt "go.etcd.io/bbolt"
)

const (
	BucketHistory  = "history"
	BucketMappings = "mappings"
)

// Check if the db exists on disk.
func DbExists() bool {
	_, err := os.Stat(mister.DbFile)
	return err == nil
}

// Open the db with the given options. If the database does not exist it
// will be created and the buckets will be initialized.
func open(options *bolt.Options) (*bolt.DB, error) {
	err := os.MkdirAll(filepath.Dir(mister.DbFile), 0755)
	if err != nil {
		return nil, err
	}

	db, err := bolt.Open(mister.DbFile, 0600, options)
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
	UID     string    `json:"uid"`
	Text    string    `json:"text"`
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
	max := 100

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

const (
	MappingTypeUID  = "uid"
	MappingTypeText = "text"
)

func MappingsKey(mt, arg string) string {
	return mt + ":" + arg
}

func normalUid(uid string) string {
	uid = strings.TrimSpace(uid)
	uid = strings.ToLower(uid)
	uid = strings.ReplaceAll(uid, ":", "")
	return uid
}

func (d *Database) AddUidMapping(orig, mapping string) error {
	return d.bdb.Update(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(BucketMappings))
		return b.Put([]byte(MappingsKey(
			MappingTypeUID,
			normalUid(orig),
		)), []byte(normalUid(mapping)))
	})
}

func (d *Database) AddTextMapping(orig, mapping string) error {
	return d.bdb.Update(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(BucketMappings))
		return b.Put([]byte(MappingsKey(
			MappingTypeText,
			orig,
		)), []byte(mapping))
	})
}

func (d *Database) GetUidMapping(orig string) (string, error) {
	var mapping string
	err := d.bdb.View(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(BucketMappings))
		v := b.Get([]byte(MappingsKey(MappingTypeUID, normalUid(orig))))
		if v == nil {
			return nil
		}

		mapping = string(v)
		return nil
	})

	return mapping, err
}

func (d *Database) GetTextMapping(orig string) (string, error) {
	var mapping string
	err := d.bdb.View(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(BucketMappings))
		v := b.Get([]byte(MappingsKey(MappingTypeText, orig)))
		if v == nil {
			return nil
		}

		mapping = string(v)
		return nil
	})

	return mapping, err
}

type Mappings struct {
	Uids  map[string]string
	Texts map[string]string
}

func (d *Database) GetMappings() (Mappings, error) {
	var mappings Mappings
	err := d.bdb.View(func(txn *bolt.Tx) error {
		b := txn.Bucket([]byte(BucketMappings))

		mappings.Uids = make(map[string]string)
		mappings.Texts = make(map[string]string)

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			key := string(k)
			val := string(v)

			parts := strings.Split(key, ":")
			if len(parts) != 2 {
				return nil
			}

			if parts[0] == MappingTypeUID {
				mappings.Uids[parts[1]] = val
			} else if parts[0] == MappingTypeText {
				mappings.Texts[parts[1]] = val
			}
		}

		return nil
	})

	return mappings, err
}
