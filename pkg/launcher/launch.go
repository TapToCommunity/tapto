package launcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/wizzomafizzo/tapto/pkg/database/gamesdb"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/utils"
)

func cmdSystem(pl platforms.Platform, env platforms.CmdEnv) error {
	if strings.EqualFold(env.Args, "menu") {
		return pl.KillLauncher()
	}

	return pl.LaunchSystem(env.Cfg, env.Args)
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

	// absolute path, try read dir and pick random file
	// TODO: won't work for zips, switch to using gamesdb when it indexes paths
	// TODO: doesn't filter on extensions
	if filepath.IsAbs(env.Args) {
		if _, err := os.Stat(env.Args); err != nil {
			return err
		}

		files, err := filepath.Glob(filepath.Join(env.Args, "*"))
		if err != nil {
			return err
		}

		if len(files) == 0 {
			return fmt.Errorf("no files found in: %s", env.Args)
		}

		file, err := utils.RandomElem(files)
		if err != nil {
			return err
		}

		return pl.LaunchFile(env.Cfg, file)
	}

	// perform a search similar to launch.search and pick randomly
	// looking for <system>/<query> format
	ps := strings.SplitN(env.Args, "/", 2)
	if len(ps) == 2 {
		systemId, query := ps[0], ps[1]

		system, err := gamesdb.LookupSystem(systemId)
		if err != nil {
			return err
		} else if system == nil {
			return fmt.Errorf("system not found: %s", systemId)
		}

		query = strings.ToLower(query)

		res, err := gamesdb.SearchNamesGlob(pl, []gamesdb.System{*system}, query)
		if err != nil {
			return err
		}

		if len(res) == 0 {
			return fmt.Errorf("no results found for: %s", query)
		}

		game, err := utils.RandomElem(res)
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

	launcher, ok := pl.Launchers()[system.Id]
	if !ok {
		return fmt.Errorf("system not supported: %s", system.Id)
	}

	for _, f := range launcher.Folders {
		systemPath := filepath.Join(f, path)
		if fp, err := findFile(pl, env.Cfg, systemPath); err == nil {
			log.Debug().Msgf("launching found system path: %s", fp)
			return pl.LaunchFile(env.Cfg, fp)
		} else {
			log.Debug().Err(err).Msgf("error finding system file: %s", path)
		}
	}

	// search if the path contains no / or file extensions
	if !strings.Contains(path, "/") && filepath.Ext(path) == "" {
		return cmdSearch(pl, env)
	}

	return fmt.Errorf("file not found: %s", env.Args)
}

func cmdSearch(pl platforms.Platform, env platforms.CmdEnv) error {
	if env.Args == "" {
		return fmt.Errorf("no query specified")
	}

	query := strings.ToLower(env.Args)
	query = strings.TrimSpace(query)

	if !strings.Contains(env.Args, "/") {
		// search all systems
		res, err := gamesdb.SearchNamesGlob(pl, gamesdb.AllSystems(), query)
		if err != nil {
			return err
		}

		if len(res) == 0 {
			return fmt.Errorf("no results found for: %s", query)
		}

		return pl.LaunchFile(env.Cfg, res[0].Path)
	}

	ps := strings.SplitN(query, "/", 2)
	if len(ps) < 2 {
		return fmt.Errorf("invalid search format: %s", query)
	}

	systemId, query := ps[0], ps[1]

	if query == "" {
		return fmt.Errorf("no query specified")
	}

	systems := make([]gamesdb.System, 0)

	if systemId == "all" {
		systems = gamesdb.AllSystems()
	} else {
		system, err := gamesdb.LookupSystem(systemId)
		if err != nil {
			return err
		}

		systems = append(systems, *system)
	}

	res, err := gamesdb.SearchNamesGlob(pl, systems, query)
	if err != nil {
		return err
	}

	if len(res) == 0 {
		return fmt.Errorf("no results found for: %s", query)
	}

	return pl.LaunchFile(env.Cfg, res[0].Path)
}
