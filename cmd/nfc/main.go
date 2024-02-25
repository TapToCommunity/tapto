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
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
	"github.com/wizzomafizzo/tapto/pkg/utils"

	gc "github.com/rthornton128/goncurses"
	"github.com/wizzomafizzo/mrext/pkg/curses"

	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/daemon"

	"github.com/clausecker/nfc/v2"
	mrextMister "github.com/wizzomafizzo/mrext/pkg/mister"
)

// TODO: something like the nfc-list utility so new users with unsupported readers can help identify them
// TODO: play sound using go library
// TODO: would it be possible to unlock the OSD with a card?
// TODO: create a test web nfc reader in separate github repo, hosted on pages
// TODO: use a tag to signal that that next tag should have the active game written to it
// TODO: if it exists, use search.db instead of on demand index for random

const (
	appName    = "tapto"
	appVersion = "1.1"
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

func handleWriteCommand(textToWrite string, svc *mister.Service, cfg config.TapToConfig) {
	// TODO: this is very tightly coupled to the mister service handling, it should
	//       be made a part of the daemon process itself without killing it

	serviceRunning := svc.Running()
	if serviceRunning {
		err := svc.Stop()
		if err != nil {
			log.Error().Msgf("error stopping service: %s", err)
			_, _ = fmt.Fprintln(os.Stderr, "Error stopping service:", err)
			os.Exit(1)
		}

		tries := 15
		for {
			if !svc.Running() {
				break
			}
			time.Sleep(100 * time.Millisecond)
			tries--
			if tries <= 0 {
				log.Error().Msgf("error stopping service: %s", err)
				_, _ = fmt.Fprintln(os.Stderr, "Error stopping service:", err)
				os.Exit(1)
			}
		}
	}

	restartService := func() {
		if serviceRunning {
			err := svc.Start()
			if err != nil {
				log.Error().Msgf("error starting service: %s", err)
				_, _ = fmt.Fprintln(os.Stderr, "Error starting service:", err)
				os.Exit(1)
			}
		}
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	cancelled := make(chan bool, 1)

	go func() {
		<-sigs
		cancelled <- true
		restartService()
		os.Exit(0)
	}()

	var pnd nfc.Device
	var err error

	pnd, err = daemon.OpenDeviceWithRetries(cfg, &daemon.State{})
	if err != nil {
		log.Error().Msgf("giving up, exiting: %s")
		_, _ = fmt.Fprintln(os.Stderr, "Could not open device:", err)
		restartService()
		os.Exit(1)
	}

	defer func(pnd nfc.Device) {
		err := pnd.Close()
		if err != nil {
			log.Warn().Msgf("error closing device: %s", err)
		}
		log.Info().Msg("closed nfc device")
	}(pnd)

	var count int
	var target nfc.Target
	tries := 6 // ~30 seconds

	for tries > 0 {
		count, target, err = pnd.InitiatorPollTarget(tokens.SupportedCardTypes, daemon.TimesToPoll, daemon.PeriodBetweenPolls)

		if err != nil && err.Error() != "timeout" {
			log.Error().Msgf("could not poll: %s", err)
			_, _ = fmt.Fprintln(os.Stderr, "Could not poll:", err)
			restartService()
			os.Exit(1)
		}

		if count > 0 {
			break
		}

		tries--
	}

	if count == 0 {
		log.Error().Msgf("could not detect a card")
		_, _ = fmt.Fprintln(os.Stderr, "Could not detect a card")
		restartService()
		os.Exit(1)
	}

	cardUid := tokens.GetCardUID(target)
	log.Info().Msgf("found card with UID: %s", cardUid)

	cardType := tokens.GetCardType(target)
	var bytesWritten []byte

	switch cardType {
	case tokens.TypeMifare:
		bytesWritten, err = tokens.WriteMifare(pnd, textToWrite, cardUid)
		if err != nil {
			log.Error().Msgf("error writing to card: %s", err)
			_, _ = fmt.Fprintln(os.Stderr, "Error writing to card:", err)
			fmt.Println("Mifare cards need to NDEF formatted. If this is a brand new card, please use NFC tools mobile app to write some text (this only needs to be done the first time)")
			restartService()
			os.Exit(1)
		}
	case tokens.TypeNTAG:
		bytesWritten, err = tokens.WriteNtag(pnd, textToWrite)
		if err != nil {
			log.Error().Msgf("error writing to card: %s", err)
			_, _ = fmt.Fprintln(os.Stderr, "Error writing to card:", err)
			restartService()
			os.Exit(1)
		}
	default:
		log.Error().Msgf("unsupported card type: %s", cardType)
		restartService()
		os.Exit(1)
	}

	log.Info().Msgf("successfully wrote to card: %s", hex.EncodeToString(bytesWritten))
	_, _ = fmt.Fprintln(os.Stderr, "Successfully wrote to card")

	restartService()
	os.Exit(0)
}

func main() {
	svcOpt := flag.String("service", "", "manage TapTo service (start, stop, restart, status)")
	writeOpt := flag.String("write", "", "write text to tag")
	flag.Parse()

	err := utils.InitLogging()
	if err != nil {
		fmt.Println("Error initializing logging:", err)
		os.Exit(1)
	}

	cfg, err := config.LoadUserConfig(appName, &config.UserConfig{
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
	log.Info().Msgf("connection_string = %s", cfg.TapTo.ConnectionString)
	log.Info().Msgf("allow_commands = %t", cfg.TapTo.AllowCommands)
	log.Info().Msgf("disable_sounds = %t", cfg.TapTo.DisableSounds)
	log.Info().Msgf("probe_device = %t", cfg.TapTo.ProbeDevice)
	log.Info().Msgf("exit_game = %t", cfg.TapTo.ExitGame)
	log.Info().Msgf("exit_game_blocklist = %s", cfg.TapTo.ExitGameBlocklist)
	log.Info().Msgf("debug = %t", cfg.TapTo.Debug)

	if cfg.TapTo.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	mister.NfcMigration()

	svc, err := mister.NewService(mister.ServiceArgs{
		Name: appName,
		Entry: func() (func() error, error) {
			return daemon.StartDaemon(cfg)
		},
	})
	if err != nil {
		log.Error().Msgf("error creating service: %s", err)
		fmt.Println("Error creating service:", err)
		os.Exit(1)
	}

	if *writeOpt != "" {
		handleWriteCommand(*writeOpt, svc, cfg.TapTo)
	}

	svc.ServiceHandler(svcOpt)

	interactive := true
	stdscr, err := curses.Setup()
	if err != nil {
		log.Error().Msgf("starting curses: %s", err)
		interactive = false
	}
	defer gc.End()

	if !interactive {
		err = addToStartup()
		if err != nil {
			log.Error().Msgf("error adding startup: %s", err)
			fmt.Println("Error adding to startup:", err)
		}
	} else {
		err = tryAddStartup(stdscr)
		if err != nil {
			log.Error().Msgf("error adding startup: %s", err)
		}
	}

	if !svc.Running() {
		err := svc.Start()
		if err != nil {
			log.Error().Msgf("error starting service: %s", err)
			if !interactive {
				fmt.Println("Error starting service:", err)
			}
			os.Exit(1)
		} else if !interactive {
			fmt.Println("Service started successfully.")
			os.Exit(0)
		}
	} else if !interactive {
		fmt.Println("Service is running.")
		os.Exit(0)
	}

	err = displayServiceInfo(stdscr, svc)
	if err != nil {
		log.Error().Msgf("error displaying service info: %s", err)
	}
}
