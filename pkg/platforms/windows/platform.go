//go:build windows

package windows

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/readers"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

type Platform struct {
}

func (p *Platform) Id() string {
	return "windows"
}

func (p *Platform) Setup(cfg *config.UserConfig) error {
	return nil
}

func (p *Platform) Stop() error {
	return nil
}

func (p *Platform) AfterScanHook(token tokens.Token) error {
	return nil
}

func (p *Platform) ReadersUpdateHook(readers map[string]*readers.Reader) error {
	return nil
}

func (p *Platform) RootFolders(cfg *config.UserConfig) []string {
	return []string{}
}

func (p *Platform) ZipsAsFolders() bool {
	return false
}

func exeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}

	return filepath.Dir(exe)
}

func (p *Platform) ConfigFolder() string {
	// this could be AppData instead
	return filepath.Join(exeDir(), "data")
}

func (p *Platform) LogFolder() string {
	return filepath.Join(exeDir(), "logs")
}

func (p *Platform) NormalizePath(cfg *config.UserConfig, path string) string {
	return path
}

func LaunchMenu() error {
	return nil
}

func (p *Platform) KillLauncher() error {
	return nil
}

func (p *Platform) LaunchingEnabled() bool {
	return true
}

func (p *Platform) SetLaunching(disabled bool) error {
	return nil
}

func (p *Platform) GetActiveLauncher() string {
	return ""
}

func (p *Platform) PlayFailSound(cfg *config.UserConfig) {
}

func (p *Platform) PlaySuccessSound(cfg *config.UserConfig) {
}

func (p *Platform) ActiveSystem() string {
	return ""
}

func (p *Platform) ActiveGame() string {
	return ""
}

func (p *Platform) ActiveGameName() string {
	return ""
}

func (p *Platform) ActiveGamePath() string {
	return ""
}

func (p *Platform) SetEventHook(f *func()) {
}

func (p *Platform) LaunchSystem(cfg *config.UserConfig, id string) error {
	log.Info().Msgf("launching system: %s", id)
	return nil
}

func (p *Platform) LaunchFile(cfg *config.UserConfig, path string) error {
	log.Info().Msgf("launching file: %s", path)
	return nil
}

func (p *Platform) Shell(cmd string) error {
	return nil
}

func (p *Platform) KeyboardInput(input string) error {
	return nil
}

func (p *Platform) KeyboardPress(name string) error {
	return nil
}

func (p *Platform) GamepadPress(name string) error {
	return nil
}

func (p *Platform) ForwardCmd(env platforms.CmdEnv) error {
	return nil
}

func (p *Platform) LookupMapping(_ tokens.Token) (string, bool) {
	return "", false
}
