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
	"regexp"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/daemon/api"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/database"
	"github.com/wizzomafizzo/tapto/pkg/launcher"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

func inExitGameBlocklist(platform platforms.Platform, cfg *config.UserConfig) bool {
	var blocklist []string
	for _, v := range cfg.GetExitGameBlocklist() {
		blocklist = append(blocklist, strings.ToLower(v))
	}
	return slices.Contains(blocklist, strings.ToLower(platform.GetActiveLauncher()))
}

func checkMappingUid(m database.Mapping, t tokens.Token) bool {
	uid := database.NormalizeUid(t.UID)

	switch {
	case m.Match == database.MatchTypeExact:
		return uid == m.Pattern
	case m.Match == database.MatchTypePartial:
		return strings.Contains(uid, m.Pattern)
	case m.Match == database.MatchTypeRegex:
		re, err := regexp.Compile(m.Pattern)
		if err != nil {
			log.Error().Err(err).Msgf("error compiling regex")
			return false
		}
		return re.MatchString(uid)
	}

	return false
}

func checkMappingText(m database.Mapping, t tokens.Token) bool {
	switch {
	case m.Match == database.MatchTypeExact:
		return t.Text == m.Pattern
	case m.Match == database.MatchTypePartial:
		return strings.Contains(t.Text, m.Pattern)
	case m.Match == database.MatchTypeRegex:
		re, err := regexp.Compile(m.Pattern)
		if err != nil {
			log.Error().Err(err).Msgf("error compiling regex")
			return false
		}
		return re.MatchString(t.Text)
	}

	return false
}

func checkMappingData(m database.Mapping, t tokens.Token) bool {
	switch {
	case m.Match == database.MatchTypeExact:
		return t.Data == m.Pattern
	case m.Match == database.MatchTypePartial:
		return strings.Contains(t.Data, m.Pattern)
	case m.Match == database.MatchTypeRegex:
		re, err := regexp.Compile(m.Pattern)
		if err != nil {
			log.Error().Err(err).Msgf("error compiling regex")
			return false
		}
		return re.MatchString(t.Data)
	}

	return false
}

func getMapping(db *database.Database, pl platforms.Platform, token tokens.Token) (string, bool) {
	// check db mappings
	ms, err := db.GetEnabledMappings()
	if err != nil {
		log.Error().Err(err).Msgf("error getting db mappings")
	}

	for _, m := range ms {
		switch {
		case m.Type == database.MappingTypeUID:
			if checkMappingUid(m, token) {
				log.Info().Msg("launching with db uid match override")
				return m.Override, true
			}
		case m.Type == database.MappingTypeText:
			if checkMappingText(m, token) {
				log.Info().Msg("launching with db text match override")
				return m.Override, true
			}
		case m.Type == database.MappingTypeData:
			if checkMappingData(m, token) {
				log.Info().Msg("launching with db data match override")
				return m.Override, true
			}
		}
	}

	// check platform mappings
	return pl.LookupMapping(token)
}

func launchToken(
	platform platforms.Platform,
	cfg *config.UserConfig,
	token tokens.Token,
	db *database.Database,
	lsq chan<- *tokens.Token,
) error {
	text := token.Text

	mappingText, mapped := getMapping(db, platform, token)
	if mapped {
		log.Info().Msgf("found mapping: %s", mappingText)
		text = mappingText
	}

	if text == "" {
		return fmt.Errorf("no text NDEF found in card or mappings")
	}

	log.Info().Msgf("launching with text: %s", text)
	cmds := strings.Split(text, "||")

	for i, cmd := range cmds {
		err, softwareSwap := launcher.LaunchToken(
			platform,
			cfg,
			cfg.GetAllowCommands() || mapped,
			cmd,
			len(cmds),
			i,
		)
		if err != nil {
			return err
		}

		if softwareSwap && !token.Remote {
			log.Info().Msgf("current software launched set to: %s", token.UID)
			lsq <- &token
		}
	}

	return nil
}

func processLaunchQueue(
	platform platforms.Platform,
	cfg *config.UserConfig,
	st *state.State,
	tq *tokens.TokenQueue,
	db *database.Database,
	lsq chan<- *tokens.Token,
) {
	for {
		select {
		case t := <-tq.Tokens:
			// TODO: change this channel to send a token pointer or something
			if t.ScanTime.IsZero() {
				// ignore empty tokens
				continue
			}

			log.Info().Msgf("processing token: %v", t)

			err := platform.AfterScanHook(t)
			if err != nil {
				log.Error().Err(err).Msgf("error writing tmp scan result")
			}

			he := database.HistoryEntry{
				Time: t.ScanTime,
				Type: t.Type,
				UID:  t.UID,
				Text: t.Text,
				Data: t.Data,
			}

			if st.IsLauncherDisabled() {
				err = db.AddHistory(he)
				if err != nil {
					log.Error().Err(err).Msgf("error adding history")
				}
				continue
			}

			// launch tokens in separate thread
			go func() {
				err = launchToken(platform, cfg, t, db, lsq)
				if err != nil {
					log.Error().Err(err).Msgf("error launching token")
				}

				he.Success = err == nil
				err = db.AddHistory(he)
				if err != nil {
					log.Error().Err(err).Msgf("error adding history")
				}
			}()
		case <-time.After(100 * time.Millisecond):
			if st.ShouldStopService() {
				tq.Close()
				return
			}
		}
	}
}

func StartDaemon(
	platform platforms.Platform,
	cfg *config.UserConfig,
) (func() error, error) {
	st := state.NewState(platform)
	tq := tokens.NewTokenQueue()
	lsq := make(chan *tokens.Token)

	log.Info().Msgf("TapTo v%s", config.Version)
	log.Info().Msgf("config path = %s", cfg.IniPath)
	log.Info().Msgf("app path = %s", cfg.AppPath)
	log.Info().Msgf("connection_string = %s", cfg.GetConnectionString())
	log.Info().Msgf("allow_commands = %t", cfg.GetAllowCommands())
	log.Info().Msgf("disable_sounds = %t", cfg.GetDisableSounds())
	log.Info().Msgf("probe_device = %t", cfg.GetProbeDevice())
	log.Info().Msgf("exit_game = %t", cfg.GetExitGame())
	log.Info().Msgf("exit_game_blocklist = %s", cfg.GetExitGameBlocklist())
	log.Info().Msgf("debug = %t", cfg.GetDebug())

	log.Debug().Msg("opening database")
	db, err := database.Open(platform)
	if err != nil {
		log.Error().Err(err).Msgf("error opening database")
		return nil, err
	}

	log.Debug().Msg("running platform setup")
	err = platform.Setup(cfg)
	if err != nil {
		log.Error().Msgf("error setting up platform: %s", err)
		return nil, err
	}

	if !platform.LaunchingEnabled() {
		st.DisableLauncher()
	}

	go api.RunApiServer(platform, cfg, st, tq, db)
	go readerManager(platform, cfg, st, tq, lsq)
	go processLaunchQueue(platform, cfg, st, tq, db, lsq)

	return func() error {
		tq.Close()
		st.StopService()
		err = platform.Stop()
		if err != nil {
			log.Warn().Msgf("error stopping platform: %s", err)
		}

		return nil
	}, nil
}
