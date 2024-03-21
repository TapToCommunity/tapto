/*
TapTo
Copyright (C) 2023 Gareth Jones
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

package daemon

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/launcher"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
)

type TokenQueue struct {
	tokens chan Token
}

func (q *TokenQueue) Enqueue(t Token) {
	q.tokens <- t
}

func (q *TokenQueue) Dequeue() Token {
	return <-q.tokens
}

func (q *TokenQueue) Close() {
	close(q.tokens)
}

func NewTokenQueue() *TokenQueue {
	return &TokenQueue{
		tokens: make(chan Token),
	}
}

func writeScanResult(card Token) error {
	f, err := os.Create(mister.TokenReadFile)
	if err != nil {
		return fmt.Errorf("unable to create scan result file %s: %s", mister.TokenReadFile, err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	_, err = f.WriteString(fmt.Sprintf("%s,%s", card.UID, card.Text))
	if err != nil {
		return fmt.Errorf("unable to write scan result file %s: %s", mister.TokenReadFile, err)
	}

	return nil
}

func inExitGameBlocklist(cfg *config.UserConfig) bool {
	var blocklist []string
	for _, v := range cfg.TapTo.ExitGameBlocklist {
		blocklist = append(blocklist, strings.ToLower(v))
	}
	return slices.Contains(blocklist, strings.ToLower(mister.GetActiveCoreName()))
}

func launchCard(cfg *config.UserConfig, state *State, kbd input.Keyboard) error {
	card := state.GetActiveCard()
	uidMap, textMap := state.GetDB()

	text := card.Text
	override := false

	if v, ok := uidMap[card.UID]; ok {
		log.Info().Msg("launching with uid match override")
		text = v
		override = true
	}

	if v, ok := textMap[card.Text]; ok {
		log.Info().Msg("launching with text match override")
		text = v
		override = true
	}

	if text == "" {
		return fmt.Errorf("no text NDEF found in card or database")
	}

	log.Info().Msgf("launching with text: %s", text)
	cmds := strings.Split(text, "||")

	for i, cmd := range cmds {
		err := launcher.LaunchToken(cfg, cfg.TapTo.AllowCommands || override, kbd, cmd, len(cmds), i)
		if err != nil {
			return err
		}
	}

	return nil
}

func processLaunchQueue(cfg *config.UserConfig, state *State, tq *TokenQueue, kbd input.Keyboard) {
	for {
		select {
		case t := <-tq.tokens:
			state.SetActiveCard(t)

			if state.IsLauncherDisabled() {
				continue
			}

			err := launchCard(cfg, state, kbd)
			if err != nil {
				log.Error().Msgf("error launching card: %s", err)
			}
		case <-time.After(1 * time.Second):
			if state.ShouldStopService() {
				tq.Close()
				return
			}
		}
	}
}

func StartDaemon(cfg *config.UserConfig) (func() error, error) {
	state := &State{}
	tq := NewTokenQueue()

	// TODO: this is platform specific
	kbd, err := input.NewKeyboard()
	if err != nil {
		log.Error().Msgf("failed to initialize keyboard: %s", err)
		return nil, err
	}

	uids, texts, err := launcher.LoadDatabase()
	if err != nil {
		log.Error().Msgf("error loading database: %s", err)
	} else {
		state.SetDB(uids, texts)
	}

	closeDbWatcher, err := launcher.StartMappingsWatcher(
		state.GetDBLoadTime,
		state.SetDB,
	)
	if err != nil {
		log.Error().Msgf("error starting database watcher: %s", err)
	}

	if _, err := os.Stat(mister.DisableLaunchFile); err == nil {
		state.DisableLauncher()
	}

	go runApiServer(cfg, state, tq, kbd)
	go readerPollLoop(cfg, state, tq, kbd)
	go processLaunchQueue(cfg, state, tq, kbd)

	socket, err := StartSocketServer(state)
	if err != nil {
		log.Error().Msgf("error starting socket server: %s", err)
		return nil, err
	}

	return func() error {
		err := socket.Close()
		if err != nil {
			log.Warn().Msgf("error closing socket: %s", err)
		}
		tq.Close()
		state.StopService()
		if closeDbWatcher != nil {
			return closeDbWatcher()
		}
		return nil
	}, nil
}
