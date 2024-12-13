package main

import (
	"github.com/ZaparooProject/zaparoo-core/pkg/utils"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"strings"

	_ "embed"
)

// TODO: allow updating if files have changed

//go:embed conf/zaparoo.service
var serviceFile string

//go:embed conf/blacklist-zaparoo.conf
var modprobeFile string

//go:embed conf/60-zaparoo.rules
var udevFile string

const (
	servicePath  = "/etc/systemd/system/zaparoo.service"
	modprobePath = "/etc/modprobe.d/blacklist-zaparoo.conf"
	udevPath     = "/etc/udev/rules.d/60-zaparoo.rules"
)

func install() error {
	// install and prep systemd service
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		log.Info().Msg("installing zaparoo service")
		serviceFile = strings.ReplaceAll(
			serviceFile,
			"%%INSTALL_DIR%%",
			utils.ExeDir(),
		)
		err = os.WriteFile(servicePath, []byte(serviceFile), 0644)
		if err != nil {
			return err
		}
		log.Info().Msg("wrote service file")
		err = exec.Command("systemctl", "daemon-reload").Run()
		if err != nil {
			return err
		}
		log.Info().Msg("reloaded systemd units")
		err = exec.Command("systemctl", "enable", "zaparoo").Run()
		if err != nil {
			return err
		}
		log.Info().Msg("enabled zaparoo service")
	}

	// install udev rules and refresh
	if _, err := os.Stat(udevPath); os.IsNotExist(err) {
		log.Info().Msg("installing udev rules")
		err = os.WriteFile(udevPath, []byte(udevFile), 0644)
		if err != nil {
			return err
		}
		log.Info().Msg("wrote udev rules")
		err = exec.Command("udevadm", "control", "--reload-rules").Run()
		if err != nil {
			return err
		}
		log.Info().Msg("reloaded udev rules")
		err = exec.Command("udevadm", "trigger").Run()
		if err != nil {
			return err
		}
		log.Info().Msg("triggered udev rules")
	}

	// install modprobe blacklist
	if _, err := os.Stat(modprobePath); os.IsNotExist(err) {
		log.Info().Msg("installing modprobe blacklist")
		err = os.WriteFile(modprobePath, []byte(modprobeFile), 0644)
		if err != nil {
			return err
		}
		log.Info().Msg("wrote modprobe blacklist")
	}

	return nil
}

func uninstall() error {
	if _, err := os.Stat(servicePath); !os.IsNotExist(err) {
		log.Info().Msg("uninstalling zaparoo service")
		err = exec.Command("systemctl", "disable", "zaparoo").Run()
		if err != nil {
			return err
		}
		log.Info().Msg("disabled zaparoo service")
		err = exec.Command("systemctl", "stop", "zaparoo").Run()
		if err != nil {
			return err
		}
		log.Info().Msg("stopped zaparoo service")
		err = exec.Command("systemctl", "daemon-reload").Run()
		if err != nil {
			return err
		}
		log.Info().Msg("reloaded systemd units")
		err = os.Remove(servicePath)
		if err != nil {
			return err
		}
		log.Info().Msg("removed service file")
	}

	if _, err := os.Stat(modprobePath); !os.IsNotExist(err) {
		log.Info().Msg("uninstalling modprobe blacklist")
		err = os.Remove(modprobePath)
		if err != nil {
			return err
		}
		log.Info().Msg("removed modprobe blacklist")
	}

	if _, err := os.Stat(udevPath); !os.IsNotExist(err) {
		log.Info().Msg("uninstalling udev rules")
		err = os.Remove(udevPath)
		if err != nil {
			return err
		}
		log.Info().Msg("removed udev rules")
	}

	return nil
}
