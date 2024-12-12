package utils

import (
	"github.com/ZaparooProject/zaparoo-core/pkg/config"
	"github.com/ZaparooProject/zaparoo-core/pkg/platforms"
	"path/filepath"
	"strings"
)

// PathIsLauncher returns true if a given path matches against any of the
// criteria defined in a launcher.
func PathIsLauncher(
	cfg *config.UserConfig,
	pl platforms.Platform,
	l platforms.Launcher,
	path string,
) bool {
	if len(path) == 0 {
		return false
	}

	lp := strings.ToLower(path)

	// ignore dot files
	if strings.HasPrefix(filepath.Base(lp), ".") {
		return false
	}

	// check uri scheme
	for _, scheme := range l.Schemes {
		if strings.HasPrefix(lp, scheme+":") {
			return true
		}
	}

	// check root folder if it's not a generic launcher
	if len(l.Folders) > 0 {
		inRoot := false
		for _, folder := range pl.RootFolders(cfg) {
			if strings.HasPrefix(lp, strings.ToLower(folder)) {
				inRoot = true
				break
			}
		}

		if !inRoot {
			return false
		} else if path[len(path)-1] == filepath.Separator {
			// skip extension check if it's a folder
			return true
		}
	}

	// check file extension
	for _, ext := range l.Extensions {
		if strings.HasSuffix(lp, ext) {
			return true
		}
	}

	return false
}

// MatchSystemFile returns true if a given path is for a given system.
func MatchSystemFile(
	cfg *config.UserConfig,
	pl platforms.Platform,
	systemId string,
	path string,
) bool {
	for _, l := range pl.Launchers() {
		if l.SystemId == systemId {
			if PathIsLauncher(cfg, pl, l, path) {
				return true
			}
		}
	}
	return false
}

// PathToLaunchers is a reverse lookup to match a given path against all
// possible launchers in a platform. Returns all matched launchers.
func PathToLaunchers(
	cfg *config.UserConfig,
	pl platforms.Platform,
	path string,
) []platforms.Launcher {
	var launchers []platforms.Launcher
	for _, l := range pl.Launchers() {
		if PathIsLauncher(cfg, pl, l, path) {
			launchers = append(launchers, l)
		}
	}
	return launchers
}
