//go:build linux || darwin

package mistex

import (
	"github.com/ZaparooProject/zaparoo-core/pkg/platforms"
	"github.com/ZaparooProject/zaparoo-core/pkg/platforms/mister"
)

var commandsMappings = map[string]func(platforms.Platform, platforms.CmdEnv) error{
	"mister.ini":  mister.CmdIni,
	"mister.core": mister.CmdLaunchCore,
	// "mister.script": cmdMisterScript,
	"mister.mgl": mister.CmdMisterMgl,

	"ini": mister.CmdIni, // DEPRECATED
}
