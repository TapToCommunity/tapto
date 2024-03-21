package mister

import (
	mrextConfig "github.com/wizzomafizzo/mrext/pkg/config"
	config "github.com/wizzomafizzo/tapto/pkg/config"
)

const (
	TempFolder        = "/tmp/tapto"
	LogFile           = TempFolder + "/tapto.log"
	DisableLaunchFile = TempFolder + "/tapto.disabled"
	SuccessSoundFile  = TempFolder + "/success.wav"
	FailSoundFile     = TempFolder + "/fail.wav"
	SocketFile        = TempFolder + "/tapto.sock"
	PidFile           = TempFolder + "/tapto.pid"
	MappingsFile      = "/media/fat/nfc.csv"
	TokenReadFile     = "/tmp/TOKENREAD"
	DbFile            = mrextConfig.ScriptsConfigFolder + "/tapto/tapto.db"
)

func UserConfigToMrext(cfg *config.UserConfig) *mrextConfig.UserConfig {
	return &mrextConfig.UserConfig{
		AppPath: cfg.AppPath,
		IniPath: cfg.IniPath,
		Nfc: mrextConfig.NfcConfig{
			ConnectionString: cfg.TapTo.ConnectionString,
			AllowCommands:    cfg.TapTo.AllowCommands,
			DisableSounds:    cfg.TapTo.DisableSounds,
			ProbeDevice:      cfg.TapTo.ProbeDevice,
		},
		Systems: mrextConfig.SystemsConfig{
			GamesFolder: cfg.Systems.GamesFolder,
			SetCore:     cfg.Systems.SetCore,
		},
	}
}
