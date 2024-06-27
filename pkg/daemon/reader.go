package daemon

import (
	"crypto/sha256"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/readers"

	//"github.com/wizzomafizzo/tapto/pkg/readers/file"
	"github.com/wizzomafizzo/tapto/pkg/readers/libnfc"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

func shouldExit(
	pl platforms.Platform,
	candidateForRemove bool,
	cfg *config.UserConfig,
	st *state.State,
	removalTime time.Time,
) bool {
	// do not exit from menu, there is nowhere to go anyway
	if pl.GetActiveLauncher() == "" {
		return false
	}

	// candidateForRemove is true from the moment in which we remove a card
	if !candidateForRemove || st.GetLastScanned().FromApi || st.IsLauncherDisabled() {
		return false
	}

	var hasTimePassed bool = false
	if !removalTime.IsZero() {
		hasTimePassed = int8(time.Since(removalTime).Seconds()) >= cfg.GetExitGameDelay()
	}

	if hasTimePassed && cfg.GetExitGame() && !inExitGameBlocklist(pl, cfg) {
		log.Info().Msgf("exiting game after %.2f seconds have passed with a configured %d seconds delay", time.Since(removalTime).Seconds(), cfg.GetExitGameDelay())
		return true
	} else {
		return false
	}
}

func tokenHash(t tokens.Token) string {
	h := sha256.New()
	h.Write([]byte(t.UID))
	h.Write([]byte(t.Text))
	return string(h.Sum(nil))
}

func connectReaders(
	cfg *config.UserConfig,
	st *state.State,
	iq chan<- readers.Scan,
) error {
	reader := st.GetReader()

	if reader == nil || !reader.Connected() {
		log.Info().Msg("reader not connected, attempting connection....")

		reader = libnfc.NewReader(cfg)
		// reader = file.NewReader(cfg)

		device := cfg.GetConnectionString()
		if device == "" {
			log.Debug().Msg("no device specified, attempting to detect...")
			device = reader.Detect(nil)
			if device == "" {
				return errors.New("no reader detected")
			}
		}

		err := reader.Open(device, iq)
		if err != nil {
			return err
		}

		st.SetReader(reader)
	}

	return nil
}

func readerManager(
	pl platforms.Platform,
	cfg *config.UserConfig,
	st *state.State,
	launchQueue *tokens.TokenQueue,
	softwareQueue <-chan tokens.Token,
) {
	// reader token input queue
	inputQueue := make(chan readers.Scan)

	var err error
	var lastError time.Time

	playFail := func() {
		if time.Since(lastError) > 1*time.Second {
			pl.PlayFailSound(cfg)
		}
	}

	// manage reader connections and exit game behaviour
	go func() {
		for !st.ShouldStopService() {
			reader := st.GetReader()
			if reader == nil || !reader.Connected() {
				err := connectReaders(cfg, st, inputQueue)
				if err != nil {
					log.Error().Msgf("error connecting readers: %s", err)
				}
			}

			if shouldExit(pl, st.IsRemovalCandidate(), cfg, st, st.GetRemovalTime()) {
				log.Debug().Msg("should exit, killing launcher...")
				st.SetRemovalCandidate(false)
				st.SetRemovalTime(time.Time{})
				_ = pl.KillLauncher()
				st.SetLoadedSoftware("")
			} else if pl.GetActiveLauncher() == "" {
				// at any time we are on the current menu we should forget old
				// values if we have anything to clear
				//log.Debug().Msg("not in launcher, clearing old values")
				st.SetRemovalCandidate(false)
				st.SetRemovalTime(time.Time{})
				st.SetLoadedSoftware("")
			}

			time.Sleep(500 * time.Millisecond)
		}
	}()

	// token pre-processing loop
	for !st.ShouldStopService() {
		var scan *tokens.Token

		select {
		case t := <-inputQueue:
			// a reader has sent a token for pre-processing
			log.Debug().Msgf("processing token: %v", t)
			if t.Error != nil {
				log.Error().Msgf("error reading card: %s", err)
				playFail()
				lastError = time.Now()
				continue
			}
			scan = t.Token
		case softwareToken := <-softwareQueue:
			// a token has been launched that starts software
			log.Debug().Msgf("set software token: %v", softwareToken)
			st.SetLoadedSoftware(tokenHash(softwareToken))
			continue
		}

		if cfg.GetExitGame() {
			log.Debug().Msgf("exit game section")

			if scan == nil && st.IsRemovalCandidate() == false {
				// if we removed but we weren't removing already, start the
				// remove countdown
				log.Info().Msgf("start countdown for removal")
				st.SetRemovalTime(time.Now())
				st.SetRemovalCandidate(true)
			} else if scan != nil && st.IsRemovalCandidate() && tokenHash(*scan) == st.GetLoadedSoftware() {
				// if we were removing but we put back the card we had before
				// then we are ok blocking the exit process
				log.Info().Msgf("token was removed but reinserted")
				st.SetRemovalCandidate(false)
				st.SetRemovalTime(time.Time{})
			}
		}

		// this will update the state for the activeCard
		// the local variable activeCard is still the previous one and will
		// be updated next loop
		if scan != nil {
			log.Info().Msgf("new card scanned: %v", scan)
			st.SetActiveCard(*scan)
		}

		// if the card has the same ID of the currently loaded software it
		// means we re-read a card that was already there
		// this could happen in combination with exit_game_delay and
		// tapping for coins or other commands not meant to interrupt
		// a game. In that case when we put back the same software card,
		// we don't want to reboot, only to keep running it
		if scan != nil && cfg.GetExitGame() && st.GetLoadedSoftware() == tokenHash(*scan) {
			// keeping a separate if to have specific logging
			log.Info().Msgf("token is same software: %s", tokenHash(*scan))
			st.SetRemovalCandidate(false)
			continue
		}

		if st.IsLauncherDisabled() || scan == nil {
			log.Debug().Msg("no active token")
			continue
		} else {
			log.Info().Msgf(
				"sending token %v: , current software: %s, activeCard: %v",
				scan,
				st.GetLoadedSoftware(),
				st.GetActiveCard(),
			)
			pl.PlaySuccessSound(cfg)
			launchQueue.Enqueue(*scan)
		}
	}

	// daemon shutdown
	reader := st.GetReader()
	if reader != nil {
		err = reader.Close()
		if err != nil {
			log.Warn().Msgf("error closing device: %s", err)
		}
	}
}
