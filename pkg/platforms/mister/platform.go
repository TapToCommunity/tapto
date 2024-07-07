package mister

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/bendahl/uinput"
	"github.com/rs/zerolog/log"
	mrextConfig "github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

type Platform struct {
	kbd                 input.Keyboard
	gpd                 uinput.Gamepad
	tr                  *Tracker
	stopTr              func() error
	dbLoadTime          time.Time
	uidMap              map[string]string
	textMap             map[string]string
	stopMappingsWatcher func() error
	cmdMappings         map[string]func(platforms.Platform, platforms.CmdEnv) error
}

type oldDb struct {
	Uids  map[string]string
	Texts map[string]string
}

func (p *Platform) getDB() oldDb {
	return oldDb{
		Uids:  p.uidMap,
		Texts: p.textMap,
	}
}

func (p *Platform) GetDBLoadTime() time.Time {
	return p.dbLoadTime
}

func (p *Platform) SetDB(uidMap map[string]string, textMap map[string]string) {
	p.dbLoadTime = time.Now()
	p.uidMap = uidMap
	p.textMap = textMap
}

func (p *Platform) Id() string {
	return "mister"
}

func (p *Platform) Setup(cfg *config.UserConfig) error {
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

	tr, stopTr, err := StartTracker(*UserConfigToMrext(cfg))
	if err != nil {
		return err
	}

	p.tr = tr
	p.stopTr = stopTr

	uids, texts, err := LoadCsvMappings()
	if err != nil {
		log.Error().Msgf("error loading mappings: %s", err)
	} else {
		p.SetDB(uids, texts)
	}

	closeMappingsWatcher, err := StartCsvMappingsWatcher(
		p.GetDBLoadTime,
		p.SetDB,
	)
	if err != nil {
		log.Error().Msgf("error starting mappings watcher: %s", err)
	}
	p.stopMappingsWatcher = closeMappingsWatcher

	err = Setup(p.tr)
	if err != nil {
		return err
	}

	p.cmdMappings = map[string]func(platforms.Platform, platforms.CmdEnv) error{
		"mister.ini":    CmdIni,
		"mister.core":   CmdLaunchCore,
		"mister.script": cmdMisterScript(*p),
		"mister.mgl":    CmdMisterMgl,

		"ini": CmdIni, // DEPRECATED
	}

	return nil
}

func (p *Platform) Stop() error {
	if p.stopTr != nil {
		err := p.stopTr()
		if err != nil {
			return err
		}
	}

	if p.gpd != nil {
		err := p.gpd.Close()
		if err != nil {
			return err
		}
	}

	if p.stopMappingsWatcher != nil {
		err := p.stopMappingsWatcher()
		if err != nil {
			return err
		}
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

func (p *Platform) LaunchingEnabled() bool {
	_, err := os.Stat(DisableLaunchFile)
	return err != nil
}

func (p *Platform) SetLaunching(disabled bool) error {
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

func (p *Platform) KeyboardPress(name string) error {
	code, ok := KeyboardMap[name]
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
	code, ok := GamepadMap[name]
	if !ok {
		return fmt.Errorf("unknown button: %s", name)
	}

	p.gpd.ButtonDown(code)
	time.Sleep(40 * time.Millisecond)
	p.gpd.ButtonUp(code)

	return nil
}

func (p *Platform) ForwardCmd(env platforms.CmdEnv) error {
	if f, ok := p.cmdMappings[env.Cmd]; ok {
		return f(p, env)
	} else {
		return fmt.Errorf("command not supported on mister: %s", env.Cmd)
	}
}

func (p *Platform) LookupMapping(t tokens.Token) (string, bool) {
	oldDb := p.getDB()

	// check nfc.csv uids
	if v, ok := oldDb.Uids[t.UID]; ok {
		log.Info().Msg("launching with csv uid match override")
		return v, true
	}

	// check nfc.csv texts
	for pattern, cmd := range oldDb.Texts {
		// check if pattern is a regex
		re, err := regexp.Compile(pattern)

		// not a regex
		if err != nil {
			if pattern, ok := oldDb.Texts[t.Text]; ok {
				log.Info().Msg("launching with csv text match override")
				return pattern, true
			}
		}

		// regex
		if re.MatchString(t.Text) {
			log.Info().Msg("launching with csv regex text match override")
			return cmd, true
		}
	}

	return "", false
}
