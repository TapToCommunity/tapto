package platforms

import (
	"github.com/wizzomafizzo/tapto/pkg/config"
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
	RootFolders(*config.UserConfig) []string
	ZipsAsFolders() bool
	ConfigFolder() string
	NormalizePath(*config.UserConfig, string) string
	KillLauncher() error
	LaunchingEnabled() bool
	SetLaunching(bool) error
	GetActiveLauncher() string
	PlayFailSound(*config.UserConfig)
	PlaySuccessSound(*config.UserConfig)
	ActiveSystem() string
	ActiveGame() string
	ActiveGameName() string
	ActiveGamePath() string
	SetEventHook(*func())
	LaunchSystem(*config.UserConfig, string) error
	LaunchFile(*config.UserConfig, string) error
	Shell(string) error
	KeyboardInput(string) error // DEPRECATED
	KeyboardPress(string) error
	ForwardCmd(CmdEnv) error
}
