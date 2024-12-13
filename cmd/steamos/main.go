/*
Zaparoo Core
Copyright (C) 2023 Gareth Jones
Copyright (C) 2023, 2024 Callan Barrett

This file is part of Zaparoo Core.

Zaparoo Core is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Zaparoo Core is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Zaparoo Core.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"flag"
	"fmt"
	"github.com/ZaparooProject/zaparoo-core/pkg/cli"
	"github.com/ZaparooProject/zaparoo-core/pkg/config"
	"github.com/ZaparooProject/zaparoo-core/pkg/platforms/steamos"
	"github.com/ZaparooProject/zaparoo-core/pkg/service"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"

	_ "embed"
)

// TODO: fix permissions on files in ~/zaparoo so root doesn't lock them

func main() {
	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	pl := &steamos.Platform{}
	flags := cli.SetupFlags()

	doInstall := flag.Bool("install", false, "install zaparoo service")
	doUninstall := flag.Bool("uninstall", false, "uninstall zaparoo service")

	flags.Pre(pl)
	cfg := cli.Setup(pl, &config.UserConfig{
		TapTo: config.TapToConfig{
			ProbeDevice:    true,
			ConsoleLogging: true,
		},
		Api: config.ApiConfig{
			Port: config.DefaultApiPort,
		},
	})

	if *doInstall {
		err := install()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error installing service: %v\n", err)
			log.Error().Err(err).Msg("error installing service")
			os.Exit(1)
		}
		fmt.Println("Service installed, restart SteamOS")
		os.Exit(0)
	} else if *doUninstall {
		err := uninstall()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error uninstalling service: %v\n", err)
			log.Error().Err(err).Msg("error uninstalling service")
			os.Exit(1)
		}
		fmt.Println("Service uninstalled")
		os.Exit(0)
	}

	flags.Post(cfg)

	stop, err := service.Start(pl, cfg)
	if err != nil {
		log.Error().Err(err).Msg("error starting service")
		os.Exit(1)
	}

	<-sigs
	err = stop()
	if err != nil {
		log.Error().Err(err).Msg("error stopping service")
		os.Exit(1)
	}

	os.Exit(0)
}
