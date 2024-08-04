package platforms

import (
	"path/filepath"
	"strings"

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

type Launcher struct {
	// Id of the launcher, visible to user
	Id string
	// System associated with this launcher
	SystemId string
	// Folders to scan for files, relative to the root folder
	Folders []string
	// Extensions to match for files
	Extensions []string
	// Launch function, takes absolute path to file as argument
	Launch func(*config.UserConfig, string) error
	// TODO: optional after scan hook to find special files or modify the
	// standard scan results
}

// MatchSystemFile returns true if a given file's extension is valid for a system.
func MatchSystemFile(pl Platform, systemId string, path string) bool {
	var launchers []Launcher
	for _, l := range pl.Launchers() {
		if l.SystemId == systemId {
			launchers = append(launchers, l)
		}
	}

	// ignore dot files
	if strings.HasPrefix(filepath.Base(path), ".") {
		return false
	}

	for _, l := range launchers {
		for _, ext := range l.Extensions {
			if strings.HasSuffix(strings.ToLower(path), ext) {
				return true
			}
		}
	}

	return false
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
	Launchers() []Launcher
}
