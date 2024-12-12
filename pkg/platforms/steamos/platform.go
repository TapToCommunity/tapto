/*
Zaparoo Core
Copyright (C) 2024 Callan Barrett

This file is part of Zaparoo Core.

Zaparoo Core is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Zaparoo Core is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Zaparoo Core.  If not, see <http://www.gnu.org/licenses/>.
*/

package steamos

import (
	"errors"
	"github.com/ZaparooProject/zaparoo-core/pkg/readers/libnfc"
	"github.com/ZaparooProject/zaparoo-core/pkg/service/tokens"
	"github.com/ZaparooProject/zaparoo-core/pkg/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ZaparooProject/zaparoo-core/pkg/api/models"

	"github.com/ZaparooProject/zaparoo-core/pkg/config"
	"github.com/ZaparooProject/zaparoo-core/pkg/database/gamesdb"
	"github.com/ZaparooProject/zaparoo-core/pkg/platforms"
	"github.com/ZaparooProject/zaparoo-core/pkg/readers"
	"github.com/ZaparooProject/zaparoo-core/pkg/readers/file"
	"github.com/ZaparooProject/zaparoo-core/pkg/readers/simple_serial"
	"github.com/andygrunwald/vdf"
	"github.com/rs/zerolog/log"
)

type Platform struct {
}

func (p *Platform) Id() string {
	return "steamos"
}

func (p *Platform) SupportedReaders(cfg *config.UserConfig) []readers.Reader {
	return []readers.Reader{
		file.NewReader(cfg),
		simple_serial.NewReader(cfg),
		libnfc.NewReader(cfg),
	}
}

func (p *Platform) Setup(_ *config.UserConfig, _ chan<- models.Notification) error {
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

func (p *Platform) ConfigFolder() string {
	// this could be AppData instead
	return filepath.Join(utils.ExeDir(), "data")
}

func (p *Platform) LogFolder() string {
	return filepath.Join(utils.ExeDir(), "logs")
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

func (p *Platform) LaunchSystem(cfg *config.UserConfig, id string) error {
	log.Info().Msgf("launching system: %s", id)
	return nil
}

func (p *Platform) LaunchFile(cfg *config.UserConfig, path string) error {
	log.Info().Msgf("launching file: %s", path)

	launchers := utils.PathToLaunchers(cfg, p, path)

	if len(launchers) == 0 {
		return errors.New("no launcher found")
	}

	return launchers[0].Launch(cfg, path)
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
				systemId string,
				results []platforms.ScanResult,
			) ([]platforms.ScanResult, error) {
				// TODO: check for external folders
				root := "/home/deck/.steam/steam/steamapps"

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
				return exec.Command("steam", "steam://rungameid/"+id).Start()
			},
		},
		{
			Id:            "Generic",
			Extensions:    []string{".sh"},
			AllowListOnly: true,
			Launch: func(cfg *config.UserConfig, path string) error {
				return exec.Command("bash", "-c", path).Start()
			},
		},
	}
}
