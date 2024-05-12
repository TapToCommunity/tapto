package daemon

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/clausecker/nfc/v2"
	"github.com/rs/zerolog/log"
	mrextConfig "github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
	"github.com/wizzomafizzo/tapto/pkg/utils"
)

const (
	timeToForgetCard   = 500 * time.Millisecond
	connectMaxTries    = 10
	TimesToPoll        = 1
	PeriodBetweenPolls = 250 * time.Millisecond
	periodBetweenLoop  = 250 * time.Millisecond
)

func pollDevice(
	cfg *config.UserConfig,
	pnd *nfc.Device,
	activeCard state.Token,
	ttp int,
	pbp time.Duration,
) (state.Token, bool, error) {
	removed := false

	count, target, err := pnd.InitiatorPollTarget(tokens.SupportedCardTypes, ttp, pbp)
	if err != nil && !errors.Is(err, nfc.Error(nfc.ETIMEOUT)) {
		return activeCard, removed, err
	}

	if count > 1 {
		log.Info().Msg("More than one card on the reader")
	}

	if count <= 0 {
		if activeCard.UID != "" && time.Since(activeCard.ScanTime) > timeToForgetCard {
			log.Info().Msg("card removed")
			activeCard = state.Token{}
			removed = true
		}

		return activeCard, removed, nil
	}

	cardUid := tokens.GetCardUID(target)
	if cardUid == "" {
		log.Warn().Msgf("unable to detect token UID: %s", target.String())
	}

	if cardUid == activeCard.UID {
		return activeCard, removed, nil
	}

	log.Info().Msgf("found token UID: %s", cardUid)

	var record []byte
	cardType := tokens.GetCardType(target)

	if cardType == tokens.TypeNTAG {
		log.Info().Msg("NTAG detected")
		record, err = tokens.ReadNtag(*pnd)
		if err != nil {
			return activeCard, removed, fmt.Errorf("error reading ntag: %s", err)
		}
		cardType = tokens.TypeNTAG
	}

	if cardType == tokens.TypeMifare {
		log.Info().Msg("Mifare detected")
		record, err = tokens.ReadMifare(*pnd, cardUid)
		if err != nil {
			log.Error().Msgf("error reading mifare: %s", err)
		}
		cardType = tokens.TypeMifare
	}

	log.Debug().Msgf("record bytes: %s", hex.EncodeToString(record))
	tagText := tokens.ParseRecordText(record)
	if tagText == "" {
		log.Warn().Msg("no text NDEF found")
	} else {
		log.Info().Msgf("decoded text NDEF: %s", tagText)
	}

	card := state.Token{
		Type:     cardType,
		UID:      cardUid,
		Text:     tagText,
		ScanTime: time.Now(),
	}

	return card, removed, nil
}

func detectConnectionString(quiet bool) string {
	if !quiet {
		log.Info().Msg("probing for serial devices")
	}
	devices, _ := utils.GetLinuxSerialDeviceList()

	for _, device := range devices {
		connectionString := "pn532_uart:" + device
		pnd, err := nfc.Open(connectionString)
		log.Info().Msgf("trying %s", connectionString)
		if err == nil {
			log.Info().Msgf("success using serial: %s", connectionString)
			pnd.Close()
			return connectionString
		}
	}

	return ""
}

func OpenDeviceWithRetries(cfg *config.UserConfig, st *state.State, quiet bool) (nfc.Device, error) {
	var connectionString = cfg.GetConnectionString()
	if connectionString == "" && cfg.GetProbeDevice() == true {
		connectionString = detectConnectionString(quiet)
	}

	if !quiet {
		log.Info().Msgf("connecting to device: %s", connectionString)
	}

	tries := 0
	for {
		pnd, err := nfc.Open(connectionString)
		if err == nil {
			log.Info().Msgf("successful connect after %d tries", tries)

			connProto := strings.SplitN(strings.ToLower(connectionString), ":", 2)[0]
			log.Info().Msgf("connection protocol: %s", connProto)
			deviceName := pnd.String()
			log.Info().Msgf("device name: %s", deviceName)

			if connProto == "pn532_uart" {
				st.SetReaderConnected(state.ReaderTypePN532)
			} else if strings.Contains(deviceName, "ACR122U") {
				st.SetReaderConnected(state.ReaderTypeACR122U)
			} else {
				st.SetReaderConnected(state.ReaderTypeUnknown)
			}

			return pnd, err
		}

		if tries >= connectMaxTries {
			if !quiet {
				log.Error().Msgf("could not open device after %d tries: %s", connectMaxTries, err)
			}
			return pnd, err
		}

		tries++
	}
}

func shouldExit(
	candidateForRemove bool,
	cfg *config.UserConfig,
	st *state.State,
) bool {
	// do not exit from menu, there is nowhere to go anyway
	if mister.GetActiveCoreName() == mrextConfig.MenuCore {
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

	if hasTimePassed && cfg.GetExitGame() && !inExitGameBlocklist(cfg) {
		return true
	} else {
		return false
	}
}

func readerPollLoop(
	cfg *config.UserConfig,
	st *state.State,
	tq *state.TokenQueue,
	kbd input.Keyboard,
) {
	var pnd nfc.Device
	var err error

	ttp := TimesToPoll
	pbp := PeriodBetweenPolls
	var lastError time.Time
	var candidateForRemove bool
	playFail := func() {
		if time.Since(lastError) > 1*time.Second {
			mister.PlayFail(cfg)
		}
	}

	log.Debug().Msgf("polling for %d times with %s delay", ttp, pbp)

	for {
		if st.ShouldStopService() {
			break
		}

		time.Sleep(periodBetweenLoop)

		if connected, _ := st.GetReaderStatus(); !connected {
			quiet := time.Since(lastError) < 1*time.Second

			// TODO: keep track of reconnect attempts?
			if !quiet {
				log.Info().Msg("reader not connected, attempting connection....")
			}

			pnd, err = OpenDeviceWithRetries(cfg, st, quiet)
			if err != nil {
				lastError = time.Now()
				continue
			}

			if err := pnd.InitiatorInit(); err != nil {
				st.SetReaderDisconnected()
				log.Error().Msgf("could not init initiator: %s", err)
				continue
			}

			log.Info().Msgf("opened connection: %s %s", pnd, pnd.Connection())
		}

		// activeCard is the card that sat on the scanner at the previous poll loop.
		// is not the card representing the current loaded core
		activeCard := st.GetActiveCard()
		writeRequest := st.GetWriteRequest()

		if writeRequest != "" {
			log.Info().Msgf("write request: %s", writeRequest)

			var count int
			var target nfc.Target
			tries := 4 * 30 // ~30 seconds

			for tries > 0 {
				count, target, err = pnd.InitiatorPollTarget(tokens.SupportedCardTypes, ttp, pbp)

				if err != nil && err.Error() != "timeout" {
					log.Error().Msgf("could not poll: %s", err)
				}

				if count > 0 {
					break
				}

				tries--
			}

			if count == 0 {
				log.Error().Msgf("could not detect a card")
				st.SetWriteRequest("")
				continue
			}

			cardUid := tokens.GetCardUID(target)
			log.Info().Msgf("found card with UID: %s", cardUid)

			cardType := tokens.GetCardType(target)
			var bytesWritten []byte

			switch cardType {
			case tokens.TypeMifare:
				bytesWritten, err = tokens.WriteMifare(pnd, writeRequest, cardUid)
				if err != nil {
					log.Error().Msgf("error writing to mifare: %s", err)
					st.SetWriteRequest("")
					continue
				}
			case tokens.TypeNTAG:
				bytesWritten, err = tokens.WriteNtag(pnd, writeRequest)
				if err != nil {
					log.Error().Msgf("error writing to ntag: %s", err)
					st.SetWriteRequest("")
					continue
				}
			default:
				log.Error().Msgf("unsupported card type: %s", cardType)
				st.SetWriteRequest("")
				continue
			}

			log.Info().Msgf("successfully wrote to card: %s", hex.EncodeToString(bytesWritten))
			st.SetWriteRequest("")
			continue
		}

		newScanned, removed, err := pollDevice(cfg, &pnd, activeCard, ttp, pbp)

		// if we removed but we weren't removing already, start the remove countdown
		if removed && candidateForRemove == false {
			st.SetCardRemovalTime(time.Now())
			candidateForRemove = true
			// if we were removing but we put back the card we had before
			// then we are ok blocking the exit process
		} else if candidateForRemove && (newScanned.UID == st.GetCurrentlyLoadedSoftware()) {
			log.Info().Msgf("Card was removed but inserted back")
			st.SetCardRemovalTime(time.Time{})
			candidateForRemove = false
		}

		if errors.Is(err, nfc.Error(nfc.EIO)) {
			st.SetReaderDisconnected()
			log.Error().Msgf("error during poll: %s", err)
			log.Error().Msg("fatal IO error, device was possibly unplugged")
			playFail()
			lastError = time.Now()
			continue
		} else if err != nil {
			log.Error().Msgf("error during poll: %s", err)
			playFail()
			lastError = time.Now()
			continue
		}

		// this will update the state for the activeCard
		// the local variable activeCard is still the previous one and will be updated next loop
		st.SetActiveCard(newScanned)

		if shouldExit(candidateForRemove, cfg, st) {
			candidateForRemove = false
			st.SetCardRemovalTime(time.Time{})
			mister.ExitGame()
			st.SetCurrentlyLoadedSoftware("")
			continue
		} else if mister.GetActiveCoreName() == mrextConfig.MenuCore {
			// at any time we are on the current menu we should forget old values
			candidateForRemove = false
			st.SetCurrentlyLoadedSoftware("")
		}

		// From here we didn't exit a game, but we want short circuit and do nothing if the following happens

		// in any case if the new scanned card has no UID we never want to go on with launching anything
		// if the card is the same as the one we have scanned before ( activeCard.UID == newScanned.UID) we don't relaunch
		// this will avoid card left on the reader to trigger the command multiple times per second
		// in order to tap a card fast, so insert a coin multiple times, you have to get on and off from the reader with the card

		if newScanned.UID == "" || activeCard.UID == newScanned.UID {
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
		mister.PlaySuccess(cfg)

		if st.IsLauncherDisabled() {
			continue
		}

		log.Info().Msgf("About to process token %s: \n current software: %s \n activeCard: %s \n", newScanned.UID, st.GetCurrentlyLoadedSoftware(), activeCard.UID)

		// we are about to exec a command, we reset timers, we evaluate next loop if we need to start exiting again
		st.SetCardRemovalTime(time.Time{})
		candidateForRemove = false
		tq.Enqueue(newScanned)
	}

	st.SetReaderDisconnected()
	err = pnd.Close()
	if err != nil {
		log.Warn().Msgf("error closing device: %s", err)
	}
}
