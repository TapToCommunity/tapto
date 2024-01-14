package utils

import (
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/mister"
	"github.com/wizzomafizzo/mrext/pkg/service"
)

// NfcMigration renames old NFC app data and removes the old service.
func NfcMigration(logger *service.Logger) {
	nfcAppPath := "/media/fat/Scripts/nfc.sh"
	nfcIniPath := "/media/fat/Scripts/nfc.ini"
	nfcPid := "/tmp/nfc.pid"
	taptoIniPath := "/media/fat/Scripts/tapto.ini"

	// check if nfc.sh is running and kill it
	if _, err := os.Stat(nfcPid); err == nil {
		pidFile, err := os.ReadFile(nfcPid)
		if err != nil {
			logger.Error("error reading pid file: %s", err)
		} else {
			pidInt, err := strconv.Atoi(string(pidFile))
			if err != nil {
				logger.Error("error parsing pid: %s", err)
			} else {
				logger.Info("killing nfc.sh")
				err := syscall.Kill(pidInt, syscall.SIGTERM)
				if err != nil {
					logger.Error("error killing nfc.sh: %s", err)
				}
				time.Sleep(1 * time.Second)
			}
		}
	}

	// remove nfc.sh from startup if it exists
	var startup mister.Startup
	err := startup.Load()
	if err != nil {
		logger.Error("failed to load startup file: %s", err)
	}

	if startup.Exists("mrext/nfc") {
		logger.Info("removing nfc.sh from startup")
		err := startup.Remove("mrext/nfc")
		if err != nil {
			logger.Error("error removing nfc.sh from startup: %s", err)
		}

		err = startup.Save()
		if err != nil {
			logger.Error("error saving startup file: %s", err)
		}
	}

	// if old nfc ini exists and new one doesn't, rename it
	if _, err := os.Stat(nfcIniPath); err == nil {
		if _, err := os.Stat(taptoIniPath); err != nil {
			err := os.Rename(nfcIniPath, taptoIniPath)
			if err != nil {
				logger.Error("error renaming nfc.ini: %s", err)
			} else {
				logger.Info("renamed nfc.ini to tapto.ini")
			}
		} else {
			// or just remove the old one
			err := os.Remove(nfcIniPath)
			if err != nil {
				logger.Error("error removing nfc.ini: %s", err)
			} else {
				logger.Info("removed nfc.ini")
			}
		}
	}

	// remove old nfc.sh
	if _, err := os.Stat(nfcAppPath); err == nil {
		err := os.Remove(nfcAppPath)
		if err != nil {
			logger.Error("error removing nfc.sh: %s", err)
		} else {
			logger.Info("removed nfc.sh")
		}
	}
}
