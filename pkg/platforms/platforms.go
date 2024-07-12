package platforms

import (
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/readers"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

// FIXME: i don't like where this is, but it's currently resolving
// a circular dependency between the launcher and platforms packages
type CmdEnv struct {
	Cmd           string
	Args          string
	Cfg           *config.UserConfig
	Manual        bool
	Text          string
	TotalCommands int
	CurrentIndex  int
}

type Platform interface {
	Id() string
	Setup(*config.UserConfig) error
	Stop() error
	AfterScanHook(tokens.Token) error
	ReadersUpdateHook(map[string]*readers.Reader) error
	SupportedReaders(*config.UserConfig) []readers.Reader
	RootFolders(*config.UserConfig) []string
	ZipsAsFolders() bool
	ConfigFolder() string // TODO: rename to data folder (because that's what it is)
	LogFolder() string
	NormalizePath(*config.UserConfig, string) string
	KillLauncher() error
	LaunchingEnabled() bool  // TODO: remove? should be mister only?
	SetLaunching(bool) error // TODO: remove? should be mister only?
	GetActiveLauncher() string
	PlayFailSound(*config.UserConfig) // TODO: change to like PlaySound?
	PlaySuccessSound(*config.UserConfig)
	ActiveSystem() string
	ActiveGame() string // TODO: check where this is used
	ActiveGameName() string
	ActiveGamePath() string
	SetEventHook(*func())
	LaunchSystem(*config.UserConfig, string) error
	LaunchFile(*config.UserConfig, string) error
	Shell(string) error
	KeyboardInput(string) error // DEPRECATED
	KeyboardPress(string) error
	GamepadPress(string) error
	ForwardCmd(CmdEnv) error
	LookupMapping(tokens.Token) (string, bool)
}
