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

func InitLogging(cfg *config.Instance, pl platforms.Platform) error {
	logFile := filepath.Join(pl.LogDir(), config.LogFile)

	err := os.MkdirAll(filepath.Dir(logFile), 0755)
	if err != nil {
		return err
	}

	var BaseLogWriters = []io.Writer{&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    1,
		MaxBackups: 2,
	}}

	// TODO: need some way to enable console logging per platform

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log.Logger = log.Output(io.MultiWriter(BaseLogWriters...))

	if cfg.DebugLogging() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return nil
}
