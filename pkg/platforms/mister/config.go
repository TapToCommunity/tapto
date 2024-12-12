//go:build linux || darwin

package mister

import (
	"github.com/ZaparooProject/zaparoo-core/pkg/config"
	mrextConfig "github.com/wizzomafizzo/mrext/pkg/config"
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
	ConfigFolder      = mrextConfig.ScriptsConfigFolder + "/tapto"
	DbFile            = ConfigFolder + "/tapto.db"
	GamesDbFile       = ConfigFolder + "/games.db"
	ArcadeDbUrl       = "https://api.github.com/repositories/521644036/contents/ArcadeDatabase_CSV"
	ArcadeDbFile      = ConfigFolder + "/ArcadeDatabase.csv"
	ScriptsFolder     = mrextConfig.ScriptsFolder
	CmdInterface      = "/dev/MiSTer_cmd"
	LinuxFolder       = "/media/fat/linux"
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
