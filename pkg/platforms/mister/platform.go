package mister

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	mrextConfig "github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

type Platform struct {
	kbd    input.Keyboard
	tr     *Tracker
	stopTr func() error
}

func (p *Platform) Setup(cfg *config.UserConfig) error {
	kbd, err := input.NewKeyboard()
	if err != nil {
		return err
	}

	p.kbd = kbd

	tr, stopTr, err := StartTracker(*UserConfigToMrext(cfg))
	if err != nil {
		return err
	}

	p.tr = tr
	p.stopTr = stopTr

	return nil
}

func (p *Platform) Stop() error {
	if p.stopTr != nil {
		return p.stopTr()
	}

	return nil
}

func (p *Platform) AfterScanHook(token tokens.Token) error {
	f, err := os.Create(TokenReadFile)
	if err != nil {
		return fmt.Errorf("unable to create scan result file %s: %s", TokenReadFile, err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	_, err = f.WriteString(fmt.Sprintf("%s,%s", token.UID, token.Text))
	if err != nil {
		return fmt.Errorf("unable to write scan result file %s: %s", TokenReadFile, err)
	}

	return nil
}

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

func (p *Platform) KillLauncher() error {
	ExitGame()
	return nil
}

func (p *Platform) IsLauncherActive() bool {
	return GetActiveCoreName() != mrextConfig.MenuCore
}

func (p *Platform) IsLauncherDisabled() bool {
	_, err := os.Stat(DisableLaunchFile)
	return err == nil
}

func (p *Platform) SetLauncherEnabled(disabled bool) error {
	if disabled {
		return os.Remove(DisableLaunchFile)
	} else {
		_, err := os.Create(DisableLaunchFile)
		return err
	}
}

func (p *Platform) GetActiveLauncher() string {
	core := GetActiveCoreName()

	if core == mrextConfig.MenuCore {
		return ""
	}

	return core
}

func (p *Platform) PlayFailSound(cfg *config.UserConfig) {
	PlayFail(cfg)
}

func (p *Platform) PlaySuccessSound(cfg *config.UserConfig) {
	PlaySuccess(cfg)
}

func (p *Platform) ActiveSystem() string {
	return p.tr.ActiveSystem
}

func (p *Platform) ActiveGame() string {
	return p.tr.ActiveGameId
}

func (p *Platform) ActiveGameName() string {
	return p.tr.ActiveGameName
}

func (p *Platform) ActiveGamePath() string {
	return p.tr.ActiveGamePath
}

func (p *Platform) SetEventHook(f *func()) {
	p.tr.SetEventHook(f)
}

func (p *Platform) LaunchSystem(cfg *config.UserConfig, id string) error {
	system, err := games.LookupSystem(id)
	if err != nil {
		return err
	}

	return mister.LaunchCore(UserConfigToMrext(cfg), *system)
}

func (p *Platform) LaunchFile(cfg *config.UserConfig, path string) error {
	return mister.LaunchGenericFile(UserConfigToMrext(cfg), path)
}

func (p *Platform) Shell(cmd string) error {
	command := exec.Command("bash", "-c", cmd)
	err := command.Start()
	if err != nil {
		return err
	}
	return nil
}

func (p *Platform) KeyboardInput(input string) error {
	code, err := strconv.Atoi(input)
	if err != nil {
		return err
	}

	p.kbd.Press(code)

	return nil
}

func (p *Platform) ForwardCmd(env platforms.CmdEnv) error {
	if f, ok := commandsMappings[env.Cmd]; ok {
		return f(*p, env)
	} else {
		return fmt.Errorf("command not supported on mister: %s", env.Cmd)
	}
}
