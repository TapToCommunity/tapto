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
	s "strings"
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

	"mister.ini": cmdIni,

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
	if s.EqualFold(env.args, "menu") {
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

	systemIds := s.Split(env.args, ",")
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
	parts := s.SplitN(env.args, "|", 3)
	if len(parts) < 3 {
		return fmt.Errorf("invalid post format: %s", env.args)
	}

	url, format, data := s.TrimSpace(parts[0]), s.TrimSpace(parts[1]), s.TrimSpace(parts[2])

	go func() {
		resp, err := http.Post(url, format, s.NewReader(data))
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

func LaunchToken(
	cfg *config.UserConfig,
	manual bool,
	kbd input.Keyboard,
	text string,
	totalCommands int,
	currentIndex int,
) error {
	// detection can never be perfect, but these characters are illegal in
	// windows filenames and heavily avoided in linux. use them to mark that
	// this is a command
	if s.HasPrefix(text, "**") {
		text = s.TrimPrefix(text, "**")
		parts := s.SplitN(text, ":", 2)
		if len(parts) < 2 {
			return fmt.Errorf("invalid command: %s", text)
		}

		cmd, args := s.ToLower(s.TrimSpace(parts[0])), s.TrimSpace(parts[1])

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

	// if it's not a command, assume it's some kind of file path
	if filepath.IsAbs(text) {
		return mrextMister.LaunchGenericFile(mister.UserConfigToMrext(cfg), text)
	}

	// if it's a relative path with no extension, assume it's a core
	if filepath.Ext(text) == "" {
		return mrextMister.LaunchShortCore(text)
	}

	// if the file is in a .zip, just check .zip exists in each games folder
	parts := s.Split(text, "/")
	for i, part := range parts {
		if s.HasSuffix(s.ToLower(part), ".zip") {
			zipPath := filepath.Join(parts[:i+1]...)
			for _, folder := range games.GetGamesFolders(mister.UserConfigToMrext(cfg)) {
				if _, err := os.Stat(filepath.Join(folder, zipPath)); err == nil {
					return mrextMister.LaunchGenericFile(mister.UserConfigToMrext(cfg), filepath.Join(folder, text))
				}
			}
			break
		}
	}

	// then try check for the whole path in each game folder
	for _, folder := range games.GetGamesFolders(mister.UserConfigToMrext(cfg)) {
		path := filepath.Join(folder, text)
		if _, err := os.Stat(path); err == nil {
			return mrextMister.LaunchGenericFile(mister.UserConfigToMrext(cfg), path)
		}
	}

	return fmt.Errorf("could not find file: %s", text)
}
