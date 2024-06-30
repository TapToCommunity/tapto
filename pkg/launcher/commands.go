/*
TapTo
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

package launcher

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/rs/zerolog/log"

	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/database/gamesdb"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
)

// TODO: adding some logging for each command
// TODO: search game file
// TODO: game file by hash

var commandMappings = map[string]func(platforms.Platform, platforms.CmdEnv) error{
	"launch.system": cmdSystem,
	"launch.random": cmdRandom,

	"shell": cmdShell,
	"delay": cmdDelay,

	"mister.ini":  forwardCmd,
	"mister.core": forwardCmd,
	// "mister.script": forwardCmd,
	"mister.mgl": forwardCmd,

	"http.get":  cmdHttpGet,
	"http.post": cmdHttpPost,

	"input.key":     cmdKey,
	"input.gamepad": cmdGamepad,
	"input.coinp1":  cmdCoinP1,
	"input.coinp2":  cmdCoinP2,

	"key":     cmdKey,     // DEPRECATED
	"coinp1":  cmdCoinP1,  // DEPRECATED
	"coinp2":  cmdCoinP2,  // DEPRECATED
	"random":  cmdRandom,  // DEPRECATED
	"command": cmdShell,   // DEPRECATED
	"ini":     forwardCmd, // DEPRECATED
	"system":  cmdSystem,  // DEPRECATED
	"get":     cmdHttpGet, // DEPRECATED
}

var softwareChangeCommands = []string{"launch.system", "launch.random", "mister.core"}

func cmdSystem(pl platforms.Platform, env platforms.CmdEnv) error {
	if strings.EqualFold(env.Args, "menu") {
		return pl.KillLauncher()
	}

	return pl.LaunchSystem(env.Cfg, env.Args)
}

func cmdShell(pl platforms.Platform, env platforms.CmdEnv) error {
	if !env.Manual {
		return fmt.Errorf("shell commands must be manually run")
	}

	return pl.Shell(env.Args)
}

func cmdRandom(pl platforms.Platform, env platforms.CmdEnv) error {
	if env.Args == "" {
		return fmt.Errorf("no system specified")
	}

	if env.Args == "all" {
		game, err := gamesdb.RandomGame(pl, gamesdb.AllSystems())
		if err != nil {
			return err
		}

		return pl.LaunchFile(env.Cfg, game.Path)
	}

	systemIds := strings.Split(env.Args, ",")
	systems := make([]gamesdb.System, 0, len(systemIds))

	for _, id := range systemIds {
		system, err := gamesdb.LookupSystem(id)
		if err != nil {
			log.Error().Err(err).Msgf("error looking up system: %s", id)
			continue
		}

		systems = append(systems, *system)
	}

	game, err := gamesdb.RandomGame(pl, systems)
	if err != nil {
		return err
	}

	return pl.LaunchFile(env.Cfg, game.Path)
}

func cmdHttpGet(pl platforms.Platform, env platforms.CmdEnv) error {
	go func() {
		resp, err := http.Get(env.Args)
		if err != nil {
			log.Error().Msgf("error getting url: %s", err)
			return
		}
		resp.Body.Close()
	}()

	return nil
}

func cmdHttpPost(pl platforms.Platform, env platforms.CmdEnv) error {
	parts := strings.SplitN(env.Args, ",", 3)
	if len(parts) < 3 {
		return fmt.Errorf("invalid post format: %s", env.Args)
	}

	url, format, data := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2])

	go func() {
		resp, err := http.Post(url, format, strings.NewReader(data))
		if err != nil {
			log.Error().Msgf("error posting to url: %s", err)
			return
		}
		resp.Body.Close()
	}()

	return nil
}

func cmdKey(pl platforms.Platform, env platforms.CmdEnv) error {
	return pl.KeyboardInput(env.Args)
}

func cmdGamepad(pl platforms.Platform, env platforms.CmdEnv) error {
	return pl.GamepadInput(strings.Split(env.Args, ""))
}

func insertCoin(pl platforms.Platform, env platforms.CmdEnv, key string) error {
	amount, err := strconv.Atoi(env.Args)
	if err != nil {
		return err
	}

	for i := 0; i < amount; i++ {
		pl.KeyboardInput(key)
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func cmdCoinP1(pl platforms.Platform, env platforms.CmdEnv) error {
	return insertCoin(pl, env, "6")
}

func cmdCoinP2(pl platforms.Platform, env platforms.CmdEnv) error {
	return insertCoin(pl, env, "7")
}

func cmdDelay(pl platforms.Platform, env platforms.CmdEnv) error {
	amount, err := strconv.Atoi(env.Args)
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(amount) * time.Millisecond)

	return nil
}

func forwardCmd(pl platforms.Platform, env platforms.CmdEnv) error {
	return pl.ForwardCmd(env)
}

// Check all games folders for a relative path to a file
func findFile(pl platforms.Platform, cfg *config.UserConfig, path string) (string, error) {
	// TODO: can do basic file exists check here too
	if filepath.IsAbs(path) {
		return path, nil
	}

	ps := strings.Split(path, string(filepath.Separator))
	statPath := path

	// if the file is inside a zip or virtual list, we just check that file exists
	// TODO: both of these things are very specific to mister, it would be good to
	//       have a more generic way of handling this for other platforms, or
	//       implement them from tapto(?)
	for i, p := range ps {
		ext := filepath.Ext(strings.ToLower(p))
		if ext == ".zip" || ext == ".txt" {
			statPath = filepath.Join(ps[:i+1]...)
			log.Debug().Msgf("found zip/txt, setting stat path: %s", statPath)
			break
		}
	}

	for _, gf := range pl.RootFolders(cfg) {
		fullPath := filepath.Join(gf, statPath)
		if _, err := os.Stat(fullPath); err == nil {
			log.Debug().Msgf("found file: %s", fullPath)
			return filepath.Join(gf, path), nil
		}
	}

	return path, fmt.Errorf("file not found: %s", path)
}

func cmdLaunch(pl platforms.Platform, env platforms.CmdEnv) error {
	// if it's an absolute path, just try launch it
	if filepath.IsAbs(env.Args) {
		log.Debug().Msgf("launching absolute path: %s", env.Args)
		return pl.LaunchFile(env.Cfg, env.Args)
	}

	// for relative paths, perform a basic check if the file exists in a games folder
	// this always takes precedence over the system/path format (but is not totally cross platform)
	if p, err := findFile(pl, env.Cfg, env.Args); err == nil {
		log.Debug().Msgf("launching found relative path: %s", p)
		return pl.LaunchFile(env.Cfg, p)
	} else {
		log.Debug().Err(err).Msgf("error finding file: %s", env.Args)
	}

	// attempt to parse the <system>/<path> format
	ps := strings.SplitN(env.Text, "/", 2)
	if len(ps) < 2 {
		return fmt.Errorf("invalid launch format: %s", env.Text)
	}

	systemId, path := ps[0], ps[1]

	system, err := gamesdb.LookupSystem(systemId)
	if err != nil {
		return err
	}

	log.Info().Msgf("launching system: %s, path: %s", systemId, path)

	for _, f := range system.Folders {
		systemPath := filepath.Join(f, path)
		if fp, err := findFile(pl, env.Cfg, systemPath); err == nil {
			log.Debug().Msgf("launching found system path: %s", fp)
			return pl.LaunchFile(env.Cfg, fp)
		} else {
			log.Debug().Err(err).Msgf("error finding system file: %s", path)
		}
	}

	return fmt.Errorf("file not found: %s", env.Args)
}

/**
 * Will launch a command related to the token, and if it is a software that will
 * change the currently loaded software will also return a boolean set to true
 */
func LaunchToken(
	pl platforms.Platform,
	cfg *config.UserConfig,
	manual bool,
	text string,
	totalCommands int,
	currentIndex int,
) (error, bool) {
	// explicit commands must begin with **
	if strings.HasPrefix(text, "**") {
		text = strings.TrimPrefix(text, "**")
		ps := strings.SplitN(text, ":", 2)
		if len(ps) < 2 {
			return fmt.Errorf("invalid command: %s", text), false
		}

		cmd, args := strings.ToLower(strings.TrimSpace(ps[0])), strings.TrimSpace(ps[1])

		env := platforms.CmdEnv{
			Cmd:           cmd,
			Args:          args,
			Cfg:           cfg,
			Manual:        manual,
			Text:          text,
			TotalCommands: totalCommands,
			CurrentIndex:  currentIndex,
		}

		if f, ok := commandMappings[cmd]; ok {
			return f(pl, env), slices.Contains(softwareChangeCommands, cmd)
		} else {
			return fmt.Errorf("unknown command: %s", cmd), false
		}
	}

	// if it's not a command, treat it as a generic launch command
	return cmdLaunch(pl, platforms.CmdEnv{
		Cmd:           "launch",
		Args:          text,
		Cfg:           cfg,
		Manual:        manual,
		Text:          text,
		TotalCommands: totalCommands,
		CurrentIndex:  currentIndex,
	}), true
}
