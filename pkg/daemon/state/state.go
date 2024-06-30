package state

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/readers"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
	"github.com/wizzomafizzo/tapto/pkg/utils"
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
	readers         map[string]readers.Reader
	softwareToken   *tokens.Token
}

func NewState(platform platforms.Platform) *State {
	return &State{
		platform: platform,
		readers:  make(map[string]readers.Reader),
	}
}

func (s *State) SetUpdateHook(hook *func(st *State)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updateHook = hook
}

func (s *State) SetActiveCard(card tokens.Token) {
	s.mu.Lock()

	if utils.TokensEqual(&s.activeCard, &card) {
		// ignore duplicate scans
		s.mu.Unlock()
		return
	}

	s.activeCard = card
	if !s.activeCard.ScanTime.IsZero() {
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
	if err := s.platform.SetLaunching(false); err != nil {
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
	if err := s.platform.SetLaunching(true); err != nil {
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

func (s *State) GetReader(device string) (readers.Reader, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.readers[device]
	return r, ok
}

func (s *State) SetReader(device string, reader readers.Reader) {
	s.mu.Lock()

	r, ok := s.readers[device]
	if ok {
		err := r.Close()
		if err != nil {
			log.Warn().Err(err).Msg("error closing reader")
		}
	}

	s.readers[device] = reader
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) RemoveReader(device string) {
	s.mu.Lock()
	r, ok := s.readers[device]
	if ok && r != nil {
		err := r.Close()
		if err != nil {
			log.Warn().Err(err).Msg("error closing reader")
		}
	}
	delete(s.readers, device)
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) ListReaders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var readers []string
	for k := range s.readers {
		readers = append(readers, k)
	}

	return readers
}

func (s *State) SetSoftwareToken(token *tokens.Token) {
	s.mu.Lock()
	s.softwareToken = token
	s.mu.Unlock()
	if s.updateHook != nil {
		(*s.updateHook)(s)
	}
}

func (s *State) GetSoftwareToken() *tokens.Token {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.softwareToken
}
