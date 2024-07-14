//go:build windows

package windows

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/readers"
	acr122pcsc "github.com/wizzomafizzo/tapto/pkg/readers/acr122_pcsc"
	"github.com/wizzomafizzo/tapto/pkg/readers/file"
	"github.com/wizzomafizzo/tapto/pkg/readers/simple_serial"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

type Platform struct {
}

func (p *Platform) Id() string {
	return "windows"
}

func (p *Platform) SupportedReaders(cfg *config.UserConfig) []readers.Reader {
	return []readers.Reader{
		file.NewReader(cfg),
		simple_serial.NewReader(cfg),
		acr122pcsc.NewAcr122Pcsc(cfg),
	}
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
	return []string{
		"C:\\scratch",
	}
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

	if filepath.Ext(path) == ".txt" {
		// get filename minus ext

		fn := filepath.Base(path)
		fn = fn[:len(fn)-4]

		return exec.Command("cmd", "/c", "C:\\Program Files (x86)\\Steam\\steam.exe", "steam://rungameid/"+fn).Start()
	}

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
