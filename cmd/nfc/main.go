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

	"github.com/wizzomafizzo/tapto/pkg/tokens"
	"github.com/wizzomafizzo/tapto/pkg/utils"

	gc "github.com/rthornton128/goncurses"
	"github.com/wizzomafizzo/mrext/pkg/curses"

	"github.com/wizzomafizzo/mrext/pkg/service"

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
	appVersion = "1.0-beta4"
)

var (
	logger = service.NewLogger(appName)
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

func handleWriteCommand(textToWrite string, svc *service.Service, cfg config.TapToConfig) {
	// TODO: this is very tightly coupled to the mister service handling, it should
	//       be made a part of the daemon process itself without killing it

	serviceRunning := svc.Running()
	if serviceRunning {
		err := svc.Stop()
		if err != nil {
			logger.Error("error stopping service: %s", err)
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
				logger.Error("error stopping service: %s", err)
				_, _ = fmt.Fprintln(os.Stderr, "Error stopping service:", err)
				os.Exit(1)
			}
		}
	}

	restartService := func() {
		if serviceRunning {
			err := svc.Start()
			if err != nil {
				logger.Error("error starting service: %s", err)
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

	pnd, err = daemon.OpenDeviceWithRetries(cfg)
	if err != nil {
		logger.Error("giving up, exiting")
		_, _ = fmt.Fprintln(os.Stderr, "Could not open device:", err)
		restartService()
		os.Exit(1)
	}

	defer func(pnd nfc.Device) {
		err := pnd.Close()
		if err != nil {
			logger.Warn("error closing device: %s", err)
		}
		logger.Info("closed nfc device")
	}(pnd)

	var count int
	var target nfc.Target
	tries := 6 // ~30 seconds

	for tries > 0 {
		count, target, err = pnd.InitiatorPollTarget(tokens.SupportedCardTypes, daemon.TimesToPoll, daemon.PeriodBetweenPolls)

		if err != nil && err.Error() != "timeout" {
			logger.Error("could not poll: %s", err)
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
		logger.Error("could not detect a card")
		_, _ = fmt.Fprintln(os.Stderr, "Could not detect a card")
		restartService()
		os.Exit(1)
	}

	cardUid := tokens.GetCardUID(target)
	logger.Info("Found card with UID: %s", cardUid)

	cardType := tokens.GetCardType(target)
	var bytesWritten []byte

	switch cardType {
	case tokens.TypeMifare:
		bytesWritten, err = tokens.WriteMifare(pnd, textToWrite, cardUid)
		if err != nil {
			logger.Error("error writing to card: %s", err)
			_, _ = fmt.Fprintln(os.Stderr, "Error writing to card:", err)
			fmt.Println("Mifare cards need to NDEF formatted. If this is a brand new card, please use NFC tools mobile app to write some text (this only needs to be done the first time)")
			restartService()
			os.Exit(1)
		}
	case tokens.TypeNTAG:
		bytesWritten, err = tokens.WriteNtag(pnd, textToWrite)
		if err != nil {
			logger.Error("error writing to card: %s", err)
			_, _ = fmt.Fprintln(os.Stderr, "Error writing to card:", err)
			restartService()
			os.Exit(1)
		}
	default:
		logger.Error("Unsupported card type: %s", cardType)
		restartService()
		os.Exit(1)
	}

	logger.Info("successfully wrote to card: %s", hex.EncodeToString(bytesWritten))
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
		logger.Error("error loading user config: %s", err)
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	logger.Info("TapTo v%s", appVersion)
	logger.Info("config path: %s", cfg.IniPath)
	logger.Info("app path: %s", cfg.AppPath)
	logger.Info("connection_string: %s", cfg.TapTo.ConnectionString)
	logger.Info("allow_commands: %t", cfg.TapTo.AllowCommands)
	logger.Info("disable_sounds: %t", cfg.TapTo.DisableSounds)
	logger.Info("probe_device: %t", cfg.TapTo.ProbeDevice)
	logger.Info("exit_game: %t", cfg.TapTo.ExitGame)

	utils.NfcMigration(logger)

	svc, err := service.NewService(service.ServiceArgs{
		Name:   appName,
		Logger: logger,
		Entry: func() (func() error, error) {
			return daemon.StartService(cfg)
		},
	})
	if err != nil {
		logger.Error("error creating service: %s", err)
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
		logger.Error("starting curses: %s", err)
		interactive = false
	}
	defer gc.End()

	if !interactive {
		err = addToStartup()
		if err != nil {
			logger.Error("error adding startup: %s", err)
			fmt.Println("Error adding to startup:", err)
		}
	} else {
		err = tryAddStartup(stdscr)
		if err != nil {
			logger.Error("error adding startup: %s", err)
		}
	}

	if !svc.Running() {
		err := svc.Start()
		if err != nil {
			logger.Error("error starting service: %s", err)
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
		logger.Error("error displaying service info: %s", err)
	}
}
