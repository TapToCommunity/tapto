//go:build linux || darwin

package mister

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"

	"github.com/wizzomafizzo/mrext/pkg/metadata"
	"github.com/wizzomafizzo/mrext/pkg/utils"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

const ArcadeSystem = "Arcade"

type NameMapping struct {
	CoreName   string
	System     string
	Name       string // TODO: use names.txt
	ArcadeName string
}

type Tracker struct {
	Config           *config.UserConfig
	mu               sync.Mutex
	eventHook        *func()
	ActiveCore       string
	ActiveSystem     string
	ActiveSystemName string
	ActiveGameId     string
	ActiveGameName   string
	ActiveGamePath   string
	NameMap          []NameMapping
}

func generateNameMap() []NameMapping {
	nameMap := make([]NameMapping, 0)

	for _, system := range games.Systems {
		if system.SetName != "" {
			nameMap = append(nameMap, NameMapping{
				CoreName: system.SetName,
				System:   system.Id,
				Name:     system.Name,
			})
		} else if len(system.Folder) > 0 {
			nameMap = append(nameMap, NameMapping{
				CoreName: system.Folder[0],
				System:   system.Id,
				Name:     system.Name,
			})
		} else {
			log.Warn().Msgf("system %s has no setname or folder", system.Id)
		}
	}

	arcadeDbEntries, err := metadata.ReadArcadeDb()
	if err != nil {
		log.Error().Msgf("error reading arcade db: %s", err)
	} else {
		for _, entry := range arcadeDbEntries {
			nameMap = append(nameMap, NameMapping{
				CoreName:   entry.Setname,
				System:     ArcadeSystem,
				Name:       ArcadeSystem,
				ArcadeName: entry.Name,
			})
		}
	}

	return nameMap
}

func NewTracker(cfg *config.UserConfig) (*Tracker, error) {
	log.Info().Msg("starting tracker")

	nameMap := generateNameMap()

	log.Info().Msgf("loaded %d name mappings", len(nameMap))

	return &Tracker{
		Config:           cfg,
		ActiveCore:       "",
		ActiveSystem:     "",
		ActiveSystemName: "",
		ActiveGameId:     "",
		ActiveGameName:   "",
		ActiveGamePath:   "",
		NameMap:          nameMap,
	}, nil
}

func (tr *Tracker) SetEventHook(hook *func()) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.eventHook = hook
}

func (tr *Tracker) runEventHook() {
	if tr.eventHook != nil {
		(*tr.eventHook)()
	}
}

func (tr *Tracker) ReloadNameMap() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	nameMap := generateNameMap()
	log.Info().Msgf("reloaded %d name mappings", len(nameMap))
	tr.NameMap = nameMap
}

func (tr *Tracker) LookupCoreName(name string, path string) NameMapping {
	log.Debug().Msgf("looking up name: %s", name)
	log.Debug().Msgf("file path: %s", path)

	for _, mapping := range tr.NameMap {
		if len(mapping.CoreName) != len(name) {
			continue
		}

		if !strings.EqualFold(mapping.CoreName, name) {
			continue
		} else if mapping.ArcadeName != "" {
			log.Debug().Msgf("arcade name: %s", mapping.ArcadeName)
			return mapping
		}

		sys, err := games.BestSystemMatch(tr.Config, path)
		if err != nil {
			log.Debug().Msgf("error finding system for game %s, %s: %s", name, path, err)
			continue
		}

		if sys.Id == "" {
			log.Debug().Msgf("no system found for game: %s, %s", name, path)
			continue
		}

		log.Info().Msgf("found mapping: %s -> %s", name, mapping.Name)
		return mapping
	}

	return NameMapping{}
}

func (tr *Tracker) stopCore() bool {
	if tr.ActiveCore != "" {
		if tr.ActiveCore == ArcadeSystem {
			tr.ActiveGameId = ""
			tr.ActiveGamePath = ""
			tr.ActiveGameName = ""
			tr.ActiveSystem = ""
			tr.ActiveSystemName = ""
		}

		tr.ActiveCore = ""

		return true
	} else {
		return false
	}
}

// LoadCore loads the current running core and set it as active.
func (tr *Tracker) LoadCore() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	data, err := os.ReadFile(config.CoreNameFile)
	coreName := string(data)

	if err != nil {
		log.Error().Msgf("error reading core name: %s", err)
		return
	}

	if coreName == config.MenuCore {
		err := mister.SetActiveGame("")
		if err != nil {
			log.Error().Msgf("error setting active game: %s", err)
		}
		coreName = ""
	}

	if coreName != tr.ActiveCore {
		tr.stopCore()

		tr.ActiveCore = coreName

		if coreName == "" {
			tr.stopGame()
			tr.runEventHook()
			return
		}

		result := tr.LookupCoreName(coreName, tr.ActiveGamePath)
		if result != (NameMapping{}) {
			if result.ArcadeName != "" {
				mister.SetActiveGame(result.CoreName)
				tr.ActiveGameId = coreName
				tr.ActiveGameName = result.ArcadeName
				tr.ActiveGamePath = "" // TODO: any way to find this?
				tr.ActiveSystem = ArcadeSystem
				tr.ActiveSystemName = ArcadeSystem
			}
		}

		tr.runEventHook()
	}
}

func (tr *Tracker) stopGame() bool {
	if tr.ActiveGameId != "" {
		tr.ActiveGameId = ""
		tr.ActiveGamePath = ""
		tr.ActiveGameName = ""
		tr.ActiveSystem = ""
		tr.ActiveSystemName = ""
		return true
	} else {
		return false
	}
}

// Load the current running game and set it as active.
func (tr *Tracker) loadGame() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	activeGame, err := mister.GetActiveGame()
	if err != nil {
		log.Error().Msgf("error getting active game: %s", err)
		tr.stopGame()
		return
	} else if !filepath.IsAbs(activeGame) {
		// arcade game, ignore handling
		// TODO: will this work ok long term?
		return
	} else if activeGame == "" {
		tr.stopGame()
		return
	}

	path := mister.ResolvePath(activeGame)
	filename := filepath.Base(path)
	name := utils.RemoveFileExt(filename)

	if filepath.Ext(strings.ToLower(filename)) == ".mgl" {
		mgl, err := mister.ReadMgl(path)
		if err != nil {
			log.Error().Msgf("error reading mgl: %s", err)
		} else {
			path = mister.ResolvePath(mgl.File.Path)
			log.Info().Msgf("mgl path: %s", path)
		}
	}

	system, err := games.BestSystemMatch(tr.Config, path)
	if err != nil {
		log.Error().Msgf("error finding system for game: %s", err)

		// temporary(?) workaround to ignore bug where presets loaded from
		// OSD are written to the recents file as a loaded game
		if strings.HasSuffix(strings.ToLower(filename), ".ini") {
			return
		}
	}

	id := fmt.Sprintf("%s/%s", system.Id, filename)

	if id != tr.ActiveGameId {
		tr.stopGame()

		tr.ActiveGameId = id
		tr.ActiveGameName = name
		tr.ActiveGamePath = path

		tr.ActiveSystem = system.Id
		tr.ActiveSystemName = system.Name

		tr.runEventHook()
	}
}

func (tr *Tracker) StopAll() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.stopCore()
	tr.stopGame()
}

// Read a core's recent file and attempt to write the newest entry's
// launch-able path to ACTIVEGAME.
func loadRecent(filename string) error {
	if !strings.Contains(filename, "_recent") {
		return nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening game file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error().Msgf("error closing file: %s", err)
		}
	}(file)

	recents, err := mister.ReadRecent(filename)
	if err != nil {
		return fmt.Errorf("error reading recent file: %w", err)
	} else if len(recents) == 0 {
		return nil
	}

	newest := recents[0]

	if strings.HasSuffix(filename, "cores_recent.cfg") {
		// main menu's recent file, written when launching mgls
		if strings.HasSuffix(strings.ToLower(newest.Name), ".mgl") {
			mglPath := mister.ResolvePath(filepath.Join(newest.Directory, newest.Name))
			mgl, err := mister.ReadMgl(mglPath)
			if err != nil {
				return fmt.Errorf("error reading mgl file: %w", err)
			}

			err = mister.SetActiveGame(mgl.File.Path)
			if err != nil {
				return fmt.Errorf("error setting active game: %w", err)
			}
		}
	} else {
		// individual core's recent file
		err = mister.SetActiveGame(filepath.Join(newest.Directory, newest.Name))
		if err != nil {
			return fmt.Errorf("error setting active game: %w", err)
		}
	}

	return nil
}

// StartFileWatch Start thread for monitoring changes to all files relating to core/game launches.
func StartFileWatch(tr *Tracker) (*fsnotify.Watcher, error) {
	log.Info().Msg("starting file watcher")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if event.Name == config.CoreNameFile {
						tr.LoadCore()
					} else if event.Name == config.ActiveGameFile {
						tr.loadGame()
					} else if strings.HasPrefix(event.Name, config.CoreConfigFolder) {
						err = loadRecent(event.Name)
						if err != nil {
							log.Error().Msgf("error loading recent file: %s", err)
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error().Msgf("error in watcher: %s", err)
			}
		}
	}()

	if _, err := os.Stat(config.CoreNameFile); os.IsNotExist(err) {
		err := os.WriteFile(config.CoreNameFile, []byte(""), 0644)
		if err != nil {
			return nil, err
		}
		log.Info().Msgf("created core name file: %s", config.CoreNameFile)
	}

	err = watcher.Add(config.CoreNameFile)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(config.CoreConfigFolder); os.IsNotExist(err) {
		err := os.MkdirAll(config.CoreConfigFolder, 0755)
		if err != nil {
			return nil, err
		}
		log.Info().Msgf("created core config folder: %s", config.CoreConfigFolder)
	}

	err = watcher.Add(config.CoreConfigFolder)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(config.ActiveGameFile); os.IsNotExist(err) {
		err := os.WriteFile(config.ActiveGameFile, []byte(""), 0644)
		if err != nil {
			return nil, err
		}
		log.Info().Msgf("created active game file: %s", config.ActiveGameFile)
	}

	err = watcher.Add(config.ActiveGameFile)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(config.CurrentPathFile); os.IsNotExist(err) {
		err := os.WriteFile(config.CurrentPathFile, []byte(""), 0644)
		if err != nil {
			return nil, err
		}
		log.Info().Msgf("created current path file: %s", config.CurrentPathFile)
	}

	err = watcher.Add(config.CurrentPathFile)
	if err != nil {
		return nil, err
	}

	return watcher, nil
}

func StartTracker(cfg config.UserConfig) (*Tracker, func() error, error) {
	tr, err := NewTracker(&cfg)
	if err != nil {
		log.Error().Msgf("error creating tracker: %s", err)
		return nil, nil, err
	}

	tr.LoadCore()
	if !mister.ActiveGameEnabled() {
		err := mister.SetActiveGame("")
		if err != nil {
			log.Error().Msgf("error setting active game: %s", err)
		}
	}

	watcher, err := StartFileWatch(tr)
	if err != nil {
		log.Error().Msgf("error starting file watch: %s", err)
		return nil, nil, err
	}

	return tr, func() error {
		err := watcher.Close()
		if err != nil {
			log.Error().Msgf("error closing file watcher: %s", err)
		}
		tr.StopAll()
		return nil
	}, nil
}
