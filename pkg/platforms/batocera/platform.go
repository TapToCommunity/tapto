package batocera

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/database/gamesdb"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/readers"
	"github.com/wizzomafizzo/tapto/pkg/readers/file"
	"github.com/wizzomafizzo/tapto/pkg/readers/libnfc"
	"github.com/wizzomafizzo/tapto/pkg/readers/simple_serial"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

type Platform struct {
}

func (p *Platform) Id() string {
	return "batocera"
}

func (p *Platform) SupportedReaders(cfg *config.UserConfig) []readers.Reader {
	return []readers.Reader{
		libnfc.NewReader(cfg),
		file.NewReader(cfg),
		simple_serial.NewReader(cfg),
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
		"/userdata/roms",
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

	relPath := path
	for _, rf := range p.RootFolders(cfg) {
		if strings.HasPrefix(relPath, rf+"/") {
			relPath = strings.TrimPrefix(relPath, rf+"/")
			break
		}
	}
	log.Info().Msgf("relative path: %s", relPath)

	root := strings.Split(relPath, "/")[0]
	log.Info().Msgf("root: %s", root)

	systemId := ""
	for _, launcher := range p.Launchers() {
		for _, folder := range launcher.Folders {
			if folder == root {
				systemId = launcher.SystemId
				break
			}
		}
	}

	if systemId == "" {
		log.Error().Msgf("system not found for path: %s", path)
	}

	for _, launcher := range p.Launchers() {
		if launcher.SystemId == systemId {
			return launcher.Launch(cfg, path)
		}
	}

	return errors.New("launcher not found")
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

func (p *Platform) Launchers() []platforms.Launcher {
	return []platforms.Launcher{
		{
			SystemId:   gamesdb.SystemGenesis,
			Folders:    []string{"megadrive"},
			Extensions: []string{".bin", ".gen", ".md", ".sg", ".smd", ".zip", ".7z"},
			Launch: func(cfg *config.UserConfig, path string) error {
				cmd := exec.Command("emulatorlauncher", "-system", "megadrive", "-rom", path)
				cmd.Env = os.Environ()
				cmd.Env = append(cmd.Env, "DISPLAY=:0.0")
				return cmd.Start()
			},
		},
	}
}
