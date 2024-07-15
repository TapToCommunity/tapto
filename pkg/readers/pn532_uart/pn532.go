package pn532_uart

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"go.bug.st/serial"
)

const (
	cmdSamConfiguration    = 0x14
	cmdGetFirmwareVersion  = 0x02
	cmdGetGeneralStatus    = 0x04
	cmdInListPassiveTarget = 0x4A
	cmdInDataExchange      = 0x40
	hostToPn532            = 0xD4
	pn532ToHost            = 0xD5
	pn532Ready             = 0x01
)

func wakeUp(port serial.Port) error {
	// over uart, pn532 must be "woken up" by sending 2 x 0x55 and then "waiting a while"
	// we send a bunch of 0x00 to wait
	_, err := port.Write([]byte{0x55, 0x55, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		return err
	}

	return nil
}

func waitAck(port serial.Port) error {
	// pn532 will send this sequence to acknowledge it received the previous command
	start := time.Now()

	buf := make([]byte, 6)
	for {
		_, err := port.Read(buf)
		if err != nil {
			return err
		}

		if bytes.Equal(buf, []byte{0x00, 0x00, 0xFF, 0x00, 0xFF, 0x00}) {
			return nil
		} else {
			if time.Since(start) > 1*time.Second {
				return errors.New("timeout waiting for ACK")
			} else {
				continue
			}
		}
	}
}

func sendFrame(port serial.Port, cmd byte, args []byte) error {
	// create frame
	frm := []byte{0x00, 0x00, 0xFF} // preamble and start code

	data := []byte{hostToPn532, cmd}
	data = append(data, args...)

	if len(data) > 255 {
		// TODO: extended frames are not implemented
		return errors.New("data too big for frame")
	}

	dlen := byte(len(data))
	frm = append(frm, dlen)    // length
	frm = append(frm, ^dlen+1) // length checksum

	checksum := byte(0)

	for _, b := range data {
		frm = append(frm, b)
		checksum += b
	}

	frm = append(frm, ^checksum+1) // data checksum
	frm = append(frm, 0x00)        // postamble

	log.Debug().Msgf("sending frame: %x", frm)

	// write frame
	err := wakeUp(port)
	if err != nil {
		return err
	}

	n, err := port.Write(frm)
	if err != nil {
		return err
	} else if n != len(frm) {
		return errors.New("write error, not all bytes written")
	}

	time.Sleep(2 * time.Millisecond)

	return waitAck(port)
}

func receiveFrame(port serial.Port) ([]byte, error) {
	buf := make([]byte, 255+7)
	_, err := port.Read(buf)
	if err != nil {
		return []byte{}, err
	}

	// find middle of packet code (0x00 0xff) and skip preamble
	off := 0
	for ; off < len(buf); off++ {
		if buf[off] == 0xFF {
			break
		}
	}
	if off == len(buf) {
		return []byte{}, errors.New("no frame found")
	}

	// check frame length value and checksum (LEN)
	off++
	frameLen := int(buf[off])
	if ((frameLen + int(buf[off+1])) & 0xFF) != 0 {
		return []byte{}, errors.New("invalid frame length")
	}

	// check frame checksum against data (LCS)
	chk := byte(0)
	for _, b := range buf[off+2 : off+2+frameLen+1] {
		chk += b
	}
	if chk != 0 {
		return []byte{}, errors.New("invalid frame checksum")
	}

	// check tfi
	off += 2
	if buf[off] != pn532ToHost {
		return []byte{}, errors.New("invalid TFI, expected PN532 to host")
	}

	// get frame data
	off++
	log.Debug().Msgf("received frame: %x", buf[off:off+frameLen-1])

	// return data part of frame
	data := make([]byte, frameLen-1)
	copy(data, buf[off:off+frameLen-1])

	return data, nil
}

func callCommand(
	port serial.Port,
	cmd byte,
	data []byte,
) ([]byte, error) {
	err := sendFrame(port, cmd, data)
	if err != nil {
		return []byte{}, err
	}

	res, err := receiveFrame(port)
	if err != nil {
		return []byte{}, err
	}

	return res, nil
}

func SamConfiguration(port serial.Port) error {
	// sets pn532 to "normal" mode
	res, err := callCommand(port, cmdSamConfiguration, []byte{0x01, 0x14, 0x01})
	if err != nil {
		return err
	} else if len(res) != 1 || res[0] != 0x15 {
		return errors.New("unexpected sam configuration response")
	}

	return nil
}

type FirmwareVersion struct {
	Version          string
	SupportIso14443a bool
	SupportIso14443b bool
	SupportIso18092  bool
}

func GetFirmwareVersion(port serial.Port) (FirmwareVersion, error) {
	res, err := callCommand(port, cmdGetFirmwareVersion, []byte{})
	if err != nil {
		return FirmwareVersion{}, err
	} else if len(res) != 5 || res[0] != 0x03 {
		return FirmwareVersion{}, errors.New("unexpected firmware version response")
	}

	if res[1] != 0x32 {
		return FirmwareVersion{}, fmt.Errorf("unexpected IC: %x", res[1])
	}

	fv := FirmwareVersion{
		Version:          fmt.Sprintf("%d.%d", res[2], res[3]),
		SupportIso14443a: res[4]&0x01 == 0x01,
		SupportIso14443b: res[4]&0x02 == 0x02,
		SupportIso18092:  res[4]&0x04 == 0x04,
	}

	return fv, nil
}

type GeneralStatus struct {
	LastError    byte
	FieldPresent bool
}

func GetGeneralStatus(port serial.Port) (GeneralStatus, error) {
	res, err := callCommand(port, cmdGetGeneralStatus, []byte{})
	if err != nil {
		return GeneralStatus{}, err
	} else if len(res) < 4 || res[0] != 0x05 {
		return GeneralStatus{}, errors.New("unexpected general status response")
	}

	gs := GeneralStatus{
		LastError:    res[1],
		FieldPresent: res[2] == 0x01,
	}

	return gs, nil
}
