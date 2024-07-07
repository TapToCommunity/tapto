package launcher

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/wizzomafizzo/tapto/pkg/database/gamesdb"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
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
