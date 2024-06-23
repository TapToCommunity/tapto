package platforms

import "github.com/wizzomafizzo/tapto/pkg/config"

type Platform interface {
	RootFolders(*config.UserConfig) []string
	ZipsAsFolders() bool
	ConfigFolder() string
	NormalizePath(*config.UserConfig, string) string
	KillSoftware() error
	IsSoftwareRunning() bool
	GetActiveLauncher() string
	PlayFailSound(*config.UserConfig)
	PlaySuccessSound(*config.UserConfig)
}
