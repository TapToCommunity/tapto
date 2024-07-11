package utils

import (
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

const LogFile = "tapto.log"

var BaseLogWriters = []io.Writer{&lumberjack.Logger{
	Filename:   LogFile,
	MaxSize:    1,
	MaxBackups: 1,
}}

func InitLogging() error {
	err := os.MkdirAll(filepath.Dir(LogFile), 0755)
	if err != nil {
		return err
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log.Logger = log.Output(io.MultiWriter(BaseLogWriters...))

	return nil
}
