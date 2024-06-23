package mister

import (
	mrextConfig "github.com/wizzomafizzo/mrext/pkg/config"
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

func (p *Platform) KillSoftware() error {
	ExitGame()
	return nil
}

func (p *Platform) IsSoftwareRunning() bool {
	return GetActiveCoreName() != mrextConfig.MenuCore
}

func (p *Platform) GetActiveLauncher() string {
	return GetActiveCoreName()
}

func (p *Platform) PlayFailSound(cfg *config.UserConfig) {
	PlayFail(cfg)
}

func (p *Platform) PlaySuccessSound(cfg *config.UserConfig) {
	PlaySuccess(cfg)
}
