package mister

import (
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/tapto/pkg/config"
)

type Platform struct{}

func (p *Platform) RootFolders(cfg *config.UserConfig) []string {
	return games.GetGamesFolders(UserConfigToMrext(cfg))
}

func (p *Platform) ZipsAsFolders() bool {
	return true
}

func (p *Platform) ConfigFolder() string {
	return ConfigFolder
}

func (p *Platform) NormalizePath(cfg *config.UserConfig, path string) string {
	return NormalizePath(cfg, path)
}
