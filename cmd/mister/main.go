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
	"flag"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/cli"
	"os"

	gc "github.com/rthornton128/goncurses"
	"github.com/wizzomafizzo/mrext/pkg/curses"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"

	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/service"

	mrextMister "github.com/wizzomafizzo/mrext/pkg/mister"
)

func addToStartup() error {
	var startup mrextMister.Startup

	err := startup.Load()
	if err != nil {
		return err
	}

	if !startup.Exists("mrext/" + config.AppName) {
		err = startup.AddService("mrext/" + config.AppName)
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

func main() {
	flags := cli.SetupFlags()
	serviceFlag := flag.String(
		"service",
		"",
		"manage TapTo service (start|stop|restart|status)",
	)
	addStartupFlag := flag.Bool(
		"add-startup",
		false,
		"add TapTo service to MiSTer startup if not already added",
	)

	pl := &mister.Platform{}
	flags.Pre(pl)

	if *addStartupFlag {
		err := addToStartup()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error adding to startup: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	cfg := cli.Setup(pl, &config.UserConfig{
		TapTo: config.TapToConfig{
			ProbeDevice: true,
		},
		Api: config.ApiConfig{
			Port: config.DefaultApiPort,
		},
	})

	svc, err := mister.NewService(mister.ServiceArgs{
		Entry: func() (func() error, error) {
			return service.Start(pl, cfg)
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("error creating service")
		_, _ = fmt.Fprintf(os.Stderr, "Error creating service: %v\n", err)
		os.Exit(1)
	}
	svc.ServiceHandler(serviceFlag)

	flags.Post(cfg)

	// display gui
	// assume gui is working from this point, don't print to stdout
	stdscr, err := curses.Setup()
	if err != nil {
		log.Error().Err(err).Msg("could not start curses")
		os.Exit(1)
	}
	defer gc.End()

	// offer to add service to MiSTer startup if it's not already there
	err = tryAddStartup(stdscr)
	if err != nil {
		log.Error().Err(err).Msgf("error displaying startup dialog")
		os.Exit(1)
	}

	// try to auto-start service if it's not running already
	if !svc.Running() {
		err := svc.Start()
		if err != nil {
			log.Error().Err(err).Msg("could not start service")
		}
	}

	// display main info gui
	err = displayServiceInfo(pl, cfg, stdscr, svc)
	if err != nil {
		log.Error().Err(err).Msg("error displaying info dialog")
	}
}
