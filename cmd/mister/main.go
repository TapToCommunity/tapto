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

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/wizzomafizzo/tapto/pkg/daemon/api"
	"github.com/wizzomafizzo/tapto/pkg/launcher"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
	"github.com/wizzomafizzo/tapto/pkg/utils"

	gc "github.com/rthornton128/goncurses"
	"github.com/wizzomafizzo/mrext/pkg/curses"

	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/daemon"

	mrextMister "github.com/wizzomafizzo/mrext/pkg/mister"
)

// TODO: something like the nfc-list utility so new users with unsupported readers can help identify them
// TODO: play sound using go library
// TODO: would it be possible to unlock the OSD with a card?
// TODO: use a tag to signal that that next tag should have the active game written to it
// TODO: if it exists, use search.db instead of on demand index for random

const (
	appName    = "tapto"
	appVersion = config.Version
)

func addToStartup() error {
	var startup mrextMister.Startup

	err := startup.Load()
	if err != nil {
		return err
	}

	if !startup.Exists("mrext/" + appName) {
		err = startup.AddService("mrext/" + appName)
		if err != nil {
			return err
		}

		err = startup.Save()
		if err != nil {
			return err
		}
	}

	return nil
}

func handleWriteCommand(textToWrite string, svc *mister.Service, cfg *config.UserConfig) {
	log.Info().Msgf("writing text to tag: %s", textToWrite)

	if !svc.Running() {
		_, _ = fmt.Fprintln(os.Stderr, "TapTo service is not running, please start it before writing.")
		log.Error().Msg("TapTo service is not running, exiting")
		os.Exit(1)
	}

	body, err := json.Marshal(api.ReaderWriteRequest{Text: textToWrite})
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error encoding request:", err)
		log.Error().Msgf("error encoding request: %s", err)
		os.Exit(1)
	}

	resp, err := http.Post(
		// TODO: don't hardcode port
		"http://localhost:7497/api/v1/readers/0/write",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error sending request:", err)
		log.Error().Msgf("error sending request: %s", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, err := io.ReadAll(resp.Body)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Error reading response:", err)
			log.Error().Msgf("error reading response: %s", err)
			os.Exit(1)
		}

		_, _ = fmt.Fprintf(os.Stderr, "Error writing to tag: %s\n", strings.TrimSpace(string(errBody)))
		log.Error().Msgf("error writing to tag: %s", strings.TrimSpace(string(errBody)))
		os.Exit(1)
	}

	_, _ = fmt.Fprintln(os.Stderr, "Successfully wrote to tag.")
	log.Info().Msg("successfully wrote to tag")
	os.Exit(0)
}

func main() {
	svcOpt := flag.String("service", "", "manage TapTo service (start, stop, restart, status)")
	writeOpt := flag.String("write", "", "write text to tag")
	launchOpt := flag.String("launch", "", "execute given text as if it were a token")
	flag.Parse()

	err := utils.InitLogging()
	if err != nil {
		fmt.Println("Error initializing logging:", err)
		os.Exit(1)
	}

	cfg, err := config.NewUserConfig(appName, &config.UserConfig{
		TapTo: config.TapToConfig{
			ProbeDevice: true,
		},
	})
	if err != nil {
		log.Error().Msgf("error loading user config: %s", err)
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	log.Info().Msgf("TapTo v%s", appVersion)
	log.Info().Msgf("config path = %s", cfg.IniPath)
	log.Info().Msgf("app path = %s", cfg.AppPath)
	log.Info().Msgf("connection_string = %s", cfg.GetConnectionString())
	log.Info().Msgf("allow_commands = %t", cfg.GetAllowCommands())
	log.Info().Msgf("disable_sounds = %t", cfg.GetDisableSounds())
	log.Info().Msgf("probe_device = %t", cfg.GetProbeDevice())
	log.Info().Msgf("exit_game = %t", cfg.GetExitGame())
	log.Info().Msgf("exit_game_blocklist = %s", cfg.GetExitGameBlocklist())
	log.Info().Msgf("debug = %t", cfg.GetDebug())

	if cfg.GetDebug() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	svc, err := mister.NewService(mister.ServiceArgs{
		Name: appName,
		Entry: func() (func() error, error) {
			return daemon.StartDaemon(&mister.Platform{}, cfg)
		},
	})
	if err != nil {
		log.Error().Msgf("error creating service: %s", err)
		fmt.Println("Error creating service:", err)
		os.Exit(1)
	}

	if *writeOpt != "" {
		handleWriteCommand(*writeOpt, svc, cfg)
	}

	if *launchOpt != "" {
		// TODO: this is doubling up on the split logic in daemon
		cmds := strings.Split(*launchOpt, "||")
		for i, cmd := range cmds {
			err, _ := launcher.LaunchToken(&mister.Platform{}, cfg, true, cmd, len(cmds), i)
			if err != nil {
				log.Error().Msgf("error launching token: %s", err)
				fmt.Println("Error launching token:", err)
				os.Exit(1)
			}
		}

		os.Exit(0)
	}

	svc.ServiceHandler(svcOpt)

	stdscr, err := curses.Setup()
	if err != nil {
		log.Error().Msgf("starting curses: %s", err)
		os.Exit(1)
	}
	defer gc.End()

	err = tryAddStartup(stdscr)
	if err != nil {
		log.Error().Msgf("error adding startup: %s", err)
		os.Exit(1)
	}

	if !svc.Running() {
		err := svc.Start()
		if err != nil {
			log.Error().Msgf("error starting service: %s", err)
		}
	}

	err = displayServiceInfo(stdscr, svc)
	if err != nil {
		log.Error().Msgf("error displaying service info: %s", err)
	}
}
