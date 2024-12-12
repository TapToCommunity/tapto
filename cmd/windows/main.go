//go:build windows

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
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/ZaparooProject/zaparoo-core/pkg/platforms/windows"
	"github.com/ZaparooProject/zaparoo-core/pkg/utils"

	"github.com/ZaparooProject/zaparoo-core/pkg/config"
	"github.com/ZaparooProject/zaparoo-core/pkg/service"
)

func main() {
	versionOpt := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versionOpt {
		fmt.Println("TapTo v" + config.Version + " (windows)")
		os.Exit(0)
	}

	cfg, err := config.NewUserConfig(&config.UserConfig{
		TapTo: config.TapToConfig{
			ProbeDevice:    true,
			ConsoleLogging: true,
		},
		Api: config.ApiConfig{
			Port: config.DefaultApiPort,
		},
	})
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	pl := &windows.Platform{}
	err = utils.InitLogging(cfg, pl)
	if err != nil {
		fmt.Println("Error initializing logging:", err)
		os.Exit(1)
	}

	if cfg.GetDebug() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	fmt.Println("TapTo v" + config.Version)

	stopSvc, err := service.Start(pl, cfg)
	if err != nil {
		log.Error().Msgf("error starting service: %s", err)
		fmt.Println("Error starting service:", err)
		os.Exit(1)
	}

	ip, err := utils.GetLocalIp()
	if err != nil {
		fmt.Println("Device address: Unknown")
	} else {
		fmt.Println("Device address:", ip.String())
	}

	fmt.Println("Press Enter to exit")
	_, _ = fmt.Scanln()

	err = stopSvc()
	if err != nil {
		log.Error().Msgf("error stopping service: %s", err)
		fmt.Println("Error stopping service:", err)
		os.Exit(1)
	}

	os.Exit(0)
}
