package daemon

import (
	"errors"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/readers"
	"github.com/wizzomafizzo/tapto/pkg/utils"

	"github.com/wizzomafizzo/tapto/pkg/readers/file"
	"github.com/wizzomafizzo/tapto/pkg/readers/libnfc"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

func shouldExit(
	cfg *config.UserConfig,
	pl platforms.Platform,
	st *state.State,
) bool {
	if !cfg.GetExitGame() {
		return false
	}

	// do not exit from menu, there is nowhere to go anyway
	if pl.GetActiveLauncher() == "" {
		return false
	}

	if st.GetLastScanned().FromApi || st.IsLauncherDisabled() {
		return false
	}

	if inExitGameBlocklist(pl, cfg) {
		return false
	}

	return true
}

func cmpTokens(a, b *tokens.Token) bool {
	if a == nil || b == nil {
		return false
	}

	return a.UID == b.UID && a.Text == b.Text
}

func connectReaders(
	cfg *config.UserConfig,
	st *state.State,
	iq chan<- readers.Scan,
) error {
	rs := st.ListReaders()
	var toConnect []string

	userDevice := cfg.GetConnectionString()
	if userDevice != "" && !utils.Contains(rs, userDevice) {
		log.Debug().Msgf("user device not connected, adding: %s", userDevice)
		toConnect = append(toConnect, userDevice)
	}

	for _, device := range toConnect {
		if _, ok := st.GetReader(device); !ok {
			ps := strings.SplitN(device, ":", 2)
			if len(ps) != 2 {
				return errors.New("invalid device string")
			}

			rt := ps[0]

			if rt == "file" {
				r := file.NewReader(cfg)
				err := r.Open(device, iq)
				if err != nil {
					log.Error().Msgf("error opening file reader: %s", err)
					continue
				} else {
					st.SetReader(device, r)
					log.Info().Msgf("opened file reader: %s", device)
				}
			} else {
				r := libnfc.NewReader(cfg)
				err := r.Open(device, iq)
				if err != nil {
					log.Error().Msgf("error opening libnfc reader: %s", err)
					continue
				} else {
					st.SetReader(device, r)
					log.Info().Msgf("opened libnfc reader: %s", device)
				}
			}
		}
	}

	lrDetect := libnfc.NewReader(cfg)
	detectDevice := lrDetect.Detect(st.ListReaders())
	if detectDevice != "" {
		// log.Info().Msgf("detected new reader: %s", detectDevice)
		err := lrDetect.Open(detectDevice, iq)
		if err != nil {
			log.Error().Msgf("error opening detected reader: %s", err)
		}

		if lrDetect != nil {
			if lrDetect.Connected() {
				st.SetReader(detectDevice, lrDetect)
			} else {
				_ = lrDetect.Close()
			}
		}
	}

	if !utils.Contains(rs, "") {
		lrAny := libnfc.NewReader(cfg)
		err := lrAny.Open("", iq)
		if err == nil {
			if lrAny != nil && lrAny.Connected() {
				st.SetReader("", lrAny)
			} else if lrAny != nil {
				_ = lrAny.Close()
			}
		}
	}

	return nil
}

func readerManager(
	pl platforms.Platform,
	cfg *config.UserConfig,
	st *state.State,
	launchQueue *tokens.TokenQueue,
	softwareQueue chan *tokens.Token,
) {
	inputQueue := make(chan readers.Scan)

	var err error
	var lastError time.Time

	var prevToken *tokens.Token
	var softwareToken *tokens.Token
	var exitTimer *time.Timer

	readerTicker := time.NewTicker(1 * time.Second)
	stopService := make(chan bool)

	playFail := func() {
		if time.Since(lastError) > 1*time.Second {
			pl.PlayFailSound(cfg)
		}
	}

	startTimedExit := func() {
		if exitTimer != nil {
			stopped := exitTimer.Stop()
			if stopped {
				log.Info().Msg("cancelling previous exit timer")
			}
		}

		timerLen := time.Second * time.Duration(cfg.GetExitGameDelay())
		log.Debug().Msgf("exit timer set to: %s seconds", timerLen)
		exitTimer = time.NewTimer(timerLen)

		go func() {
			<-exitTimer.C

			if pl.GetActiveLauncher() == "" {
				log.Debug().Msg("no active launcher, not exiting")
				return
			}

			log.Info().Msg("exiting software")
			err := pl.KillLauncher()
			if err != nil {
				log.Warn().Msgf("error killing launcher: %s", err)
			}

			softwareQueue <- nil
		}()
	}

	// manage reader connections
	go func() {
		for {
			select {
			case <-stopService:
				return
			case <-readerTicker.C:
				readers := st.ListReaders()
				for _, device := range readers {
					r, ok := st.GetReader(device)
					if ok && r != nil && !r.Connected() {
						log.Debug().Msgf("pruning disconnected reader: %s", device)
						st.RemoveReader(device)
					}
				}

				err := connectReaders(cfg, st, inputQueue)
				if err != nil {
					log.Error().Msgf("error connecting readers: %s", err)
				}
			}
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
		case st := <-softwareQueue:
			// a token has been launched that starts software
			log.Debug().Msgf("new software token: %v", st)

			if exitTimer != nil && !cmpTokens(st, softwareToken) {
				if stopped := exitTimer.Stop(); stopped {
					log.Info().Msg("different software token inserted, cancelling exit")
				}
			}

			softwareToken = st
			continue
		}

		if cmpTokens(scan, prevToken) {
			log.Debug().Msg("ignoring duplicate scan")
			continue
		}

		prevToken = scan

		if scan != nil {
			log.Info().Msgf("new token scanned: %v", scan)
			if !st.IsLauncherDisabled() {
				if exitTimer != nil {
					stopped := exitTimer.Stop()
					if stopped && cmpTokens(scan, softwareToken) {
						log.Info().Msg("same token reinserted, cancelling exit")
						continue
					} else if stopped {
						log.Info().Msg("new token inserted, restarting exit timer")
						startTimedExit()
					}
				}

				log.Info().Msgf("sending token: %v", scan)
				pl.PlaySuccessSound(cfg)
				launchQueue.Enqueue(*scan)
			}
		} else {
			log.Info().Msg("token was removed")
			st.SetActiveCard(tokens.Token{})
			if shouldExit(cfg, pl, st) {
				startTimedExit()
			}
		}
	}

	// daemon shutdown
	stopService <- true
	readers := st.ListReaders()
	for _, device := range readers {
		r, ok := st.GetReader(device)
		if ok && r != nil {
			err := r.Close()
			if err != nil {
				log.Warn().Msg("error closing reader")
			}
		}
	}
}
