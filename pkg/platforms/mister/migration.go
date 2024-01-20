package mister

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/mrext/pkg/mister"
)

// NfcMigration migrates NFC app data and removes the old service.
func NfcMigration() {
	nfcAppPath := "/media/fat/Scripts/nfc.sh"
	nfcIniPath := "/media/fat/Scripts/nfc.ini"
	nfcPid := "/tmp/nfc.pid"
	taptoIniPath := "/media/fat/Scripts/tapto.ini"

	// check if nfc.sh is running and kill it
	if _, err := os.Stat(nfcPid); err == nil {
		pidFile, err := os.ReadFile(nfcPid)
		if err != nil {
			log.Error().Msgf("migration: error reading pid file: %s", err)
		} else {
			pidInt, err := strconv.Atoi(string(pidFile))
			if err != nil {
				log.Error().Msgf("migration: error parsing pid: %s", err)
			} else {
				log.Info().Msg("migration: nfc.sh is running, killing it")
				err := syscall.Kill(pidInt, syscall.SIGTERM)
				if err != nil {
					log.Error().Msgf("migration: error killing nfc.sh: %s", err)
				}
				time.Sleep(1 * time.Second)
			}
		}
	}

	// remove nfc.sh from startup if it exists
	var startup mister.Startup
	err := startup.Load()
	if err != nil {
		log.Error().Msgf("migration: failed to load startup file: %s", err)
	}

	if startup.Exists("mrext/nfc") {
		log.Info().Msg("migration: removing nfc.sh from startup")
		err := startup.Remove("mrext/nfc")
		if err != nil {
			log.Error().Msgf("migration: error removing nfc.sh from startup: %s", err)
		}

		err = startup.Save()
		if err != nil {
			log.Error().Msgf("migration: error saving startup file: %s", err)
		}
	}

	// if old nfc ini exists and new one doesn't, migrate it
	if _, err := os.Stat(nfcIniPath); err == nil {
		if _, err := os.Stat(taptoIniPath); err != nil {
			err := os.Rename(nfcIniPath, taptoIniPath)
			if err != nil {
				log.Error().Msgf("migration: error renaming nfc.ini: %s", err)
			} else {
				log.Info().Msg("migration: renamed nfc.ini to tapto.ini")
			}

			// replace [nfc] header with [tapto]
			contents, err := os.ReadFile(taptoIniPath)
			if err != nil {
				log.Error().Msgf("migration: error reading tapto.ini: %s", err)
			}

			lines := strings.Split(string(contents), "\n")

			changed := false
			for i, line := range lines {
				if strings.HasPrefix(line, "[nfc]") {
					changed = true
					lines[i] = "[tapto]"
				}
			}

			if changed {
				err = os.WriteFile(taptoIniPath, []byte(strings.Join(lines, "\n")), 0644)
				if err != nil {
					log.Error().Msgf("migration: error writing tapto.ini: %s", err)
				} else {
					log.Info().Msg("migration: replaced [nfc] header with [tapto]")
				}
			}
		} else {
			// or just remove the old one
			err := os.Remove(nfcIniPath)
			if err != nil {
				log.Error().Msgf("migration: error removing nfc.ini: %s", err)
			} else {
				log.Info().Msg("migration: removed nfc.ini")
			}
		}
	}

	// remove old nfc.sh
	if _, err := os.Stat(nfcAppPath); err == nil {
		err := os.Remove(nfcAppPath)
		if err != nil {
			log.Error().Msgf("migration: error removing nfc.sh: %s", err)
		} else {
			log.Info().Msg("migration: removed nfc.sh")
		}
	}
}
