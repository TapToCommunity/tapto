package platforms

import (
	"github.com/wizzomafizzo/tapto/pkg/service/notifications"
	"path/filepath"
	"strings"

	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/readers"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

type CmdEnv struct {
	Cmd           string
	Args          string
	NamedArgs     map[string]string
	Cfg           *config.UserConfig
	Manual        bool
	Text          string
	TotalCommands int
	CurrentIndex  int
}

type ScanResult struct {
	Path string
	Name string
}

type Launcher struct {
	// Unique ID of the launcher, visible to user.
	Id string
	// System associated with this launcher.
	SystemId string
	// Folders to scan for files, relative to the root folders of the platform.
	// TODO: Support absolute paths?
	Folders []string
	// Extensions to match for files during a standard scan.
	Extensions []string
	// Accepted schemes for URI-style launches.
	Schemes []string
	// Launch function, takes a direct as possible path/ID media file.
	Launch func(*config.UserConfig, string) error
	// Optional function to perform custom media scanning. Takes the list of
	// results from the standard scan, if any, and returns the final list.
	Scanner func(*config.UserConfig, []ScanResult) ([]ScanResult, error)
	// If true, all resolved paths must be in the allow list before they
	// can be launched.
	AllowListOnly bool
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
	// Unique ID of the platform.
	Id() string
	// Any initial setup required before daemon is fully started.
	Setup(*config.UserConfig, chan<- notifications.Notification) error
	// TOOD: what is this?
	Stop() error
	// Run immediately after a successful scan, before it is processed for launching.
	AfterScanHook(tokens.Token) error
	// Run after the active readers have been updated.
	ReadersUpdateHook(map[string]*readers.Reader) error
	// List of supported readers for this platform.
	SupportedReaders(*config.UserConfig) []readers.Reader
	// List of root folders to scan for media files.
	RootFolders(*config.UserConfig) []string
	// Whether to treat zip files as folders during media scanning.
	ZipsAsFolders() bool
	// Path to the configuration/database data for TapTo.
	ConfigFolder() string // TODO: rename to data folder (because that's what it is)
	// Path to the log folder for TapTo.
	LogFolder() string
	// Convert a path to a normalized form for the platform, the shortest
	// possible path that can interpreted and lanched by TapTo. For writing
	// to tokens.
	NormalizePath(*config.UserConfig, string) string
	// Kill the currently running launcher process if possible.
	KillLauncher() error
	LaunchingEnabled() bool  // TODO: remove? should be mister only?
	SetLaunching(bool) error // TODO: remove? should be mister only?
	// Return the ID of the currently active launcher. Empty string if none.
	GetActiveLauncher() string
	// Play a sound effect for error feedback.
	PlayFailSound(*config.UserConfig) // TODO: change to like PlaySound?
	// Play a sound effect for success feedback.
	PlaySuccessSound(*config.UserConfig)
	// Returns the currently active system ID.
	ActiveSystem() string
	// Returns the currently active game ID.
	ActiveGame() string // TODO: check where this is used
	// Returns the currently active game name.
	ActiveGameName() string
	// Returns the currently active game path.
	ActiveGamePath() string
	// Launch a system by ID.
	LaunchSystem(*config.UserConfig, string) error
	// Launch a file by path.
	LaunchFile(*config.UserConfig, string) error
	// Launch a shell command.
	Shell(string) error
	KeyboardInput(string) error // DEPRECATED
	KeyboardPress(string) error
	GamepadPress(string) error
	// Process a token command that has been resolved to a platform command.
	ForwardCmd(CmdEnv) error
	LookupMapping(tokens.Token) (string, bool)
	Launchers() []Launcher
}
