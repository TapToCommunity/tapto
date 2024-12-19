//go:build linux || darwin

package mister

import (
	"github.com/ZaparooProject/zaparoo-core/pkg/config"
	mrextConfig "github.com/wizzomafizzo/mrext/pkg/config"
)

const (
	TempFolder        = "/tmp/zaparoo"
	DisableLaunchFile = TempFolder + "/zaparoo.disabled"
	SuccessSoundFile  = TempFolder + "/success.wav"
	FailSoundFile     = TempFolder + "/fail.wav"
	SocketFile        = TempFolder + "/zaparoo.sock"
	MappingsFile      = "/media/fat/nfc.csv"
	TokenReadFile     = "/tmp/TOKENREAD"
	ConfigFolder      = mrextConfig.ScriptsConfigFolder + "/zaparoo"
	ArcadeDbUrl       = "https://api.github.com/repositories/521644036/contents/ArcadeDatabase_CSV"
	ArcadeDbFile      = ConfigFolder + "/ArcadeDatabase.csv"
	ScriptsFolder     = mrextConfig.ScriptsFolder
	CmdInterface      = "/dev/MiSTer_cmd"
	LinuxFolder       = "/media/fat/linux"
)

func UserConfigToMrext(cfg *config.Instance) *mrextConfig.UserConfig {
	var setCore []string
	for _, v := range cfg.SystemDefaults() {
		setCore = append(setCore, v.System+":"+v.Launcher)
	}
	return &mrextConfig.UserConfig{
		Systems: mrextConfig.SystemsConfig{
			GamesFolder: cfg.IndexRoots(),
			SetCore:     setCore,
		},
	}
}
