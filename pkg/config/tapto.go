package config

import (
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
)

const (
	Version            = "2.0.1"
	GamesDbFilename    = "games.db"
	TapToDbFilename    = "tapto.db"
	DefaultApiPort     = "7497"
	LogFilename        = "tapto.log"
	AppName            = "tapto"
	UserConfigFilename = "tapto.ini"
	PidFilename        = "tapto.pid"
)

func TempDir() string {
	path := filepath.Join(os.TempDir(), AppName)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		log.Error().Err(err).Msg("error creating temp folder")
	}
	return path
}
