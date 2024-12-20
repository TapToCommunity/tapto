package utils

import (
	"io"
	"os"
	"path/filepath"

	"github.com/ZaparooProject/zaparoo-core/pkg/config"
	"github.com/ZaparooProject/zaparoo-core/pkg/platforms"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogging(cfg *config.UserConfig, pl platforms.Platform) error {
	logFile := filepath.Join(pl.LogFolder(), config.LogFilename)

	err := os.MkdirAll(filepath.Dir(logFile), 0755)
	if err != nil {
		return err
	}

	var BaseLogWriters = []io.Writer{&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    1,
		MaxBackups: 2,
	}}

	if cfg.TapTo.ConsoleLogging {
		// BaseLogWriters = append(BaseLogWriters, zerolog.ConsoleWriter{Out: os.Stderr})
		BaseLogWriters = append(BaseLogWriters, os.Stderr)
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log.Logger = log.Output(io.MultiWriter(BaseLogWriters...))

	if cfg.GetDebug() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return nil
}
