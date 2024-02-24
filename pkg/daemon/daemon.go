/*
TapTo
Copyright (C) 2023 Gareth Jones
Copyright (C) 2023, 2024 Callan Barrett

This file is part of TapTo.

TapTo is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

TapTo is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with TapTo.  If not, see <http://www.gnu.org/licenses/>.
*/

package daemon

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/clausecker/nfc/v2"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/launcher"
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
	activeCard Token,
	ttp int,
	pbp time.Duration,
) (Token, bool, error) {
	removed := false

	count, target, err := pnd.InitiatorPollTarget(tokens.SupportedCardTypes, ttp, pbp)
	if err != nil && !errors.Is(err, nfc.Error(nfc.ETIMEOUT)) {
		return activeCard, removed, err
	}

	if count <= 0 {
		if activeCard.UID != "" && time.Since(activeCard.ScanTime) > timeToForgetCard {
			log.Info().Msg("card removed")
			activeCard = Token{}
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

	card := Token{
		Type:     cardType,
		UID:      cardUid,
		Text:     tagText,
		ScanTime: time.Now(),
	}

	return card, removed, nil
}

func detectConnectionString() string {
	log.Info().Msg("probing for serial devices")
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

func OpenDeviceWithRetries(config config.TapToConfig, state *State) (nfc.Device, error) {
	var connectionString = config.ConnectionString
	if connectionString == "" && config.ProbeDevice == true {
		connectionString = detectConnectionString()
	}

	log.Info().Msgf("connecting to device: %s", connectionString)

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
				state.SetReaderConnected(ReaderTypePN532)
			} else if strings.Contains(deviceName, "ACR122U") {
				state.SetReaderConnected(ReaderTypeACR122U)
			} else {
				state.SetReaderConnected(ReaderTypeUnknown)
			}

			return pnd, err
		}

		if tries >= connectMaxTries {
			log.Error().Msgf("could not open device after %d tries: %s", connectMaxTries, err)
			return pnd, err
		}

		tries++
	}
}

func writeScanResult(card Token) error {
	f, err := os.Create(mister.TokenReadFile)
	if err != nil {
		return fmt.Errorf("unable to create scan result file %s: %s", mister.TokenReadFile, err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	_, err = f.WriteString(fmt.Sprintf("%s,%s", card.UID, card.Text))
	if err != nil {
		return fmt.Errorf("unable to write scan result file %s: %s", mister.TokenReadFile, err)
	}

	return nil
}

func launchCard(cfg *config.UserConfig, state *State, kbd input.Keyboard) error {
	card := state.GetActiveCard()
	uidMap, textMap := state.GetDB()

	text := card.Text
	override := false

	if v, ok := uidMap[card.UID]; ok {
		log.Info().Msg("launching with uid match override")
		text = v
		override = true
	}

	if v, ok := textMap[card.Text]; ok {
		log.Info().Msg("launching with text match override")
		text = v
		override = true
	}

	if text == "" {
		return fmt.Errorf("no text NDEF found in card or database")
	}

	log.Info().Msgf("launching with text: %s", text)
	cmds := strings.Split(text, "||")

	for i, cmd := range cmds {
		err := launcher.LaunchToken(cfg, cfg.TapTo.AllowCommands || override, kbd, cmd, len(cmds), i)
		if err != nil {
			return err
		}
	}

	return nil
}

func pollLoop(
	cfg *config.UserConfig,
	state *State,
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
		if state.ShouldStopService() {
			break
		}

		time.Sleep(periodBetweenLoop)

		if connected, _ := state.GetReaderStatus(); !connected {
			// TODO: keep track of reconnect attempts?
			log.Info().Msg("reader not connected, attempting connection....")

			pnd, err = OpenDeviceWithRetries(cfg.TapTo, state)
			if err != nil {
				continue
			}

			if err := pnd.InitiatorInit(); err != nil {
				state.SetReaderDisconnected()
				log.Error().Msgf("could not init initiator: %s", err)
				continue
			}

			log.Info().Msgf("opened connection: %s %s", pnd, pnd.Connection())
		}

		activeCard := state.GetActiveCard()

		log.Debug().Msgf("polling for %d times with %s delay", ttp, pbp)
		newScanned, removed, err := pollDevice(cfg, &pnd, activeCard, ttp, pbp)

		if errors.Is(err, nfc.Error(nfc.EIO)) {
			state.SetReaderDisconnected()
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

		state.SetActiveCard(newScanned)

		if removed && cfg.TapTo.ExitGame && cfg.TapTo.ExitGame && !strings.Contains(cfg.TapTo.ExitGameBlocklikst, mister.GetActiveCoreName()) && !state.IsLauncherDisabled() {
			mister.ExitGame()
			continue
		}

		if newScanned.UID == "" || activeCard.UID == newScanned.UID {
			continue
		}

		mister.PlaySuccess(cfg)

		err = writeScanResult(newScanned)
		if err != nil {
			log.Error().Msgf("error writing tmp scan result: %s", err)
		}

		if state.IsLauncherDisabled() {
			log.Info().Msg("launcher disabled, skipping command action")
			continue
		}

		err = launchCard(cfg, state, kbd)
		if err != nil {
			log.Error().Msgf("error launching card: %s", err)
			playFail()
			lastError = time.Now()
			continue
		}
	}

	state.SetReaderDisconnected()
	err = pnd.Close()
	if err != nil {
		log.Warn().Msgf("error closing device: %s", err)
	}
}

func StartDaemon(cfg *config.UserConfig) (func() error, error) {
	state := &State{}

	// TODO: this is platform specific
	kbd, err := input.NewKeyboard()
	if err != nil {
		log.Error().Msgf("failed to initialize keyboard: %s", err)
		return nil, err
	}

	uids, texts, err := launcher.LoadDatabase()
	if err != nil {
		log.Error().Msgf("error loading database: %s", err)
	} else {
		state.SetDB(uids, texts)
	}

	closeDbWatcher, err := launcher.StartMappingsWatcher(
		state.GetDBLoadTime,
		state.SetDB,
	)
	if err != nil {
		log.Error().Msgf("error starting database watcher: %s", err)
	}

	if _, err := os.Stat(mister.DisableLaunchFile); err == nil {
		state.DisableLauncher()
	}

	go pollLoop(cfg, state, kbd)

	socket, err := StartSocketServer(state)
	if err != nil {
		log.Error().Msgf("error starting socket server: %s", err)
		return nil, err
	}

	return func() error {
		err := socket.Close()
		if err != nil {
			log.Warn().Msgf("error closing socket: %s", err)
		}
		state.StopService()
		if closeDbWatcher != nil {
			return closeDbWatcher()
		}
		return nil
	}, nil
}
