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

package service

import (
	"fmt"
	"github.com/wizzomafizzo/tapto/pkg/api"
	"github.com/wizzomafizzo/tapto/pkg/service/playlists"
	"github.com/wizzomafizzo/tapto/pkg/service/tokens"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/database"
	"github.com/wizzomafizzo/tapto/pkg/launcher"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/service/state"
)

func inExitGameBlocklist(platform platforms.Platform, cfg *config.UserConfig) bool {
	var blocklist []string
	for _, v := range cfg.GetExitGameBlocklist() {
		blocklist = append(blocklist, strings.ToLower(v))
	}
	return slices.Contains(blocklist, strings.ToLower(platform.GetActiveLauncher()))
}

func launchToken(
	platform platforms.Platform,
	cfg *config.UserConfig,
	token tokens.Token,
	db *database.Database,
	lsq chan<- *tokens.Token,
	plsc playlists.PlaylistController,
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
			plsc,
			token,
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

func processTokenQueue(
	platform platforms.Platform,
	cfg *config.UserConfig,
	st *state.State,
	itq <-chan tokens.Token,
	db *database.Database,
	lsq chan<- *tokens.Token,
	plq chan *playlists.Playlist,
) {
	var activePlaylist *playlists.Playlist

	for {
		select {
		case pls := <-plq:
			log.Info().Msgf("processing playlist update: %v", pls)

			if pls == nil {
				if activePlaylist != nil {
					log.Info().Msg("clearing active playlist")
				} else {
					log.Debug().Msg("no active playlist to clear")
				}
				activePlaylist = nil
				continue
			} else if activePlaylist == nil {
				log.Info().Msg("setting new active playlist, launching token")
				activePlaylist = pls
				go func() {
					t := tokens.Token{
						Text:     pls.Current(),
						ScanTime: time.Now(),
						Source:   tokens.SourcePlaylist,
					}
					plsc := playlists.PlaylistController{
						Active: activePlaylist,
						Queue:  plq,
					}
					err := launchToken(platform, cfg, t, db, lsq, plsc)
					if err != nil {
						log.Error().Err(err).Msgf("error launching token")
					}
				}()
				continue
			} else {
				if pls.Current() == activePlaylist.Current() {
					log.Debug().Msg("playlist current token unchanged, skipping")
					continue
				}

				log.Info().Msg("updating active playlist, launching token")
				activePlaylist = pls
				go func() {
					t := tokens.Token{
						Text:     pls.Current(),
						ScanTime: time.Now(),
						Source:   tokens.SourcePlaylist,
					}
					plsc := playlists.PlaylistController{
						Active: activePlaylist,
						Queue:  plq,
					}
					err := launchToken(platform, cfg, t, db, lsq, plsc)
					if err != nil {
						log.Error().Err(err).Msgf("error launching token")
					}
				}()
				continue
			}
		case t := <-itq:
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
				plsc := playlists.PlaylistController{
					Active: activePlaylist,
					Queue:  plq,
				}

				err = launchToken(platform, cfg, t, db, lsq, plsc)
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
				break
			}
		}
	}
}

func Start(
	platform platforms.Platform,
	cfg *config.UserConfig,
) (func() error, error) {
	// TODO: define the notifications chan here instead of in state
	st, ns := state.NewState(platform)
	// TODO: convert this to a *token channel
	itq := make(chan tokens.Token)
	lsq := make(chan *tokens.Token)
	plq := make(chan *playlists.Playlist)

	log.Info().Msgf("Zaparoo v%s", config.Version)
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

	log.Debug().Msg("starting API service")
	go api.Start(platform, cfg, st, itq, db, ns)

	log.Debug().Msg("running platform setup")
	err = platform.Setup(cfg, st.Notifications)
	if err != nil {
		log.Error().Msgf("error setting up platform: %s", err)
		return nil, err
	}

	if !platform.LaunchingEnabled() {
		st.DisableLauncher()
	}

	log.Debug().Msg("starting reader manager")
	go readerManager(platform, cfg, st, itq, lsq)

	log.Debug().Msg("starting token queue manager")
	go processTokenQueue(platform, cfg, st, itq, db, lsq, plq)

	return func() error {
		err = platform.Stop()
		if err != nil {
			log.Warn().Msgf("error stopping platform: %s", err)
		}
		st.StopService()
		close(plq)
		close(lsq)
		close(itq)
		return nil
	}, nil
}
