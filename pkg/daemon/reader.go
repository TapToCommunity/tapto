package daemon

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/clausecker/nfc/v2"
	"github.com/rs/zerolog/log"
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
	TimesToPoll        = 20
	PeriodBetweenPolls = 300 * time.Millisecond
	periodBetweenLoop  = 300 * time.Millisecond
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

func OpenDeviceWithRetries(config config.TapToConfig, st *state.State, quiet bool) (nfc.Device, error) {
	var connectionString = config.ConnectionString
	if connectionString == "" && config.ProbeDevice == true {
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
	playFail := func() {
		if time.Since(lastError) > 1*time.Second {
			mister.PlayFail(cfg)
		}
	}

	if cfg.TapTo.ExitGame {
		// FIXME: this method makes the activity indicator flicker, is there another way?
		ttp = 1
		// TODO: value requires investigation, originally set to 150 which worked for pn532
		//       but not for acr122u (read once then never again). 200 seems to work ok
		pbp = 200 * time.Millisecond
	}

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

			pnd, err = OpenDeviceWithRetries(cfg.TapTo, st, quiet)
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

		activeCard := st.GetActiveCard()
		writeRequest := st.GetWriteRequest()

		if writeRequest != "" {
			log.Info().Msgf("write request: %s", writeRequest)

			var count int
			var target nfc.Target
			tries := 6 // ~30 seconds

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

		log.Debug().Msgf("polling for %d times with %s delay", ttp, pbp)
		newScanned, removed, err := pollDevice(cfg, &pnd, activeCard, ttp, pbp)

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

		st.SetActiveCard(newScanned)

		if removed && cfg.TapTo.ExitGame && !inExitGameBlocklist(cfg) && !st.IsLauncherDisabled() {
			mister.ExitGame()
			continue
		}

		if newScanned.UID == "" || activeCard.UID == newScanned.UID {
			continue
		}

		mister.PlaySuccess(cfg)

		if st.IsLauncherDisabled() {
			continue
		}

		tq.Enqueue(newScanned)
	}

	st.SetReaderDisconnected()
	err = pnd.Close()
	if err != nil {
		log.Warn().Msgf("error closing device: %s", err)
	}
}
