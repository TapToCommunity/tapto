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
	ScanTime time.Time
}

type State struct {
	mu              sync.Mutex
	readerConnected bool
	readerType      string
	activeCard      Token
	lastScanned     Token
	stopService     bool
	disableLauncher bool
	writeRequest    string
	dbLoadTime      time.Time
	uidMap          map[string]string
	textMap         map[string]string
}

func (s *State) SetActiveCard(card Token) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeCard = card
	if s.activeCard.UID != "" {
		s.lastScanned = card
	}
}

func (s *State) GetActiveCard() Token {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.activeCard
}

func (s *State) GetLastScanned() Token {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastScanned
}

func (s *State) StopService() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopService = true
}

func (s *State) ShouldStopService() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopService
}

func (s *State) DisableLauncher() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.disableLauncher = true
	if _, err := os.Create(mister.DisableLaunchFile); err != nil {
		log.Error().Msgf("cannot create disable launch file: %s", err)
	}
}

func (s *State) EnableLauncher() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.disableLauncher = false
	if err := os.Remove(mister.DisableLaunchFile); err != nil {
		log.Error().Msgf("cannot remove disable launch file: %s", err)
	}
}

func (s *State) IsLauncherDisabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.disableLauncher
}

func (s *State) GetDB() (map[string]string, map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.uidMap, s.textMap
}

func (s *State) GetDBLoadTime() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dbLoadTime
}

func (s *State) SetDB(uidMap map[string]string, textMap map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dbLoadTime = time.Now()
	s.uidMap = uidMap
	s.textMap = textMap
}

func (s *State) SetReaderConnected(rt string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.readerConnected = true
	s.readerType = rt
}

func (s *State) SetReaderDisconnected() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.readerConnected = false
	s.readerType = ""
}

func (s *State) GetReaderStatus() (bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readerConnected, s.readerType
}

func (s *State) SetWriteRequest(req string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writeRequest = req
}

func (s *State) GetWriteRequest() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writeRequest
}
