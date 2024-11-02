///go:build linux && cgo

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
	"github.com/wizzomafizzo/tapto/pkg/cli"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mistex"
	"github.com/wizzomafizzo/tapto/pkg/utils"
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"

	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
	"github.com/wizzomafizzo/tapto/pkg/service"
)

func tryAddToStartup() (bool, error) {
	unitPath := "/etc/systemd/system/tapto.service"
	unitFile := `[Unit]
Description=TapTo service

[Service]
Type=forking
Restart=no
ExecStart=/media/fat/Scripts/tapto.sh -service start

[Install]
WantedBy=multi-user.target
`

	_, err := os.Stat(unitPath)
	if err == nil {
		return false, nil
	}

	err = os.WriteFile(unitPath, []byte(unitFile), 0644)
	if err != nil {
		return false, err
	}

	cmd := exec.Command("systemctl", "daemon-reload")
	err = cmd.Run()
	if err != nil {
		return false, err
	}

	cmd = exec.Command("systemctl", "enable", "tapto.service")
	err = cmd.Run()
	if err != nil {
		return false, err
	}

	return true, nil
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

	pl := &mistex.Platform{}
	flags.Pre(pl)

	if *addStartupFlag {
		_, err := tryAddToStartup()
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

	fmt.Println("TapTo v" + config.Version)

	added, err := tryAddToStartup()
	if err != nil {
		log.Error().Msgf("error adding to startup: %s", err)
		fmt.Println("Error adding to startup:", err)
		os.Exit(1)
	} else if added {
		log.Info().Msg("added to startup")
		fmt.Println("Added TapTo to MiSTeX startup.")
	}

	if !svc.Running() {
		err := svc.Start()
		fmt.Println("TapTo service not running, starting...")
		if err != nil {
			log.Error().Msgf("error starting service: %s", err)
			fmt.Println("Error starting TapTo service:", err)
		} else {
			log.Info().Msg("service started manually")
			fmt.Println("TapTo service started.")
		}
	} else {
		fmt.Println("TapTo service is running.")
	}

	ip, err := utils.GetLocalIp()
	if err != nil {
		fmt.Println("Device address: Unknown")
	} else {
		fmt.Println("Device address:", ip.String())
	}
}
