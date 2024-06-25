package daemon

import (
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/readers/libnfc"
)

func shouldExit(
	pl platforms.Platform,
	candidateForRemove bool,
	cfg *config.UserConfig,
	st *state.State,
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
	var removalTime = st.GetCardRemovalTime()
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

func readerLoop(
	platform platforms.Platform,
	cfg *config.UserConfig,
	st *state.State,
	tq *state.TokenQueue,
) {
	var err error
	var lastError time.Time
	var candidateForRemove bool

	// keep track of core switch for menu reset
	var lastCoreName string = ""

	playFail := func() {
		if time.Since(lastError) > 1*time.Second {
			platform.PlayFailSound(cfg)
		}
	}

	for {
		if st.ShouldStopService() {
			break
		}

		reader := st.GetReader()

		if reader == nil || !reader.Connected() {
			log.Info().Msg("reader not connected, attempting connection....")

			reader = libnfc.NewReader(cfg)

			device := reader.Detect(nil)
			if device == "" {
				log.Error().Msg("no reader detected")
				continue
			}

			err = reader.Open(device)
			if err != nil {
				log.Error().Msgf("error opening device: %s", err)
				continue
			}

			st.SetReader(reader)
			continue
		}

		// activeCard is the card that sat on the scanner at the previous poll loop.
		// is not the card representing the current loaded core
		activeCard := st.GetActiveCard()

		newScanned, err := reader.Read()
		if err != nil {
			log.Error().Msgf("error reading card: %s", err)
			playFail()
			lastError = time.Now()
			continue
		}

		removed := newScanned == nil // TODO: is this right?

		if cfg.GetExitGame() {
			// if we removed but we weren't removing already, start the remove countdown
			if removed && candidateForRemove == false {
				log.Info().Msgf("Start countdown for removal")
				st.SetCardRemovalTime(time.Now())
				candidateForRemove = true
				// if we were removing but we put back the card we had before
				// then we are ok blocking the exit process
			} else if candidateForRemove && (newScanned.UID == st.GetCurrentlyLoadedSoftware()) {
				log.Info().Msgf("Card was removed but inserted back")
				st.SetCardRemovalTime(time.Time{})
				candidateForRemove = false
			}
		}

		if err != nil {
			log.Error().Msgf("error during read: %s", err)
			playFail()
			lastError = time.Now()
			continue
		}

		// this will update the state for the activeCard
		// the local variable activeCard is still the previous one and will be updated next loop
		if newScanned != nil {
			st.SetActiveCard(*newScanned)
		}

		if shouldExit(platform, candidateForRemove, cfg, st) {
			candidateForRemove = false
			st.SetCardRemovalTime(time.Time{})
			_ = platform.KillLauncher()
			st.SetCurrentlyLoadedSoftware("")
			continue
		} else if !platform.IsLauncherActive() && lastCoreName != "" {
			// at any time we are on the current menu we should forget old values if we have anything to clear
			candidateForRemove = false
			st.SetCardRemovalTime(time.Time{})
			st.SetCurrentlyLoadedSoftware("")
		}

		lastCoreName = platform.GetActiveLauncher()

		// From here we didn't exit a game, but we want short circuit and do nothing if the following happens

		// in any case if the new scanned card has no UID we never want to go on with launching anything
		// if the card is the same as the one we have scanned before ( activeCard.UID == newScanned.UID) we don't relaunch
		// this will avoid card left on the reader to trigger the command multiple times per second
		// in order to tap a card fast, so insert a coin multiple times, you have to get on and off from the reader with the card

		if newScanned == nil || activeCard.UID == newScanned.UID {
			continue
		}

		// if the card has the same ID of the currently loaded software it means we re-read a card that was already there
		// this could happen in combination with exit_game_delay and tapping for coins or other commands not meant to interrupt
		// a game. In that case when we put back the same software card, we don't want to reboot, only to keep running it
		if st.GetCurrentlyLoadedSoftware() == newScanned.UID {
			// keeping a separate if to have specific logging
			log.Info().Msgf("Token with UID %s has been skipped because is the currently loaded software", newScanned.UID)
			candidateForRemove = false
			continue
		}

		// should we play success if launcher is disabled?
		platform.PlaySuccessSound(cfg)

		if st.IsLauncherDisabled() {
			continue
		}

		log.Info().Msgf("About to process token %s: \n current software: %s \n activeCard: %s \n", newScanned.UID, st.GetCurrentlyLoadedSoftware(), activeCard.UID)

		// we are about to exec a command, we reset timers, we evaluate next loop if we need to start exiting again
		st.SetCardRemovalTime(time.Time{})
		candidateForRemove = false

		if newScanned != nil {
			tq.Enqueue(*newScanned)
		}

		// time.Sleep(100 * time.Millisecond)
	}

	reader := st.GetReader()
	err = reader.Close()
	if err != nil {
		log.Warn().Msgf("error closing device: %s", err)
	}
}
