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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/mrext/pkg/input"

	"github.com/wizzomafizzo/mrext/pkg/games"
	mrextMister "github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
)

// TODO: adding some logging for each command
// TODO: search game file
// TODO: game file by hash

var commandMappings = map[string]func(*cmdEnv) error{
	"launch.system": cmdSystem,
	"launch.random": cmdRandom,

	"shell": cmdShell,
	"delay": cmdDelay,

	"mister.ini":  cmdIni,
	"mister.core": cmdLaunchCore,

	"http.get":  cmdHttpGet,
	"http.post": cmdHttpPost,

	"input.key":    cmdKey,
	"input.coinp1": cmdCoinP1,
	"input.coinp2": cmdCoinP2,

	"key":     cmdKey,     // DEPRECATED
	"coinp1":  cmdCoinP1,  // DEPRECATED
	"coinp2":  cmdCoinP2,  // DEPRECATED
	"random":  cmdRandom,  // DEPRECATED
	"command": cmdShell,   // DEPRECATED
	"ini":     cmdIni,     // DEPRECATED
	"system":  cmdSystem,  // DEPRECATED
	"get":     cmdHttpGet, // DEPRECATED
}

type cmdEnv struct {
	args          string
	cfg           *config.UserConfig
	manual        bool
	kbd           input.Keyboard
	text          string
	totalCommands int
	currentIndex  int
}

func cmdSystem(env *cmdEnv) error {
	if strings.EqualFold(env.args, "menu") {
		return mrextMister.LaunchMenu()
	}

	system, err := games.LookupSystem(env.args)
	if err != nil {
		return err
	}

	return mrextMister.LaunchCore(mister.UserConfigToMrext(env.cfg), *system)
}

func cmdShell(env *cmdEnv) error {
	if !env.manual {
		return fmt.Errorf("commands must be manually run")
	}

	command := exec.Command("bash", "-c", env.args)
	err := command.Start()
	if err != nil {
		return err
	}

	return nil
}

func cmdRandom(env *cmdEnv) error {
	if env.args == "" {
		return fmt.Errorf("no system specified")
	}

	if env.args == "all" {
		return mrextMister.LaunchRandomGame(mister.UserConfigToMrext(env.cfg), games.AllSystems())
	}

	systemIds := strings.Split(env.args, ",")
	systems := make([]games.System, 0, len(systemIds))

	for _, id := range systemIds {
		system, err := games.LookupSystem(id)
		if err != nil {
			log.Error().Err(err).Msgf("error looking up system: %s", id)
			continue
		}

		systems = append(systems, *system)
	}

	return mrextMister.LaunchRandomGame(
		mister.UserConfigToMrext(env.cfg),
		systems,
	)
}

func cmdIni(env *cmdEnv) error {
	inis, err := mrextMister.GetAllMisterIni()
	if err != nil {
		return err
	}

	if len(inis) == 0 {
		return fmt.Errorf("no ini files found")
	}

	id, err := strconv.Atoi(env.args)
	if err != nil {
		return err
	}

	if id < 1 || id > len(inis) {
		return fmt.Errorf("ini id out of range: %d", id)
	}

	doRelaunch := true
	// only relaunch if there aren't any more commands
	if env.totalCommands > 1 && env.currentIndex < env.totalCommands-1 {
		doRelaunch = false
	}

	err = mrextMister.SetActiveIni(id, doRelaunch)
	if err != nil {
		return err
	}

	return nil
}

func cmdHttpGet(env *cmdEnv) error {
	go func() {
		resp, err := http.Get(env.args)
		if err != nil {
			log.Error().Msgf("error getting url: %s", err)
			return
		}
		resp.Body.Close()
	}()

	return nil
}

func cmdHttpPost(env *cmdEnv) error {
	parts := strings.SplitN(env.args, ",", 3)
	if len(parts) < 3 {
		return fmt.Errorf("invalid post format: %s", env.args)
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

func cmdKey(env *cmdEnv) error {
	code, err := strconv.Atoi(env.args)
	if err != nil {
		return err
	}

	env.kbd.Press(code)

	return nil
}

func insertCoin(env *cmdEnv, key int) error {
	amount, err := strconv.Atoi(env.args)
	if err != nil {
		return err
	}

	for i := 0; i < amount; i++ {
		env.kbd.Press(key)
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func cmdCoinP1(env *cmdEnv) error {
	return insertCoin(env, 6)
}

func cmdCoinP2(env *cmdEnv) error {
	return insertCoin(env, 7)
}

func cmdDelay(env *cmdEnv) error {
	amount, err := strconv.Atoi(env.args)
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(amount) * time.Millisecond)

	return nil
}

func cmdLaunchCore(env *cmdEnv) error {
	return mrextMister.LaunchShortCore(env.args)
}

// Check all games folders for a relative path to a file
func findFile(cfg *config.UserConfig, path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	ps := filepath.SplitList(path)
	statPath := path

	// if the file is inside a zip or virtual list, we just check that file exists
	// TODO: both of these things are very specific to mister, it would be good to
	//       have a more generic way of handling this for other platforms, or
	//       implement them from tapto(?)
	for i, p := range ps {
		ext := filepath.Ext(strings.ToLower(p))
		if ext == ".zip" || ext == ".txt" {
			statPath = filepath.Join(ps[:i]...)
			break
		}
	}

	// TODO: make sure these games folders are in the correct order
	for _, gf := range games.GetGamesFolders(mister.UserConfigToMrext(cfg)) {
		fullPath := filepath.Join(gf, statPath)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	return path, fmt.Errorf("file not found: %s (%s)", path, statPath)
}

func cmdLaunch(env *cmdEnv) error {
	// if it's an absolute path, just try launch it
	if filepath.IsAbs(env.args) {
		log.Debug().Msgf("launching absolute path: %s", env.args)
		return mrextMister.LaunchGenericFile(mister.UserConfigToMrext(env.cfg), env.args)
	}

	// for relative paths, perform a basic check if the file exists in a games folder
	// this always takes precedence over the system/path format (but is not totally cross platform)
	if p, err := findFile(env.cfg, env.args); err == nil {
		log.Debug().Msgf("launching found relative path: %s", p)
		return mrextMister.LaunchGenericFile(mister.UserConfigToMrext(env.cfg), p)
	} else {
		log.Debug().Err(err).Msgf("error finding file: %s", env.args)
	}

	// attempt to parse the <system>/<path> format
	ps := strings.SplitN(env.text, "/", 2)
	if len(ps) < 2 {
		return fmt.Errorf("invalid launch format: %s", env.text)
	}

	systemId, path := ps[0], ps[1]

	system, err := games.LookupSystem(systemId)
	if err != nil {
		return err
	}

	log.Info().Msgf("launching system: %s, path: %s", systemId, path)

	for _, f := range system.Folder {
		systemPath := filepath.Join(f, path)
		if fp, err := findFile(env.cfg, systemPath); err == nil {
			log.Debug().Msgf("launching found system path: %s", fp)
			return mrextMister.LaunchGenericFile(mister.UserConfigToMrext(env.cfg), fp)
		} else {
			log.Debug().Err(err).Msgf("error finding system file: %s", path)
		}
	}

	return fmt.Errorf("file not found: %s", env.args)
}

func LaunchToken(
	cfg *config.UserConfig,
	manual bool,
	kbd input.Keyboard,
	text string,
	totalCommands int,
	currentIndex int,
) error {
	// explicit commands must begin with **
	if strings.HasPrefix(text, "**") {
		text = strings.TrimPrefix(text, "**")
		ps := strings.SplitN(text, ":", 2)
		if len(ps) < 2 {
			return fmt.Errorf("invalid command: %s", text)
		}

		cmd, args := strings.ToLower(strings.TrimSpace(ps[0])), strings.TrimSpace(ps[1])

		env := &cmdEnv{
			args:          args,
			cfg:           cfg,
			manual:        manual,
			kbd:           kbd,
			text:          text,
			totalCommands: totalCommands,
			currentIndex:  currentIndex,
		}

		if f, ok := commandMappings[cmd]; ok {
			return f(env)
		} else {
			return fmt.Errorf("unknown command: %s", cmd)
		}
	}

	// if it's not a command, treat it as a generic launch command
	return cmdLaunch(&cmdEnv{
		args:          text,
		cfg:           cfg,
		manual:        manual,
		kbd:           kbd,
		text:          text,
		totalCommands: totalCommands,
		currentIndex:  currentIndex,
	})
}
