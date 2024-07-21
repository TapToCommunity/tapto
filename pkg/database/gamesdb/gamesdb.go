package gamesdb

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	bolt "go.etcd.io/bbolt"
	"golang.org/x/sync/errgroup"

	"github.com/gobwas/glob"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/utils"
)

const (
	BucketNames       = "names"
	indexedSystemsKey = "meta:indexedSystems"
)

// Return the key for a name in the names index.
func NameKey(systemId string, name string) string {
	return systemId + ":" + name
}

// Check if the gamesdb exists on disk.
func DbExists(platform platforms.Platform) bool {
	_, err := os.Stat(filepath.Join(platform.ConfigFolder(), config.GamesDbFilename))
	return err == nil
}

// Open the gamesdb with the given options. If the database does not exist it
// will be created and the buckets will be initialized.
func open(platform platforms.Platform, options *bolt.Options) (*bolt.DB, error) {
	err := os.MkdirAll(filepath.Dir(filepath.Join(platform.ConfigFolder(), config.GamesDbFilename)), 0755)
	if err != nil {
		return nil, err
	}

	db, err := bolt.Open(filepath.Join(platform.ConfigFolder(), config.GamesDbFilename), 0600, options)
	if err != nil {
		return nil, err
	}

	db.Update(func(txn *bolt.Tx) error {
		for _, bucket := range []string{BucketNames} {
			_, err := txn.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return err
			}
		}

		return nil
	})

	return db, nil
}

// Open the gamesdb with default options for generating names index.
func openNames(platform platforms.Platform) (*bolt.DB, error) {
	return open(platform, &bolt.Options{
		NoSync:         true,
		NoFreelistSync: true,
	})
}

func readIndexedSystems(db *bolt.DB) ([]string, error) {
	var systems []string

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketNames))
		v := b.Get([]byte(indexedSystemsKey))
		if v != nil {
			systems = strings.Split(string(v), ",")
		}
		return nil
	})

	return systems, err
}

func writeIndexedSystems(db *bolt.DB, systems []string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketNames))
		v := b.Get([]byte(indexedSystemsKey))
		if v == nil {
			v = []byte(strings.Join(systems, ","))
			return b.Put([]byte(indexedSystemsKey), v)
		} else {
			existing := strings.Split(string(v), ",")
			for _, s := range systems {
				if !utils.Contains(existing, s) {
					existing = append(existing, s)
				}
			}
			return b.Put([]byte(indexedSystemsKey), []byte(strings.Join(existing, ",")))
		}
	})
}

type fileInfo struct {
	SystemId string
	Path     string
}

// Delete all names in index for the given system.
func deleteSystemNames(db *bolt.DB, systemId string) error {
	return db.Batch(func(tx *bolt.Tx) error {
		bns := tx.Bucket([]byte(BucketNames))

		c := bns.Cursor()
		for k, _ := c.Seek([]byte(systemId + ":")); k != nil && strings.HasPrefix(string(k), systemId+":"); k, _ = c.Next() {
			err := bns.Delete(k)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// Update the names index with the given files.
func updateNames(db *bolt.DB, files []fileInfo) error {
	return db.Batch(func(tx *bolt.Tx) error {
		bns := tx.Bucket([]byte(BucketNames))

		for _, file := range files {
			base := filepath.Base(file.Path)
			name := strings.TrimSuffix(base, filepath.Ext(base))

			nk := NameKey(file.SystemId, name)
			err := bns.Put([]byte(nk), []byte(file.Path))
			if err != nil {
				return err
			}
		}

		return nil
	})
}

type IndexStatus struct {
	Total    int
	Step     int
	SystemId string
	Files    int
}

// Given a list of systems, index all valid game files on disk and write a
// names index to the DB. Overwrites any existing names index, but does not
// clean up old missing files.
//
// Takes a function which will be called with the current status of the index
// during key steps.
//
// Returns the total number of files indexed.
func NewNamesIndex(
	platform platforms.Platform,
	cfg *config.UserConfig,
	systems []System,
	update func(IndexStatus),
) (int, error) {
	status := IndexStatus{
		Total: len(systems) + 1,
		Step:  1,
	}

	db, err := openNames(platform)
	if err != nil {
		return status.Files, fmt.Errorf("error opening gamesdb: %s", err)
	}
	defer db.Close()

	indexed, err := readIndexedSystems(db)
	if err != nil {
		log.Info().Msg("no indexed systems found")
	}

	for _, v := range indexed {
		err := deleteSystemNames(db, v)
		if err != nil {
			return status.Files, fmt.Errorf("error deleting system names: %s", err)
		} else {
			log.Debug().Msgf("deleted names for system: %s", v)
		}
	}

	update(status)
	systemPaths := make(map[string][]string, 0)
	for _, v := range GetSystemPaths(platform, platform.RootFolders(cfg), systems) {
		systemPaths[v.System.Id] = append(systemPaths[v.System.Id], v.Path)
	}

	g := new(errgroup.Group)

	for _, k := range utils.AlphaMapKeys(systemPaths) {
		status.SystemId = k
		status.Step++
		update(status)

		files := make([]fileInfo, 0)

		for _, path := range systemPaths[k] {
			pathFiles, err := GetFiles(platform, k, path)
			if err != nil {
				return status.Files, fmt.Errorf("error getting files: %s", err)
			}

			if len(pathFiles) == 0 {
				continue
			}

			for pf := range pathFiles {
				files = append(files, fileInfo{SystemId: k, Path: pathFiles[pf]})
			}
		}

		if len(files) == 0 {
			continue
		}

		status.Files += len(files)

		g.Go(func() error {
			return updateNames(db, files)
		})
	}

	status.Step++
	status.SystemId = ""
	update(status)

	err = g.Wait()
	if err != nil {
		return status.Files, fmt.Errorf("error updating names index: %s", err)
	}

	err = writeIndexedSystems(db, utils.AlphaMapKeys(systemPaths))
	if err != nil {
		return status.Files, fmt.Errorf("error writing indexed systems: %s", err)
	}

	err = db.Sync()
	if err != nil {
		return status.Files, fmt.Errorf("error syncing database: %s", err)
	}

	return status.Files, nil
}

type SearchResult struct {
	SystemId string
	Name     string
	Path     string
}

// Iterate all indexed names and return matches to test func against query.
func searchNamesGeneric(
	platform platforms.Platform,
	systems []System,
	query string,
	test func(string, string) bool,
) ([]SearchResult, error) {
	if !DbExists(platform) {
		return nil, fmt.Errorf("gamesdb does not exist")
	}

	db, err := open(platform, &bolt.Options{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var results []SearchResult

	err = db.View(func(tx *bolt.Tx) error {
		bn := tx.Bucket([]byte(BucketNames))

		for _, system := range systems {
			pre := []byte(system.Id + ":")
			nameIdx := bytes.Index(pre, []byte(":"))

			c := bn.Cursor()
			for k, v := c.Seek([]byte(pre)); k != nil && bytes.HasPrefix(k, pre); k, v = c.Next() {
				keyName := string(k[nameIdx+1:])

				if test(query, keyName) {
					results = append(results, SearchResult{
						SystemId: system.Id,
						Name:     keyName,
						Path:     string(v),
					})
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// Return indexed names matching exact query (case insensitive).
func SearchNamesExact(platform platforms.Platform, systems []System, query string) ([]SearchResult, error) {
	return searchNamesGeneric(platform, systems, query, func(query, keyName string) bool {
		return strings.EqualFold(query, keyName)
	})
}

// Return indexed names partially matching query (case insensitive).
func SearchNamesPartial(platform platforms.Platform, systems []System, query string) ([]SearchResult, error) {
	return searchNamesGeneric(platform, systems, query, func(query, keyName string) bool {
		return strings.Contains(strings.ToLower(keyName), strings.ToLower(query))
	})
}

// Return indexed names that include every word in query (case insensitive).
func SearchNamesWords(platform platforms.Platform, systems []System, query string) ([]SearchResult, error) {
	return searchNamesGeneric(platform, systems, query, func(query, keyName string) bool {
		qWords := strings.Fields(strings.ToLower(query))

		for _, word := range qWords {
			if !strings.Contains(strings.ToLower(keyName), word) {
				return false
			}
		}

		return true
	})
}

// Return indexed names matching query using regular expression.
func SearchNamesRegexp(platform platforms.Platform, systems []System, query string) ([]SearchResult, error) {
	return searchNamesGeneric(platform, systems, query, func(query, keyName string) bool {
		// TODO: this should be cached
		r, err := regexp.Compile(query)
		if err != nil {
			return false
		}

		return r.MatchString(keyName)
	})
}

var globCache = make(map[string]glob.Glob)
var globCacheMutex = &sync.RWMutex{}

func SearchNamesGlob(platform platforms.Platform, systems []System, query string) ([]SearchResult, error) {
	return searchNamesGeneric(platform, systems, query, func(query, keyName string) bool {
		globCacheMutex.RLock()
		cached, ok := globCache[query]
		if !ok {
			globCacheMutex.RUnlock()

			g, err := glob.Compile(query)
			if err != nil {
				return false
			}

			globCacheMutex.Lock()
			globCache[query] = g
			globCacheMutex.Unlock()
			cached = g
		} else {
			globCacheMutex.RUnlock()
		}

		return cached.Match(strings.ToLower(keyName))
	})
}

// Return true if a specific system is indexed in the gamesdb
func SystemIndexed(platform platforms.Platform, system System) bool {
	if !DbExists(platform) {
		return false
	}

	db, err := open(platform, &bolt.Options{ReadOnly: true})
	if err != nil {
		return false
	}
	defer db.Close()

	systems, err := readIndexedSystems(db)
	if err != nil {
		return false
	}

	return utils.Contains(systems, system.Id)
}

// Return all systems indexed in the gamesdb
func IndexedSystems(platform platforms.Platform) ([]string, error) {
	if !DbExists(platform) {
		return nil, fmt.Errorf("gamesdb does not exist")
	}

	db, err := open(platform, &bolt.Options{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer db.Close()

	systems, err := readIndexedSystems(db)
	if err != nil {
		return nil, err
	}

	return systems, nil
}

// Return a random game from specified systems.
func RandomGame(platform platforms.Platform, systems []System) (SearchResult, error) {
	if !DbExists(platform) {
		return SearchResult{}, fmt.Errorf("gamesdb does not exist")
	}

	db, err := open(platform, &bolt.Options{ReadOnly: true})
	if err != nil {
		return SearchResult{}, err
	}
	defer db.Close()

	var result SearchResult

	system, err := utils.RandomElem(systems)
	if err != nil {
		return result, err
	}

	possible := make([]SearchResult, 0)

	err = db.View(func(tx *bolt.Tx) error {
		bn := tx.Bucket([]byte(BucketNames))

		pre := []byte(system.Id + ":")
		nameIdx := bytes.Index(pre, []byte(":"))

		c := bn.Cursor()
		for k, v := c.Seek([]byte(pre)); k != nil && bytes.HasPrefix(k, pre); k, v = c.Next() {
			keyName := string(k[nameIdx+1:])
			possible = append(possible, SearchResult{
				SystemId: system.Id,
				Name:     keyName,
				Path:     string(v),
			})
		}

		return nil
	})
	if err != nil {
		return result, err
	}

	if len(possible) == 0 {
		return result, fmt.Errorf("no games found for system: %s", system.Id)
	}

	result, err = utils.RandomElem(possible)
	if err != nil {
		return result, err
	}

	return result, nil
}
