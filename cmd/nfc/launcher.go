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
	"github.com/gocarina/gocsv"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"io"
	"os"
	"strings"
)

// Breviceps (https://freesound.org/people/Breviceps/sounds/445978/)
// Licence: CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
//
//go:embed sounds/success.wav
var successSound []byte

// PaulMorek (https://freesound.org/people/PaulMorek/sounds/330046/)
// Licence: CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
//
//go:embed sounds/fail.wav
var failSound []byte

type NfcMappingEntry struct {
	MatchUID  string `csv:"match_uid"`
	MatchText string `csv:"match_text"`
	Text      string `csv:"text"`
}

func loadDatabase(state *ServiceState) error {
	uids := make(map[string]string)
	texts := make(map[string]string)

	if _, err := os.Stat(config.NfcDatabaseFile); errors.Is(err, os.ErrNotExist) {
		logger.Info("no database file found, skipping")
		return nil
	}

	f, err := os.Open(config.NfcDatabaseFile)
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
	for i, entry := range entries {
		if entry.MatchUID == "" && entry.MatchText == "" {
			logger.Warn("entry %d has no UID or text, skipping", i+1)
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
	logger.Info("loaded %d entries from database", count)

	state.SetDB(uids, texts)

	return nil
}

func launchCard(cfg *config.UserConfig, state *ServiceState, kbd input.Keyboard) error {
	card := state.GetActiveCard()
	uidMap, textMap := state.GetDB()

	text := card.Text
	override := false

	if v, ok := uidMap[card.UID]; ok {
		logger.Info("launching with uid match override")
		text = v
		override = true
	}

	if v, ok := textMap[card.Text]; ok {
		logger.Info("launching with text match override")
		text = v
		override = true
	}

	if text == "" {
		return fmt.Errorf("no text NDEF found in card or database")
	}

	logger.Info("launching with text: %s", text)
	cmds := strings.Split(text, "||")

	for _, cmd := range cmds {
		err := mister.LaunchToken(cfg, cfg.Nfc.AllowCommands || override, kbd, cmd)
		if err != nil {
			return err
		}
	}

	return nil
}
