package state

import (
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
)

const (
	ReaderTypePN532   = "PN532"
	ReaderTypeACR122U = "ACR122U"
	ReaderTypeUnknown = "Unknown"
)

type Token struct {
	Type     string
	UID      string
	Text     string
	Data     string
	ScanTime time.Time
	FromApi  bool
}

type State struct {
	mu                      sync.RWMutex
	updateHook              *func(st *State)
	readerConnected         bool
	readerType              string
	activeCard              Token
	lastScanned             Token
	stopService             bool
	disableLauncher         bool
	writeRequest            string
	writeError              error
	dbLoadTime              time.Time
	uidMap                  map[string]string
	textMap                 map[string]string
	cardRemovalTime         time.Time
	currentlyLoadedSoftware string
}

func (s *State) SetUpdateHook(hook *func(st *State)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateHook = hook
}

func (s *State) SetActiveCard(card Token) {
	s.mu.Lock()

	if s.activeCard == card {
		// ignore duplicate scans
		s.mu.Unlock()
		return
	}

	s.activeCard = card
	if s.activeCard.UID != "" {
		s.lastScanned = card
	}

	s.mu.Unlock()

	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) SetCardRemovalTime(removalTime time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cardRemovalTime = removalTime
}

func (s *State) SetCurrentlyLoadedSoftware(command string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentlyLoadedSoftware = command
	log.Debug().Msgf("current software launched set to: %s", command)
}

func (s *State) GetActiveCard() Token {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeCard
}

func (s *State) GetLastScanned() Token {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastScanned
}

func (s *State) GetCardRemovalTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cardRemovalTime
}

func (s *State) GetCurrentlyLoadedSoftware() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentlyLoadedSoftware
}

func (s *State) StopService() {
	s.mu.Lock()
	s.stopService = true
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) ShouldStopService() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stopService
}

func (s *State) DisableLauncher() {
	s.mu.Lock()
	s.disableLauncher = true
	if _, err := os.Create(mister.DisableLaunchFile); err != nil {
		log.Error().Msgf("cannot create disable launch file: %s", err)
	}
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) EnableLauncher() {
	s.mu.Lock()
	s.disableLauncher = false
	if err := os.Remove(mister.DisableLaunchFile); err != nil {
		log.Error().Msgf("cannot remove disable launch file: %s", err)
	}
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) IsLauncherDisabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.disableLauncher
}

type OldDb struct {
	Uids  map[string]string
	Texts map[string]string
}

func (s *State) GetDB() OldDb {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return OldDb{
		Uids:  s.uidMap,
		Texts: s.textMap,
	}
}

func (s *State) GetDBLoadTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dbLoadTime
}

func (s *State) SetDB(uidMap map[string]string, textMap map[string]string) {
	s.mu.Lock()
	s.dbLoadTime = time.Now()
	s.uidMap = uidMap
	s.textMap = textMap
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) SetReaderConnected(rt string) {
	s.mu.Lock()
	s.readerConnected = true
	s.readerType = rt
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) SetReaderDisconnected() {
	s.mu.Lock()
	s.readerConnected = false
	s.readerType = ""
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) GetReaderStatus() (bool, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.readerConnected, s.readerType
}

func (s *State) SetWriteRequest(req string) {
	s.mu.Lock()
	s.writeRequest = req
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) GetWriteRequest() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.writeRequest
}

func (s *State) SetWriteError(err error) {
	s.mu.Lock()
	s.writeError = err
	s.mu.Unlock()
}

func (s *State) GetWriteError() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.writeError
}
