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
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/clausecker/nfc/v2"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/launcher"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

const (
	timeToForgetCard   = 500 * time.Millisecond
	connectMaxTries    = 10
	TimesToPoll        = 20
	PeriodBetweenPolls = 300 * time.Millisecond
	periodBetweenLoop  = 300 * time.Millisecond
)

type Card struct {
	CardType string
	UID      string
	Text     string
	ScanTime time.Time
}

type ServiceState struct {
	mu              sync.Mutex
	activeCard      Card
	lastScanned     Card
	stopService     bool
	disableLauncher bool
	dbLoadTime      time.Time
	uidMap          map[string]string
	textMap         map[string]string
}

func (s *ServiceState) SetActiveCard(card Card) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeCard = card
	if s.activeCard.UID != "" {
		s.lastScanned = card
	}
}

func (s *ServiceState) GetActiveCard() Card {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.activeCard
}

func (s *ServiceState) GetLastScanned() Card {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastScanned
}

func (s *ServiceState) StopService() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopService = true
}

func (s *ServiceState) ShouldStopService() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopService
}

func (s *ServiceState) DisableLauncher() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.disableLauncher = true
	if _, err := os.Create(mister.DisableLaunchFile); err != nil {
		log.Error().Msgf("cannot create disable launch file: %s", err)
	}
}

func (s *ServiceState) EnableLauncher() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.disableLauncher = false
	if err := os.Remove(mister.DisableLaunchFile); err != nil {
		log.Error().Msgf("cannot remove disable launch file: %s", err)
	}
}

func (s *ServiceState) IsLauncherDisabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.disableLauncher
}

func (s *ServiceState) GetDB() (map[string]string, map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.uidMap, s.textMap
}

func (s *ServiceState) GetDBLoadTime() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dbLoadTime
}

func (s *ServiceState) SetDB(uidMap map[string]string, textMap map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dbLoadTime = time.Now()
	s.uidMap = uidMap
	s.textMap = textMap
}

func pollDevice(
	cfg *config.UserConfig,
	pnd *nfc.Device,
	activeCard Card,
	ttp int,
	pbp time.Duration,
) (Card, error) {
	count, target, err := pnd.InitiatorPollTarget(tokens.SupportedCardTypes, ttp, pbp)
	if err != nil && !errors.Is(err, nfc.Error(nfc.ETIMEOUT)) {
		return activeCard, err
	}

	if count <= 0 {
		if activeCard.UID != "" && time.Since(activeCard.ScanTime) > timeToForgetCard {
			log.Info().Msg("card removed")
			activeCard = Card{}

			if cfg.TapTo.ExitGame {
				mister.ExitGame()
			}
		}

		return activeCard, nil
	}

	cardUid := tokens.GetCardUID(target)
	if cardUid == "" {
		log.Warn().Msgf("unable to detect token UID: %s", target.String())
	}

	if cardUid == activeCard.UID {
		return activeCard, nil
	}

	log.Info().Msgf("found token UID: %s", cardUid)

	var record []byte
	cardType := tokens.GetCardType(target)

	if cardType == tokens.TypeNTAG {
		log.Info().Msg("NTAG detected")
		record, err = tokens.ReadNtag(*pnd)
		if err != nil {
			return activeCard, fmt.Errorf("error reading ntag: %s", err)
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

	card := Card{
		CardType: cardType,
		UID:      cardUid,
		Text:     tagText,
		ScanTime: time.Now(),
	}

	return card, nil
}

func getSerialDeviceList() ([]string, error) {
	path := "/dev/serial/by-id/"
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	files, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}

	var devices []string

	for _, v := range files {
		if !v.IsDir() {
			devices = append(devices, path+v.Name())
		}
	}

	return devices, nil
}

func detectConnectionString() string {
	log.Info().Msg("probing for serial devices")
	devices, _ := getSerialDeviceList()

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

func OpenDeviceWithRetries(config config.TapToConfig) (nfc.Device, error) {
	var connectionString = config.ConnectionString
	if connectionString == "" && config.ProbeDevice == true {
		connectionString = detectConnectionString()
	}

	tries := 0
	for {
		pnd, err := nfc.Open(connectionString)
		if err == nil {
			log.Info().Msgf("successful connect after %d tries", tries)
			return pnd, err
		}

		if tries >= connectMaxTries {
			log.Error().Msgf("could not open device after %d tries: %s", connectMaxTries, err)
			return pnd, err
		}

		tries++
	}
}

func writeScanResult(card Card) error {
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

func launchCard(cfg *config.UserConfig, state *ServiceState, kbd input.Keyboard) error {
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

	for _, cmd := range cmds {
		err := launcher.LaunchToken(cfg, cfg.TapTo.AllowCommands || override, kbd, cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func StartService(cfg *config.UserConfig) (func() error, error) {
	state := &ServiceState{}

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

	var closeDbWatcher func() error
	dbWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error().Msgf("error creating mappings watcher: %s", err)
	} else {
		closeDbWatcher = dbWatcher.Close
	}

	go func() {
		// this turned out to be not trivial to say the least, mostly due to
		// the fact the fsnotify library does not implement the IN_CLOSE_WRITE
		// inotify event, which signals the file has finished being written
		// see: https://github.com/fsnotify/fsnotify/issues/372
		//
		// during a standard write operation, a file may emit multiple write
		// events, including when the file could be half-written
		//
		// it's also the case that editors may delete the file and create a new
		// one, which kills the active watcher
		const delay = 1 * time.Second
		for {
			select {
			case event, ok := <-dbWatcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					// usually receives multiple write events, just act on the first
					if time.Since(state.GetDBLoadTime()) < delay {
						continue
					}
					time.Sleep(delay)
					log.Info().Msg("database changed, reloading")
					uids, texts, err := launcher.LoadDatabase()
					if err != nil {
						log.Error().Msgf("error loading database: %s", err)
					} else {
						state.SetDB(uids, texts)
					}
				} else if event.Has(fsnotify.Remove) {
					// editors may also delete the file on write
					time.Sleep(delay)
					_, err := os.Stat(mister.MappingsFile)
					if err == nil {
						err = dbWatcher.Add(mister.MappingsFile)
						if err != nil {
							log.Error().Msgf("error watching database: %s", err)
						}
						log.Info().Msg("database changed, reloading")
						uids, texts, err := launcher.LoadDatabase()
						if err != nil {
							log.Error().Msgf("error loading database: %s", err)
						} else {
							state.SetDB(uids, texts)
						}
					}
				}
			case err, ok := <-dbWatcher.Errors:
				if !ok {
					return
				}
				log.Error().Msgf("watcher error: %s", err)
			}
		}
	}()

	err = dbWatcher.Add(mister.MappingsFile)
	if err != nil {
		log.Error().Msgf("error watching database: %s", err)
	}

	if _, err := os.Stat(mister.DisableLaunchFile); err == nil {
		state.DisableLauncher()
	}

	go func() {
		var pnd nfc.Device
		var err error

		ttp := TimesToPoll
		pbp := PeriodBetweenPolls

		if cfg.TapTo.ExitGame {
			// FIXME: this method makes the activity indicator flicker, is there another way?
			ttp = 1
			// TODO: value requires investigation, originally set to 150 which worked for pn532
			//       but not for acr122u (read once then never again). 200 seems to work ok
			pbp = 200 * time.Millisecond
		}

	reconnect:
		pnd, err = OpenDeviceWithRetries(cfg.TapTo)
		if err != nil {
			return
		}

		defer func(pnd nfc.Device) {
			err := pnd.Close()
			if err != nil {
				log.Warn().Msgf("error closing device: %s", err)
			}
		}(pnd)

		if err := pnd.InitiatorInit(); err != nil {
			log.Error().Msgf("could not init initiator: %s", err)
			return
		}

		log.Info().Msgf("opened connection: %s %s", pnd, pnd.Connection())
		log.Info().Msgf("polling for %d times with %s delay", ttp, pbp)
		var lastError time.Time

		for {
			if state.ShouldStopService() {
				break
			}

			activeCard := state.GetActiveCard()
			newScanned, err := pollDevice(cfg, &pnd, activeCard, ttp, pbp)
			if errors.Is(err, nfc.Error(nfc.EIO)) {
				log.Error().Msgf("error during poll: %s", err)
				log.Error().Msg("fatal IO error, device was unplugged, exiting...")
				if time.Since(lastError) > 1*time.Second {
					mister.PlayFail(cfg)
				}
				goto reconnect
			} else if err != nil {
				log.Error().Msgf("error during poll: %s", err)
				if time.Since(lastError) > 1*time.Second {
					mister.PlayFail(cfg)
				}
				lastError = time.Now()
				goto end
			}

			state.SetActiveCard(newScanned)

			if newScanned.UID == "" || activeCard.UID == newScanned.UID {
				goto end
			}

			mister.PlaySuccess(cfg)

			err = writeScanResult(newScanned)
			if err != nil {
				log.Error().Msgf("error writing tmp scan result: %s", err)
			}

			if state.IsLauncherDisabled() {
				log.Info().Msg("launcher disabled, skipping")
				goto end
			}

			err = launchCard(cfg, state, kbd)
			if err != nil {
				log.Error().Msgf("error launching card: %s", err)
				if time.Since(lastError) > 1*time.Second {
					mister.PlayFail(cfg)
				}
				lastError = time.Now()
				goto end
			}

		end:
			time.Sleep(periodBetweenLoop)
		}
	}()

	socket, err := net.Listen("unix", mister.SocketFile)
	if err != nil {
		log.Error().Msgf("error creating socket: %s", err)
		return nil, err
	}

	go func() {
		for {
			if state.ShouldStopService() {
				break
			}

			conn, err := socket.Accept()
			if err != nil {
				log.Error().Msgf("error accepting connection: %s", err)
				return
			}

			go func(conn net.Conn) {
				log.Debug().Msg("new socket connection")

				defer func(conn net.Conn) {
					err := conn.Close()
					if err != nil {
						log.Warn().Msgf("error closing connection: %s", err)
					}
				}(conn)

				buf := make([]byte, 4096)

				n, err := conn.Read(buf)
				if err != nil {
					log.Error().Msgf("error reading from connection: %s", err)
					return
				}

				if n == 0 {
					return
				}
				log.Debug().Msgf("received %d bytes", n)

				payload := ""

				switch strings.TrimSpace(string(buf[:n])) {
				case "status":
					lastScanned := state.GetLastScanned()
					if lastScanned.UID != "" {
						payload = fmt.Sprintf(
							"%d,%s,%t,%s",
							lastScanned.ScanTime.Unix(),
							lastScanned.UID,
							!state.IsLauncherDisabled(),
							lastScanned.Text,
						)
					} else {
						payload = fmt.Sprintf("0,,%t,", !state.IsLauncherDisabled())
					}
				case "disable":
					state.DisableLauncher()
					log.Info().Msg("launcher disabled")
				case "enable":
					state.EnableLauncher()
					log.Info().Msg("launcher enabled")
				default:
					log.Warn().Msgf("unknown command: %s", strings.TrimRight(string(buf[:n]), "\n"))
				}

				_, err = conn.Write([]byte(payload))
				if err != nil {
					log.Error().Msgf("error writing to socket: %s", err)
					return
				}
			}(conn)
		}
	}()

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
