//go:build linux && cgo

package libnfc

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/clausecker/nfc/v2"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/readers"
	"github.com/wizzomafizzo/tapto/pkg/readers/libnfc/tags"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
	"github.com/wizzomafizzo/tapto/pkg/utils"
)

const (
	timeToForgetCard   = 500 * time.Millisecond
	connectMaxTries    = 10
	timesToPoll        = 1
	periodBetweenPolls = 250 * time.Millisecond
	periodBetweenLoop  = 250 * time.Millisecond
)

type WriteRequestResult struct {
	Token *tokens.Token
	Err   error
}

type WriteRequest struct {
	Text   string
	Result chan WriteRequestResult
}

type Reader struct {
	cfg       *config.UserConfig
	conn      string
	pnd       *nfc.Device
	polling   bool
	prevToken *tokens.Token
	write     chan WriteRequest
}

func NewReader(cfg *config.UserConfig) *Reader {
	return &Reader{
		cfg:   cfg,
		write: make(chan WriteRequest),
	}
}

func (r *Reader) Open(device string, iq chan<- readers.Scan) error {
	pnd, err := openDeviceWithRetries(device)
	if err != nil {
		return err
	}

	r.conn = device
	r.pnd = &pnd
	r.polling = true
	r.prevToken = nil

	go func() {
		for r.polling {
			select {
			case req := <-r.write:
				r.writeTag(req)
			case <-time.After(periodBetweenLoop):
				// continue with reading
			}

			token, removed, err := r.pollDevice(r.pnd, r.prevToken, timesToPoll, periodBetweenPolls)
			if errors.Is(err, nfc.Error(nfc.EIO)) {
				log.Error().Msgf("error during poll: %s", err)
				log.Error().Msg("fatal IO error, device was possibly unplugged")

				err = r.Close()
				if err != nil {
					log.Warn().Msgf("error closing device: %s", err)
				}

				continue
			} else if err != nil {
				log.Error().Msgf("error polling device: %s", err)
				continue
			}

			if removed {
				log.Info().Msg("token removed, sending to input queue")
				iq <- readers.Scan{
					Source: r.conn,
					Token:  nil,
				}
				r.prevToken = nil
			} else if token != nil {
				if r.prevToken != nil && token.UID == r.prevToken.UID {
					continue
				}

				log.Info().Msg("new token detected, sending to input queue")
				iq <- readers.Scan{
					Source: r.conn,
					Token:  token,
				}
				r.prevToken = token
			}
		}
	}()

	return nil
}

func (r *Reader) Close() error {
	r.polling = false

	if r.pnd == nil {
		return nil
	} else {
		return r.pnd.Close()
	}
}

func (r *Reader) Ids() []string {
	return []string{
		"pn532_uart",
		"pn532_i2c",
		"acr122_usb",
		"pcsc",
	}
}

func (r *Reader) Detect(connected []string) string {
	if !r.cfg.GetProbeDevice() {
		// log.Debug().Msg("device probing disabled")
		return ""
	}

	device := detectSerialReaders(connected)
	if device == "" {
		// log.Debug().Msg("no serial nfc reader detected")
		return ""
	}

	if utils.Contains(connected, device) {
		// log.Debug().Msgf("already connected to: %s", device)
		return ""
	}

	log.Info().Msgf("detected nfc reader: %s", device)

	return device
}

func (r *Reader) Device() string {
	return r.conn
}

func (r *Reader) Connected() bool {
	return r.pnd != nil && r.pnd.Connection() != ""
}

func (r *Reader) Info() string {
	if !r.Connected() {
		return ""
	}

	return r.pnd.String()
}

func (r *Reader) Write(text string) (*tokens.Token, error) {
	if !r.Connected() {
		return nil, errors.New("not connected")
	}

	req := WriteRequest{
		Text:   text,
		Result: make(chan WriteRequestResult),
	}

	r.write <- req

	res := <-req.Result
	if res.Err != nil {
		log.Error().Msgf("error writing to tag: %s", res.Err)
		return nil, res.Err
	}

	return res.Token, nil
}

// keep track of serial devices that had failed opens
var serialCacheMu = &sync.RWMutex{}
var serialBlockList = []string{}

func detectSerialReaders(connected []string) string {
	devices, err := utils.GetLinuxSerialDeviceList()
	if err != nil {
		log.Error().Msgf("error getting serial devices: %s", err)
		return ""
	}

	for _, device := range devices {
		// the libnfc open is extremely disruptive to other devices, we want
		// to minimise the number of times we try to open a device
		connStr := "pn532_uart:" + device

		// ignore if device is in block list
		serialCacheMu.RLock()
		if utils.Contains(serialBlockList, device) {
			serialCacheMu.RUnlock()
			continue
		}
		serialCacheMu.RUnlock()

		// ignore if exact same device and reader are connected
		if utils.Contains(connected, connStr) {
			continue
		}

		// resolve device symlink if necessary
		realPath := ""
		symPath, err := os.Readlink(device)
		if err == nil {
			parent := filepath.Dir(device)
			abs, err := filepath.Abs(filepath.Join(parent, symPath))
			if err == nil {
				realPath = abs
			}
		}

		// ignore if same resolved device and reader connected
		if realPath != "" && utils.Contains(connected, realPath) {
			continue
		}

		// ignore if different reader already connected
		if strings.HasSuffix(device, ":"+device) {
			continue
		}

		// ignore if different resolved device and reader connected
		if realPath != "" && strings.HasSuffix(realPath, ":"+realPath) {
			continue
		}

		pnd, err := nfc.Open(connStr)
		if err != nil {
			serialCacheMu.Lock()
			serialBlockList = append(serialBlockList, device)
			serialCacheMu.Unlock()
		} else {
			pnd.Close()
			return connStr
		}
	}

	return ""
}

func openDeviceWithRetries(device string) (nfc.Device, error) {
	tries := 0
	for {
		pnd, err := nfc.Open(device)
		if err == nil {
			log.Info().Msgf("successful connect, after %d tries", tries)

			connProto := strings.SplitN(strings.ToLower(device), ":", 2)[0]
			log.Info().Msgf("connection protocol: %s", connProto)
			deviceName := pnd.String()
			log.Info().Msgf("device name: %s", deviceName)

			if err := pnd.InitiatorInit(); err != nil {
				log.Error().Msgf("could not init initiator: %s", err)
				continue
			}

			return pnd, err
		}

		if tries >= connectMaxTries {
			// log.Debug().Msgf("could not open device after %d tries: %s", connectMaxTries, err)
			return pnd, err
		}

		tries++
	}
}

func (r *Reader) pollDevice(
	pnd *nfc.Device,
	activeToken *tokens.Token,
	ttp int,
	pbp time.Duration,
) (*tokens.Token, bool, error) {
	removed := false

	count, target, err := pnd.InitiatorPollTarget(tags.SupportedCardTypes, ttp, pbp)
	if err != nil && !errors.Is(err, nfc.Error(nfc.ETIMEOUT)) {
		return nil, removed, err
	}

	if count > 1 {
		log.Info().Msg("more than one card on the reader")
	}

	if count <= 0 {
		if activeToken != nil && time.Since(activeToken.ScanTime) > timeToForgetCard {
			log.Info().Msg("card removed")
			activeToken = nil
			removed = true
		}

		return activeToken, removed, nil
	}

	tagUid := tags.GetTagUID(target)
	if tagUid == "" {
		log.Warn().Msgf("unable to detect token UID: %s", target.String())
	}

	// no change in tag
	if activeToken != nil && tagUid == activeToken.UID {
		return activeToken, removed, nil
	}

	log.Info().Msgf("found token UID: %s", tagUid)

	var record tags.TagData
	cardType := tags.GetTagType(target)

	if cardType == tokens.TypeNTAG {
		log.Info().Msg("NTAG detected")
		record, err = tags.ReadNtag(*pnd)
		if err != nil {
			return activeToken, removed, fmt.Errorf("error reading ntag: %s", err)
		}
		cardType = tokens.TypeNTAG
	} else if cardType == tokens.TypeMifare {
		log.Info().Msg("MIFARE detected")
		record, err = tags.ReadMifare(*pnd, tagUid)
		if err != nil {
			log.Error().Msgf("error reading mifare: %s", err)
		}
		cardType = tokens.TypeMifare
	}

	log.Debug().Msgf("record bytes: %s", hex.EncodeToString(record.Bytes))
	tagText, err := tags.ParseRecordText(record.Bytes)
	if err != nil {
		log.Error().Err(err).Msgf("error parsing NDEF record")
		tagText = ""
	}

	if tagText == "" {
		log.Warn().Msg("no text NDEF found")
	} else {
		log.Info().Msgf("decoded text NDEF: %s", tagText)
	}

	card := &tokens.Token{
		Type:     record.Type,
		UID:      tagUid,
		Text:     tagText,
		Data:     hex.EncodeToString(record.Bytes),
		ScanTime: time.Now(),
		Source:   r.conn,
	}

	return card, removed, nil
}

func (r *Reader) writeTag(req WriteRequest) {
	log.Info().Msgf("libnfc write request: %s", req.Text)

	var count int
	var target nfc.Target
	var err error
	tries := 4 * 30 // ~30 seconds

	for tries > 0 {
		count, target, err = r.pnd.InitiatorPollTarget(
			tags.SupportedCardTypes,
			timesToPoll,
			periodBetweenPolls,
		)

		if err != nil && err.Error() != "timeout" {
			log.Error().Msgf("could not poll: %s", err)
		}

		if count > 0 {
			break
		}

		tries--
	}

	if count == 0 {
		log.Error().Msgf("could not detect a tag")
		req.Result <- WriteRequestResult{
			Err: errors.New("could not detect a tag"),
		}
		return
	}

	cardUid := tags.GetTagUID(target)
	log.Info().Msgf("found tag with UID: %s", cardUid)

	cardType := tags.GetTagType(target)
	var bytesWritten []byte

	switch cardType {
	case tokens.TypeMifare:
		bytesWritten, err = tags.WriteMifare(*r.pnd, req.Text, cardUid)
		if err != nil {
			log.Error().Msgf("error writing to mifare: %s", err)
			req.Result <- WriteRequestResult{
				Err: err,
			}
			return
		}
	case tokens.TypeNTAG:
		bytesWritten, err = tags.WriteNtag(*r.pnd, req.Text)
		if err != nil {
			log.Error().Msgf("error writing to ntag: %s", err)
			req.Result <- WriteRequestResult{
				Err: err,
			}
			return
		}
	default:
		log.Error().Msgf("unsupported tag type: %s", cardType)
		req.Result <- WriteRequestResult{
			Err: err,
		}
		return
	}

	t, _, err := r.pollDevice(r.pnd, nil, timesToPoll, periodBetweenPolls)
	if err != nil || t == nil {
		log.Error().Msgf("error reading written tag: %s", err)
		req.Result <- WriteRequestResult{
			Err: err,
		}
		return
	}

	if t.UID != cardUid {
		log.Error().Msgf("UID mismatch after write: %s != %s", t.UID, cardUid)
		req.Result <- WriteRequestResult{
			Err: errors.New("UID mismatch after write"),
		}
		return
	}

	if t.Text != req.Text {
		log.Error().Msgf("text mismatch after write: %s != %s", t.Text, req.Text)
		req.Result <- WriteRequestResult{
			Err: errors.New("text mismatch after write"),
		}
		return
	}

	log.Info().Msgf("successfully wrote to card: %s", hex.EncodeToString(bytesWritten))
	req.Result <- WriteRequestResult{
		Token: t,
	}
}
