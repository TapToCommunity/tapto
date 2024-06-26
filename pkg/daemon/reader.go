package daemon

import (
	"crypto/sha256"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/readers"
	"github.com/wizzomafizzo/tapto/pkg/readers/file"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
	// "github.com/wizzomafizzo/tapto/pkg/readers/libnfc"
)

func shouldExit(
	pl platforms.Platform,
	candidateForRemove bool,
	cfg *config.UserConfig,
	st *state.State,
	removalTime time.Time,
) bool {
	// do not exit from menu, there is nowhere to go anyway
	if !pl.IsLauncherActive() {
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
		log.Info().Msgf("Exiting game after %.2f seconds have passed with a configured %d seconds delay", time.Since(removalTime).Seconds(), cfg.GetExitGameDelay())
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

func readerManager(
	pl platforms.Platform,
	cfg *config.UserConfig,
	st *state.State,
	lq *tokens.TokenQueue,
	lsq <-chan tokens.Token,
) {
	// reader token input queue
	iq := make(chan readers.Scan)

	var err error
	var lastError time.Time
	var candidateForRemove bool

	var cardRemovalTime time.Time
	var loadedSoftware string

	// keep track of core switch for menu reset
	var lastCoreName string = ""

	playFail := func() {
		if time.Since(lastError) > 1*time.Second {
			pl.PlayFailSound(cfg)
		}
	}

	for !st.ShouldStopService() {
		reader := st.GetReader()

		if reader == nil || !reader.Connected() {
			log.Info().Msg("reader not connected, attempting connection....")

			// reader = libnfc.NewReader(cfg)
			reader = file.NewReader(cfg)

			device := cfg.GetConnectionString()
			if device == "" {
				device := reader.Detect(nil)
				if device == "" {
					log.Error().Msg("no reader detected")
					continue
				}
			}

			err = reader.Open(device, iq)
			if err != nil {
				log.Error().Msgf("error opening device: %s", err)
				continue
			}

			st.SetReader(reader)
			continue
		}

		// activeCard is the card that sat on the scanner at the previous poll loop.
		// is not the card representing the current loaded core
		var newScanned *tokens.Token

		select {
		case t := <-iq:
			log.Debug().Msgf("processing token: %v", t)
			if t.Error != nil {
				log.Error().Msgf("error reading card: %s", err)
				playFail()
				lastError = time.Now()
				continue
			}
			newScanned = t.Token
		case softwareToken := <-lsq:
			loadedSoftware = tokenHash(softwareToken)
			continue
		}

		removed := newScanned == nil // TODO: is this right?

		if cfg.GetExitGame() {
			// if we removed but we weren't removing already, start the remove countdown
			if removed && candidateForRemove == false {
				log.Info().Msgf("Start countdown for removal")
				cardRemovalTime = time.Now()
				candidateForRemove = true
				// if we were removing but we put back the card we had before
				// then we are ok blocking the exit process
			} else if candidateForRemove && tokenHash(*newScanned) == loadedSoftware {
				log.Info().Msgf("card was removed but inserted back")
				cardRemovalTime = time.Time{}
				candidateForRemove = false
			}
		}

		// this will update the state for the activeCard
		// the local variable activeCard is still the previous one and will be updated next loop
		if newScanned != nil {
			log.Info().Msgf("new card scanned: %s", newScanned.UID)
			st.SetActiveCard(*newScanned)
		}

		if shouldExit(pl, candidateForRemove, cfg, st, cardRemovalTime) {
			candidateForRemove = false
			cardRemovalTime = time.Time{}
			_ = pl.KillLauncher()
			loadedSoftware = ""
			continue
		} else if !pl.IsLauncherActive() && lastCoreName != "" {
			// at any time we are on the current menu we should forget old values if we have anything to clear
			candidateForRemove = false
			cardRemovalTime = time.Time{}
			loadedSoftware = ""
		}

		lastCoreName = pl.GetActiveLauncher()

		// From here we didn't exit a game, but we want short circuit and do nothing if the following happens

		// in any case if the new scanned card has no UID we never want to go on with launching anything
		// if the card is the same as the one we have scanned before ( activeCard.UID == newScanned.UID) we don't relaunch
		// this will avoid card left on the reader to trigger the command multiple times per second
		// in order to tap a card fast, so insert a coin multiple times, you have to get on and off from the reader with the card

		if tokenHash(st.GetActiveCard()) == tokenHash(*newScanned) {
			continue
		}

		// if the card has the same ID of the currently loaded software it means we re-read a card that was already there
		// this could happen in combination with exit_game_delay and tapping for coins or other commands not meant to interrupt
		// a game. In that case when we put back the same software card, we don't want to reboot, only to keep running it
		if loadedSoftware == tokenHash(*newScanned) {
			// keeping a separate if to have specific logging
			log.Info().Msgf("token with UID %s has been skipped because is the currently loaded software", tokenHash(*newScanned))
			candidateForRemove = false
			continue
		}

		if st.IsLauncherDisabled() {
			continue
		} else {
			pl.PlaySuccessSound(cfg)
		}

		log.Info().Msgf("about to process token %s: \n current software: %s \n activeCard: %s \n", newScanned.UID, loadedSoftware, st.GetActiveCard())

		// we are about to exec a command, we reset timers, we evaluate next loop if we need to start exiting again
		cardRemovalTime = time.Time{}
		candidateForRemove = false

		if newScanned != nil {
			lq.Enqueue(*newScanned)
		}
	}

	reader := st.GetReader()
	if reader != nil {
		err = reader.Close()
		if err != nil {
			log.Warn().Msgf("error closing device: %s", err)
		}
	}
}
