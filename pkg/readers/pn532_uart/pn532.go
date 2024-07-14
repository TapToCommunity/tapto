package pn532_uart

import (
	"bytes"
	"errors"
	"time"

	"go.bug.st/serial"
)

const (
	CmdSamConfiguration    = 0x14
	CmdGetFirmwareVersion  = 0x02
	CmdGetGeneralStatus    = 0x04
	CmdInListPassiveTarget = 0x4A
	CmdInDataExchange      = 0x40
	HostToPn532            = 0xD4
	Pn532ToHost            = 0xD5
	Pn532Ready             = 0x01
)

var Ack = []byte{0x00, 0x00, 0xFF, 0x00, 0xFF, 0x00}

func NewFrame(cmd byte, data []byte) []byte {
	frm := []byte{0x00, 0x00, 0xFF}

	len := byte(len(data))
	frm = append(frm, len+2)
	frm = append(frm, ^len+1)
	frm = append(frm, HostToPn532)
	frm = append(frm, cmd)
	frm = append(frm, data...)

	sum := byte(0)
	for _, b := range frm[6:] {
		sum += b
	}

	frm = append(frm, ^sum+1)
	frm = append(frm, 0x00)

	return frm
}

func readAck(port serial.Port) error {
	buf := make([]byte, 6)
	_, err := port.Read(buf)
	if err != nil {
		return err
	}

	if !bytes.Equal(buf, Ack) {
		return errors.New("invalid ACK")
	}

	return nil
}

func wakeUp(port serial.Port) error {
	_, err := port.Write([]byte{0x55, 0x55, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0x03, 0xFD, 0xD4, 0x14, 0x01, 0x17, 0x00})
	if err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	return nil
}

func SamConfiguration(port serial.Port) error {
	err := wakeUp(port)
	if err != nil {
		return err
	}

	frm := NewFrame(CmdSamConfiguration, []byte{0x01, 0x14, 0x01})
	_, err = port.Write(frm)
	if err != nil {
		return err
	}

	return readAck(port)
}

func readFrame(port serial.Port) ([]byte, error) {
	buf := make([]byte, 255+7)
	_, err := port.Read(buf)
	if err != nil {
		return []byte{}, err
	}

	if len(buf) < 6 {
		return []byte{}, errors.New("response too short")
	}

	i := 0
	for ; i < len(buf); i++ {
		if buf[i] == 0xFF {
			break
		}
	}

	if i == len(buf) {
		return []byte{}, errors.New("no frame start found")
	}

	buf = buf[i:]
	if len(buf) < 7 {
		return []byte{}, errors.New("frame too short")
	}

	flen := int(buf[0])
	if len(buf) < flen+7 {
		return []byte{}, errors.New("length mismatch")
	}

	return buf, nil
}

func GetFirmwareVersion(port serial.Port) ([]byte, error) {
	err := wakeUp(port)
	if err != nil {
		return []byte{}, err
	}

	frm := NewFrame(CmdGetFirmwareVersion, nil)
	_, err = port.Write(frm)
	if err != nil {
		return []byte{}, err
	}

	buf := make([]byte, 64)
	_, err = port.Read(buf)
	if err != nil {
		return []byte{}, err
	}

	return buf, nil
}
