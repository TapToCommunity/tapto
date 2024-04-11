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
	"github.com/wizzomafizzo/tapto/pkg/daemon/api"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/database"
	"github.com/wizzomafizzo/tapto/pkg/launcher"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
)

func writeScanResult(card state.Token) error {
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
	for _, v := range cfg.GetExitGameBlocklist() {
		blocklist = append(blocklist, strings.ToLower(v))
	}
	return slices.Contains(blocklist, strings.ToLower(mister.GetActiveCoreName()))
}

func launchCard(cfg *config.UserConfig, state *state.State, db *database.Database, kbd input.Keyboard) error {
	card := state.GetActiveCard()
	uidMap, textMap := state.GetDB()

	text := card.Text
	override := false

	if v, ok := uidMap[card.UID]; ok {
		log.Info().Msg("launching with csv uid match override")
		text = v
		override = true
	}

	if v, err := db.GetUidMapping(card.UID); err == nil {
		if err != nil {
			log.Error().Err(err).Msgf("error getting db uid mapping")
		} else if v != "" {
			log.Info().Msg("launching with db uid match override")
			text = v
			override = true
		}
	}

	if v, ok := textMap[card.Text]; ok {
		log.Info().Msg("launching with csv text match override")
		text = v
		override = true
	}

	if text == "" {
		return fmt.Errorf("no text NDEF found in card or database")
	}

	log.Info().Msgf("launching with text: %s", text)
	cmds := strings.Split(text, "||")

	for i, cmd := range cmds {
		err := launcher.LaunchToken(cfg, cfg.GetAllowCommands() || override, kbd, cmd, len(cmds), i)
		if err != nil {
			return err
		}
	}

	return nil
}

func processLaunchQueue(
	cfg *config.UserConfig,
	st *state.State,
	tq *state.TokenQueue,
	db *database.Database,
	kbd input.Keyboard,
) {
	for {
		select {
		case t := <-tq.Tokens:
			st.SetActiveCard(t)

			err := writeScanResult(t)
			if err != nil {
				log.Error().Err(err).Msgf("error writing tmp scan result")
			}

			he := database.HistoryEntry{
				Time: t.ScanTime,
				UID:  t.UID,
				Text: t.Text,
			}

			if st.IsLauncherDisabled() {
				err = db.AddHistory(he)
				if err != nil {
					log.Error().Err(err).Msgf("error adding history")
				}
				continue
			}

			err = launchCard(cfg, st, db, kbd)
			if err != nil {
				log.Error().Err(err).Msgf("error launching card")
			}

			he.Success = err == nil
			err = db.AddHistory(he)
			if err != nil {
				log.Error().Err(err).Msgf("error adding history")
			}
		case <-time.After(1 * time.Second):
			if st.ShouldStopService() {
				tq.Close()
				return
			}
		}
	}
}

func StartDaemon(cfg *config.UserConfig) (func() error, error) {
	st := &state.State{}
	tq := state.NewTokenQueue()

	db, err := database.Open()
	if err != nil {
		log.Error().Err(err).Msgf("error opening database")
		return nil, err
	}

	// TODO: this is platform specific
	kbd, err := input.NewKeyboard()
	if err != nil {
		log.Error().Msgf("failed to initialize keyboard: %s", err)
		return nil, err
	}

	// TODO: this is platform specific
	tr, stopTr, err := mister.StartTracker(*mister.UserConfigToMrext(cfg))

	// TODO: this is platform specific
	err = mister.Setup(tr)
	if err != nil {
		log.Error().Msgf("error setting up mister platform: %s", err)
		return nil, err
	}

	uids, texts, err := launcher.LoadMappings()
	if err != nil {
		log.Error().Msgf("error loading mappings: %s", err)
	} else {
		st.SetDB(uids, texts)
	}

	closeMappingsWatcher, err := launcher.StartMappingsWatcher(
		st.GetDBLoadTime,
		st.SetDB,
	)
	if err != nil {
		log.Error().Msgf("error starting mappings watcher: %s", err)
	}

	if _, err := os.Stat(mister.DisableLaunchFile); err == nil {
		st.DisableLauncher()
	}

	go api.RunApiServer(cfg, st, tq, db, tr)
	go readerPollLoop(cfg, st, tq, kbd)
	go processLaunchQueue(cfg, st, tq, db, kbd)

	socket, err := StartSocketServer(st)
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

		st.StopService()

		err = stopTr()
		if err != nil {
			log.Warn().Msgf("error stopping tracker: %s", err)
		}

		if closeMappingsWatcher != nil {
			return closeMappingsWatcher()
		}
		return nil
	}, nil
}
