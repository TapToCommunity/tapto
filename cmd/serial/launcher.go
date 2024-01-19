/*
TapTo
Copyright (C) 2023, 2024 Callan Barrett

This file is part of TapTo.

TapTo is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

TapTo is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with TapTo.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocarina/gocsv"
	mrextConfig "github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

type Card struct {
	CardType string
	UID      string
	Text     string
	ScanTime time.Time
}

type ServiceState struct {
	mu              sync.Mutex
	activeCard      Card
	lastScanned     Card
	stopService     bool
	disableLauncher bool
	dbLoadTime      time.Time
	uidMap          map[string]string
	textMap         map[string]string
}

func (s *ServiceState) SetActiveCard(card Card) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeCard = card
	if s.activeCard.UID != "" {
		s.lastScanned = card
	}
}

func (s *ServiceState) GetActiveCard() Card {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.activeCard
}

func (s *ServiceState) GetLastScanned() Card {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastScanned
}

func (s *ServiceState) StopService() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopService = true
}

func (s *ServiceState) ShouldStopService() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopService
}

func (s *ServiceState) DisableLauncher() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.disableLauncher = true
}

func (s *ServiceState) EnableLauncher() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.disableLauncher = false
}

func (s *ServiceState) IsLauncherDisabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.disableLauncher
}

func (s *ServiceState) GetDB() (map[string]string, map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.uidMap, s.textMap
}

func (s *ServiceState) GetDBLoadTime() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dbLoadTime
}

func (s *ServiceState) SetDB(uidMap map[string]string, textMap map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dbLoadTime = time.Now()
	s.uidMap = uidMap
	s.textMap = textMap
}

type NfcMappingEntry struct {
	MatchUID  string `csv:"match_uid"`
	MatchText string `csv:"match_text"`
	Text      string `csv:"text"`
}

func loadDatabase(state *ServiceState) error {
	uids := make(map[string]string)
	texts := make(map[string]string)

	if _, err := os.Stat(mrextConfig.NfcDatabaseFile); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	f, err := os.Open(mrextConfig.NfcDatabaseFile)
	if err != nil {
		return err
	}
	defer func(c io.Closer) {
		_ = c.Close()
	}(f)

	entries := make([]NfcMappingEntry, 0)
	err = gocsv.Unmarshal(f, &entries)
	if err != nil {
		return err
	}

	count := 0
	for _, entry := range entries {
		if entry.MatchUID == "" && entry.MatchText == "" {
			continue
		}

		if entry.MatchUID != "" {
			uid := strings.TrimSpace(entry.MatchUID)
			uid = strings.ToLower(uid)
			uid = strings.ReplaceAll(uid, ":", "")
			uids[uid] = strings.TrimSpace(entry.Text)
		}

		if entry.MatchText != "" {
			text := strings.TrimSpace(entry.MatchText)
			texts[text] = strings.TrimSpace(entry.Text)
		}

		count++
	}

	state.SetDB(uids, texts)

	return nil
}

func launchCard(cfg *config.UserConfig, state *ServiceState, kbd input.Keyboard) error {
	card := state.GetActiveCard()
	uidMap, textMap := state.GetDB()

	text := card.Text
	override := false

	if v, ok := uidMap[card.UID]; ok {
		text = v
		override = true
	}

	if v, ok := textMap[card.Text]; ok {
		text = v
		override = true
	}

	if text == "" {
		return fmt.Errorf("no text NDEF found in card or database")
	}

	cmds := strings.Split(text, "||")

	for _, cmd := range cmds {
		err := tokens.LaunchToken(cfg, cfg.TapTo.AllowCommands || override, kbd, cmd)
		if err != nil {
			return err
		}
	}

	return nil
}
