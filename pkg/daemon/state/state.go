package state

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/readers"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

type State struct {
	mu              sync.RWMutex
	updateHook      *func(st *State)
	activeCard      tokens.Token
	lastScanned     tokens.Token
	stopService     bool
	disableLauncher bool
	dbLoadTime      time.Time
	uidMap          map[string]string
	textMap         map[string]string
	platform        platforms.Platform
	reader          readers.Reader
}

func (s *State) SetUpdateHook(hook *func(st *State)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateHook = hook
}

func (s *State) SetActiveCard(card tokens.Token) {
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

func (s *State) GetActiveCard() tokens.Token {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeCard
}

func (s *State) GetLastScanned() tokens.Token {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastScanned
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
	if err := s.platform.SetLauncherEnabled(false); err != nil {
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
	if err := s.platform.SetLauncherEnabled(true); err != nil {
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

func (s *State) SetReader(reader readers.Reader) {
	s.mu.Lock()
	s.reader = reader
	s.mu.Unlock()
}

func (s *State) GetReader() readers.Reader {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.reader
}
