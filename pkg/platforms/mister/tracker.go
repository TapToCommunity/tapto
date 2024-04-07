package mister

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"

	"github.com/wizzomafizzo/mrext/pkg/metadata"
	"github.com/wizzomafizzo/mrext/pkg/utils"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

const (
	EventActionCoreStart = iota
	EventActionCoreStop
	EventActionGameStart
	EventActionGameStop
)

const ArcadeSystem = "Arcade"

type EventAction struct {
	Timestamp  time.Time
	Action     int
	Target     string
	TargetPath string
	TotalTime  int // for recovery from power loss
	ActiveCore struct {
		Core       string
		System     string
		SystemName string
	}
	ActiveGame struct {
		Path string
		Name string
	}
}

type CoreTime struct {
	Name string
	Time int
}

type GameTime struct {
	Id     string
	Path   string
	Name   string
	Folder string
	Time   int
}

type NameMapping struct {
	CoreName   string
	System     string
	Name       string // TODO: use names.txt
	ArcadeName string
}

type Tracker struct {
	Config           *config.UserConfig
	mu               sync.Mutex
	eventHook        *func(tr *Tracker, action int, target string)
	ActiveCore       string
	ActiveSystem     string
	ActiveSystemName string
	ActiveGame       string
	ActiveGameName   string
	ActiveGamePath   string
	Events           []EventAction
	CoreTimes        map[string]CoreTime
	GameTimes        map[string]GameTime
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
		ActiveGame:       "",
		ActiveGameName:   "",
		ActiveGamePath:   "",
		Events:           []EventAction{},
		CoreTimes:        map[string]CoreTime{},
		GameTimes:        map[string]GameTime{},
		NameMap:          nameMap,
	}, nil
}

func (tr *Tracker) SetEventHook(hook *func(tr *Tracker, action int, target string)) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.eventHook = hook
}

func (tr *Tracker) ReloadNameMap() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	nameMap := generateNameMap()
	log.Info().Msgf("reloaded %d name mappings", len(nameMap))
	tr.NameMap = nameMap
}

func (tr *Tracker) LookupName(name string, game string) NameMapping {
	for _, mapping := range tr.NameMap {
		if len(mapping.CoreName) != len(name) {
			continue
		}

		if !strings.EqualFold(mapping.CoreName, name) {
			continue
		}

		sys, err := games.BestSystemMatch(tr.Config, game)
		if err != nil {
			continue
		}

		if sys.Id != mapping.System {
			continue
		}

		return mapping
	}

	return NameMapping{}
}

func (tr *Tracker) addEvent(action int, target string) {
	totalTime := 0

	if action == EventActionCoreStart || action == EventActionCoreStop {
		if ct, ok := tr.CoreTimes[target]; ok {
			totalTime = ct.Time
		}
	} else if action == EventActionGameStart || action == EventActionGameStop {
		if gt, ok := tr.GameTimes[target]; ok {
			totalTime = gt.Time
		}
	}

	ev := EventAction{
		Timestamp: time.Now(),
		Action:    action,
		Target:    target,
		TotalTime: totalTime,
	}

	core, ok := tr.CoreTimes[tr.ActiveCore]
	if ok {
		ev.ActiveCore.Core = core.Name
		ev.ActiveCore.System = tr.ActiveSystem
		ev.ActiveCore.SystemName = tr.ActiveSystemName
	}

	game, ok := tr.GameTimes[tr.ActiveGame]
	if ok {
		ev.ActiveGame.Path = game.Path
		ev.ActiveGame.Name = game.Name
	}

	targetTime, ok := tr.GameTimes[target]
	if ok {
		ev.TargetPath = targetTime.Path
	}

	tr.Events = append(tr.Events, ev)

	actionLabel := ""
	switch action {
	case EventActionCoreStart:
		actionLabel = "core started"
	case EventActionCoreStop:
		actionLabel = "core stopped"
	case EventActionGameStart:
		actionLabel = "game started"
	case EventActionGameStop:
		actionLabel = "game stopped"
	}

	log.Info().Msgf("%s: %s (%ds)", actionLabel, target, totalTime)

	if tr.eventHook != nil {
		(*tr.eventHook)(tr, action, target)
	}
}

func (tr *Tracker) stopCore() bool {
	if tr.ActiveCore != "" {
		tr.addEvent(EventActionCoreStop, tr.ActiveCore)

		if tr.ActiveCore == ArcadeSystem {
			tr.ActiveGame = ""
			tr.ActiveGamePath = ""
			tr.ActiveGameName = ""
			tr.addEvent(EventActionGameStop, ArcadeSystem)
		}

		tr.ActiveCore = ""
		tr.ActiveSystem = ""
		tr.ActiveSystemName = ""

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
		tr.stopCore()
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
			return
		}

		result := tr.LookupName(coreName, tr.ActiveGamePath)
		if result != (NameMapping{}) {
			tr.ActiveSystem = result.System
			tr.ActiveSystemName = result.Name

			if result.System == ArcadeSystem {
				tr.ActiveGame = coreName
				tr.ActiveGameName = result.ArcadeName
				tr.addEvent(EventActionGameStart, coreName)
			} else if result.System == "" {
				tr.ActiveSystem = coreName
				tr.ActiveSystemName = coreName
			}
		} else {
			tr.ActiveSystem = ""
			tr.ActiveSystemName = ""
		}

		if _, ok := tr.CoreTimes[coreName]; !ok {
			tr.CoreTimes[coreName] = CoreTime{
				Name: coreName,
				Time: 0,
			}
		}

		tr.addEvent(EventActionCoreStart, coreName)
	}
}

func (tr *Tracker) stopGame() bool {
	if tr.ActiveGame != "" {
		target := tr.ActiveGame
		tr.ActiveGame = ""
		tr.ActiveGamePath = ""
		tr.ActiveGameName = ""
		tr.addEvent(EventActionGameStop, target)
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
	}

	var folder string
	if err != nil && len(system.Folder) > 0 {
		folder = system.Folder[0]
	}

	id := fmt.Sprintf("%s/%s", system.Id, filename)

	if id != tr.ActiveGame {
		tr.stopGame()

		tr.ActiveGame = id
		tr.ActiveGameName = name
		tr.ActiveGamePath = path

		result := tr.LookupName(tr.ActiveCore, path)
		if result != (NameMapping{}) {
			tr.ActiveSystem = result.System
			tr.ActiveSystemName = result.Name

			if result.System == ArcadeSystem {
				tr.ActiveGame = tr.ActiveCore
				tr.ActiveGameName = result.ArcadeName
			}
		} else {
			tr.ActiveSystem = ""
			tr.ActiveSystemName = ""
		}

		if _, ok := tr.GameTimes[id]; !ok {
			tr.GameTimes[id] = GameTime{
				Id:     id,
				Path:   path,
				Name:   name,
				Folder: folder,
				Time:   0,
			}
		}

		tr.addEvent(EventActionGameStart, id)
	}
}

func (tr *Tracker) StopAll() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.stopCore()
	tr.stopGame()
}

// Increment time of active core and game.
func (tr *Tracker) tick() {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if tr.ActiveCore != "" {
		if ct, ok := tr.CoreTimes[tr.ActiveCore]; ok {
			ct.Time++
			tr.CoreTimes[tr.ActiveCore] = ct
		}
	}

	if tr.ActiveGame != "" {
		if gt, ok := tr.GameTimes[tr.ActiveGame]; ok {
			gt.Time++
			tr.GameTimes[tr.ActiveGame] = gt
		}
	}
}

// StartTicker starts the thread for updating core/game play times.
func (tr *Tracker) StartTicker() {
	log.Info().Msgf("starting tracker ticker")
	ticker := time.NewTicker(time.Second)
	go func() {
		count := 0
		for range ticker.C {
			tr.tick()
			count++
		}
	}()
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

	err = watcher.Add(config.CoreNameFile)
	if err != nil {
		return nil, err
	}

	err = watcher.Add(config.CoreConfigFolder)
	if err != nil {
		return nil, err
	}

	err = watcher.Add(config.ActiveGameFile)
	if err != nil {
		return nil, err
	}

	_, fileExistsError := os.Stat(config.CurrentPathFile)
	if fileExistsError == nil {
		err = watcher.Add(config.CurrentPathFile)
		if err != nil {
			return nil, err
		}
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

	tr.StartTicker()

	return tr, func() error {
		err := watcher.Close()
		if err != nil {
			log.Error().Msgf("error closing file watcher: %s", err)
		}
		tr.StopAll()
		return nil
	}, nil
}
