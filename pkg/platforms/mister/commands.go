package mister

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
)

var commandsMappings = map[string]func(Platform, platforms.CmdEnv) error{
	"mister.ini":  cmdIni,
	"mister.core": cmdLaunchCore,
	// "mister.script": cmdMisterScript,
	"mister.mgl": cmdMisterMgl,

	"ini": cmdIni, // DEPRECATED
}

func cmdIni(pl Platform, env platforms.CmdEnv) error {
	inis, err := mister.GetAllMisterIni()
	if err != nil {
		return err
	}

	if len(inis) == 0 {
		return fmt.Errorf("no ini files found")
	}

	id, err := strconv.Atoi(env.Args)
	if err != nil {
		return err
	}

	if id < 1 || id > len(inis) {
		return fmt.Errorf("ini id out of range: %d", id)
	}

	doRelaunch := true
	// only relaunch if there aren't any more commands
	if env.TotalCommands > 1 && env.CurrentIndex < env.TotalCommands-1 {
		doRelaunch = false
	}

	err = mister.SetActiveIni(id, doRelaunch)
	if err != nil {
		return err
	}

	return nil
}

func cmdLaunchCore(pl Platform, env platforms.CmdEnv) error {
	return mister.LaunchShortCore(env.Args)
}

func cmdMisterScript(pl Platform, env platforms.CmdEnv) error {
	// TODO: escaping arguments
	// TODO: does this work if game is running?

	if mister.IsScriptRunning() {
		return fmt.Errorf("script already running")
	}

	args := strings.Fields(env.Args)

	if len(args) == 0 {
		return fmt.Errorf("no script specified")
	}

	script := args[0]

	if !strings.HasSuffix(script, ".sh") {
		return fmt.Errorf("invalid script: %s", script)
	}

	scriptPath := filepath.Join(ScriptsFolder, script)
	if _, err := os.Stat(scriptPath); err != nil {
		return fmt.Errorf("script not found: %s", script)
	}

	script = scriptPath

	args = args[1:]
	if len(args) > 0 {
		scriptArgs := strings.Join(args, " ")
		script = fmt.Sprintf("%s %s", script, scriptArgs)
	}

	return mister.RunScript(pl.kbd, script)
}

func cmdMisterMgl(pl Platform, env platforms.CmdEnv) error {
	if env.Args == "" {
		return fmt.Errorf("no mgl specified")
	}

	tmpFile, err := os.CreateTemp("", "*.mgl")
	if err != nil {
		return err
	}

	_, err = tmpFile.WriteString(env.Args)
	if err != nil {
		return err
	}

	err = tmpFile.Close()
	if err != nil {
		return err
	}

	cmd, err := os.OpenFile(CmdInterface, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer cmd.Close()

	_, err = cmd.WriteString(fmt.Sprintf("load_core %s\n", tmpFile.Name()))
	if err != nil {
		return err
	}

	go func() {
		time.Sleep(5 * time.Second)
		_ = os.Remove(tmpFile.Name())
	}()

	return nil
}
