package state

import (
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/readers"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
	"github.com/wizzomafizzo/tapto/pkg/utils"
)

type Notification struct {
	Method string `json:"method"`
	Params any    `json:"params,omitempty"`
}

type State struct {
	mu              sync.RWMutex
	activeCard      tokens.Token
	lastScanned     tokens.Token
	stopService     bool
	disableLauncher bool
	platform        platforms.Platform
	readers         map[string]readers.Reader
	softwareToken   *tokens.Token
	wroteToken      *tokens.Token
	Notifications   chan<- Notification
}

func NewState(platform platforms.Platform) *State {
	return &State{
		platform:      platform,
		readers:       make(map[string]readers.Reader),
		Notifications: make(chan<- Notification),
	}
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

	s.Notifications <- Notification{
		Method: "state.activeCard",
		Params: card,
	}
	s.mu.Unlock()
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
	s.Notifications <- Notification{
		Method: "state.launching",
		Params: false,
	}
	s.mu.Unlock()
}

func (s *State) EnableLauncher() {
	s.mu.Lock()
	s.disableLauncher = false
	if err := s.platform.SetLaunching(true); err != nil {
		log.Error().Msgf("cannot remove disable launch file: %s", err)
	}
	s.Notifications <- Notification{
		Method: "state.launching",
		Params: true,
	}
	s.mu.Unlock()
}

func (s *State) IsLauncherDisabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.disableLauncher
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
	s.Notifications <- Notification{
		Method: "state.readerChanged",
		Params: device,
	}
	s.mu.Unlock()
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
	s.Notifications <- Notification{
		Method: "state.readerRemoved",
		Params: device,
	}
	s.mu.Unlock()
}

func (s *State) ListReaders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var rs []string
	for k := range s.readers {
		rs = append(rs, k)
	}

	return rs
}

func (s *State) SetSoftwareToken(token *tokens.Token) {
	s.mu.Lock()
	s.softwareToken = token
	s.mu.Unlock()
}

func (s *State) GetSoftwareToken() *tokens.Token {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.softwareToken
}

func (s *State) SetWroteToken(token *tokens.Token) {
	s.mu.Lock()
	s.wroteToken = token
	s.mu.Unlock()
}

func (s *State) GetWroteToken() *tokens.Token {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.wroteToken
}
