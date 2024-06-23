package platforms

import "github.com/wizzomafizzo/tapto/pkg/config"

type Platform interface {
	RootFolders(*config.UserConfig) []string
	ZipsAsFolders() bool
	ConfigFolder() string
	NormalizePath(*config.UserConfig, string) string
}
