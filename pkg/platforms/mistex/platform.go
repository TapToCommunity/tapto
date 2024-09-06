//go:build linux || darwin

package mistex

import (
	"fmt"
	"github.com/wizzomafizzo/tapto/pkg/service/notifications"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/bendahl/uinput"
	mrextConfig "github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/input"
	mm "github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
	"github.com/wizzomafizzo/tapto/pkg/readers"
	"github.com/wizzomafizzo/tapto/pkg/readers/file"
	"github.com/wizzomafizzo/tapto/pkg/readers/libnfc"
	"github.com/wizzomafizzo/tapto/pkg/readers/simple_serial"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

type Platform struct {
	kbd    input.Keyboard
	gpd    uinput.Gamepad
	tr     *mister.Tracker
	stopTr func() error
}

func (p *Platform) Id() string {
	return "mistex"
}

func (p *Platform) SupportedReaders(cfg *config.UserConfig) []readers.Reader {
	return []readers.Reader{
		libnfc.NewReader(cfg),
		file.NewReader(cfg),
		simple_serial.NewReader(cfg),
	}
}

func (p *Platform) Setup(cfg *config.UserConfig, ns chan<- notifications.Notification) error {
	kbd, err := input.NewKeyboard()
	if err != nil {
		return err
	}
	p.kbd = kbd

	gpd, err := uinput.CreateGamepad(
		"/dev/uinput",
		[]byte("tapto"),
		0x1234,
		0x5678,
	)
	if err != nil {
		return err
	}
	p.gpd = gpd

	tr, stopTr, err := mister.StartTracker(*mister.UserConfigToMrext(cfg), ns)
	if err != nil {
		return err
	}

	p.tr = tr
	p.stopTr = stopTr

	err = mister.Setup(p.tr)
	if err != nil {
		return err
	}

	return nil
}

func (p *Platform) Stop() error {
	if p.stopTr != nil {
		return p.stopTr()
	}

	if p.gpd != nil {
		err := p.gpd.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Platform) AfterScanHook(token tokens.Token) error {
	f, err := os.Create(mister.TokenReadFile)
	if err != nil {
		return fmt.Errorf("unable to create scan result file %s: %s", mister.TokenReadFile, err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	_, err = f.WriteString(fmt.Sprintf("%s,%s", token.UID, token.Text))
	if err != nil {
		return fmt.Errorf("unable to write scan result file %s: %s", mister.TokenReadFile, err)
	}

	return nil
}

func (p *Platform) ReadersUpdateHook(readers map[string]*readers.Reader) error {
	return nil
}

func (p *Platform) RootFolders(cfg *config.UserConfig) []string {
	return games.GetGamesFolders(mister.UserConfigToMrext(cfg))
}

func (p *Platform) ZipsAsFolders() bool {
	return true
}

func (p *Platform) ConfigFolder() string {
	return mister.ConfigFolder
}

func (p *Platform) LogFolder() string {
	return mister.TempFolder
}

func (p *Platform) NormalizePath(cfg *config.UserConfig, path string) string {
	return mister.NormalizePath(cfg, path)
}

func LaunchMenu() error {
	if _, err := os.Stat(mrextConfig.CmdInterface); err != nil {
		return fmt.Errorf("command interface not accessible: %s", err)
	}

	cmd, err := os.OpenFile(mrextConfig.CmdInterface, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer cmd.Close()

	// TODO: hardcoded for xilinx variant, should read pref from mister.ini
	cmd.WriteString(fmt.Sprintf("load_core %s\n", filepath.Join(mrextConfig.SdFolder, "menu.bit")))

	return nil
}

func (p *Platform) KillLauncher() error {
	return LaunchMenu()
}

func (p *Platform) LaunchingEnabled() bool {
	_, err := os.Stat(mister.DisableLaunchFile)
	return err != nil
}

func (p *Platform) SetLaunching(disabled bool) error {
	if disabled {
		return os.Remove(mister.DisableLaunchFile)
	} else {
		_, err := os.Create(mister.DisableLaunchFile)
		return err
	}
}

func (p *Platform) GetActiveLauncher() string {
	core := mister.GetActiveCoreName()

	if core == mrextConfig.MenuCore {
		return ""
	}

	return core
}

func (p *Platform) PlayFailSound(cfg *config.UserConfig) {
	mister.PlayFail(cfg)
}

func (p *Platform) PlaySuccessSound(cfg *config.UserConfig) {
	mister.PlaySuccess(cfg)
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

func (p *Platform) LaunchSystem(cfg *config.UserConfig, id string) error {
	system, err := games.LookupSystem(id)
	if err != nil {
		return err
	}

	return mm.LaunchCore(mister.UserConfigToMrext(cfg), *system)
}

func (p *Platform) LaunchFile(cfg *config.UserConfig, path string) error {
	return mm.LaunchGenericFile(mister.UserConfigToMrext(cfg), path)
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

func (p *Platform) KeyboardPress(name string) error {
	code, ok := mister.KeyboardMap[name]
	if !ok {
		return fmt.Errorf("unknown key: %s", name)
	}

	if code < 0 {
		p.kbd.Combo(42, -code)
	} else {
		p.kbd.Press(code)
	}

	return nil
}

func (p *Platform) GamepadPress(name string) error {
	code, ok := mister.GamepadMap[name]
	if !ok {
		return fmt.Errorf("unknown button: %s", name)
	}

	p.gpd.ButtonDown(code)
	time.Sleep(40 * time.Millisecond)
	p.gpd.ButtonUp(code)

	return nil
}

func (p *Platform) ForwardCmd(env platforms.CmdEnv) error {
	if f, ok := commandsMappings[env.Cmd]; ok {
		return f(p, env)
	} else {
		return fmt.Errorf("command not supported on mister: %s", env.Cmd)
	}
}

func (p *Platform) LookupMapping(_ tokens.Token) (string, bool) {
	return "", false
}

func (p *Platform) Launchers() []platforms.Launcher {
	return mister.Launchers
}
