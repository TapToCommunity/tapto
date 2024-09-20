//go:build linux && cgo

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
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mistex"
	"github.com/wizzomafizzo/tapto/pkg/service"
	"github.com/wizzomafizzo/tapto/pkg/service/api"
	"github.com/wizzomafizzo/tapto/pkg/utils"

	"github.com/wizzomafizzo/tapto/pkg/config"
)

const appName = "tapto"

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
		"http://127.0.0.1:"+cfg.Api.Port+"/api/v1/readers/0/write",
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

func handleLaunchCommand(tokenToLaunch string, svc *mister.Service, cfg *config.UserConfig) {
	log.Info().Msgf("launching token: %s", tokenToLaunch)

	if !svc.Running() {
		_, _ = fmt.Fprintln(os.Stderr, "TapTo service is not running, please start it before launching.")
		log.Error().Msg("TapTo service is not running, exiting")
		os.Exit(1)
	}

	resp, err := http.Get("http://127.0.0.1:" + cfg.Api.Port + "/api/v1/launch/" + url.QueryEscape(tokenToLaunch))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error sending request:", err)
		log.Error().Msgf("error sending request: %s", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	_, _ = fmt.Fprintln(os.Stderr, "Successfully launched token.")
	log.Info().Msg("successfully launched token")
	os.Exit(0)
}

func main() {
	svcOpt := flag.String("service", "", "manage TapTo service (start, stop, restart, status)")
	writeOpt := flag.String("write", "", "write text to tag")
	launchOpt := flag.String("launch", "", "execute text as if it were a token")
	versionOpt := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versionOpt {
		fmt.Println("TapTo v" + config.Version + " (mistex)")
		os.Exit(0)
	}

	pl := &mistex.Platform{}
	err := utils.InitLogging(pl)
	if err != nil {
		fmt.Println("Error initializing logging:", err)
		os.Exit(1)
	}

	cfg, err := config.NewUserConfig(appName, &config.UserConfig{
		TapTo: config.TapToConfig{
			ProbeDevice: true,
		},
		Api: config.ApiConfig{
			Port: config.DefaultApiPort,
		},
	})
	if err != nil {
		log.Error().Msgf("error loading user config: %s", err)
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	if cfg.GetDebug() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	svc, err := mister.NewService(mister.ServiceArgs{
		Name: appName,
		Entry: func() (func() error, error) {
			return service.Start(pl, cfg)
		},
	})
	if err != nil {
		log.Error().Msgf("error creating service: %s", err)
		fmt.Println("Error creating service:", err)
		os.Exit(1)
	}

	if *writeOpt != "" {
		handleWriteCommand(*writeOpt, svc, cfg)
	} else if *launchOpt != "" {
		handleLaunchCommand(*launchOpt, svc, cfg)
	}

	svc.ServiceHandler(svcOpt)

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
