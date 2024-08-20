package windows

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/andygrunwald/vdf"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/database/gamesdb"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/readers"
	"github.com/wizzomafizzo/tapto/pkg/readers/acr122_pcsc"
	"github.com/wizzomafizzo/tapto/pkg/readers/file"
	"github.com/wizzomafizzo/tapto/pkg/readers/pn532_uart"
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
		acr122_pcsc.NewAcr122Pcsc(cfg),
		pn532_uart.NewReader(cfg),
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

	launchers := make([]platforms.Launcher, 0)
	lp := strings.ToLower(path)

	// TODO: move to matchsystemfile
	for _, l := range p.Launchers() {
		match := false

		// check for global extensions
		for _, ext := range l.Extensions {
			if filepath.Ext(lp) == ext && l.Folders == nil {
				launchers = append(launchers, l)
				match = true
				break
			}
		}
		if match {
			continue
		}

		// check for scheme
		for _, scheme := range l.Schemes {
			if strings.HasPrefix(lp, scheme+"://") {
				launchers = append(launchers, l)
				break
			}
		}
	}

	if len(launchers) == 0 {
		return errors.New("no launcher found for file")
	}

	l := launchers[0]

	if l.Launch != nil {
		if l.AllowListOnly {
			if cfg.IsFileAllowed(path) {
				return l.Launch(cfg, path)
			} else {
				return errors.New("file not in allow list: " + path)
			}
		}

		return l.Launch(cfg, path)
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

func (p *Platform) Launchers() []platforms.Launcher {
	return []platforms.Launcher{
		{
			Id:       "Steam",
			SystemId: gamesdb.SystemPC,
			Schemes:  []string{"steam"},
			Scanner: func(
				cfg *config.UserConfig,
				results []platforms.ScanResult,
			) ([]platforms.ScanResult, error) {
				// TODO: detect this path from registry
				root := "C:\\Program Files (x86)\\Steam\\steamapps"

				f, err := os.Open(filepath.Join(root, "libraryfolders.vdf"))
				if err != nil {
					log.Error().Err(err).Msg("error opening libraryfolders.vdf")
					return results, nil
				}

				p := vdf.NewParser(f)
				m, err := p.Parse()
				if err != nil {
					log.Error().Err(err).Msg("error parsing libraryfolders.vdf")
					return results, nil
				}

				lfs := m["libraryfolders"].(map[string]interface{})
				for l, v := range lfs {
					log.Debug().Msgf("library id: %s", l)
					ls := v.(map[string]interface{})

					libraryPath := ls["path"].(string)
					apps := ls["apps"].(map[string]interface{})

					for id := range apps {
						log.Debug().Msgf("app id: %s", id)

						manifestPath := filepath.Join(libraryPath, "steamapps", "appmanifest_"+id+".acf")
						af, err := os.Open(manifestPath)
						if err != nil {
							log.Error().Err(err).Msgf("error opening manifest: %s", manifestPath)
							return results, nil
						}

						ap := vdf.NewParser(af)
						am, err := ap.Parse()
						if err != nil {
							log.Error().Err(err).Msgf("error parsing manifest: %s", manifestPath)
							return results, nil
						}

						appState := am["AppState"].(map[string]interface{})
						log.Debug().Msgf("app name: %v", appState["name"])

						results = append(results, platforms.ScanResult{
							Path: "steam://" + id,
							Name: appState["name"].(string),
						})
					}
				}

				return results, nil
			},
			Launch: func(cfg *config.UserConfig, path string) error {
				id := strings.TrimPrefix(path, "steam://")
				id = strings.TrimPrefix(id, "rungameid/")
				return exec.Command(
					"cmd", "/c",
					"start",
					"steam://rungameid/"+id,
				).Start()
			},
		},
		{
			Id:       "Flashpoint",
			SystemId: gamesdb.SystemPC,
			Schemes:  []string{"flashpoint"},
			Launch: func(cfg *config.UserConfig, path string) error {
				id := strings.TrimPrefix(path, "flashpoint://")
				id = strings.TrimPrefix(id, "run/")
				return exec.Command(
					"cmd", "/c",
					"start",
					"flashpoint://run/"+id,
				).Start()
			},
		},
		{
			Id:            "Generic",
			Extensions:    []string{".exe", ".bat", ".cmd", ".lnk"},
			AllowListOnly: true,
			Launch: func(cfg *config.UserConfig, path string) error {
				return exec.Command("cmd", "/c", path).Start()
			},
		},
	}
}
