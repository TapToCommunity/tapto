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
	GamesDbFile       = mrextConfig.ScriptsConfigFolder + "/tapto/games.db"
	ArcadeDbUrl       = "https://api.github.com/repositories/521644036/contents/ArcadeDatabase_CSV"
	ArcadeDbFile      = mrextConfig.ScriptsConfigFolder + "/tapto/ArcadeDatabase.csv"
)

func UserConfigToMrext(cfg *config.UserConfig) *mrextConfig.UserConfig {
	return &mrextConfig.UserConfig{
		AppPath: cfg.AppPath,
		IniPath: cfg.IniPath,
		Nfc: mrextConfig.NfcConfig{
			ConnectionString: cfg.GetConnectionString(),
			AllowCommands:    cfg.GetAllowCommands(),
			DisableSounds:    cfg.GetDisableSounds(),
			ProbeDevice:      cfg.GetProbeDevice(),
		},
		Systems: mrextConfig.SystemsConfig{
			GamesFolder: cfg.Systems.GamesFolder,
			SetCore:     cfg.Systems.SetCore,
		},
	}
}
