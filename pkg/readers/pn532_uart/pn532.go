package pn532_uart

import (
	"bytes"
	"errors"

	"go.bug.st/serial"
)

const (
	CmdSamConfiguration    = 0x14
	CmdGetFirmwareVersion  = 0x02
	CmdGetGeneralStatus    = 0x04
	CmdInListPassiveTarget = 0x4A
	CmdInDataExchange      = 0x40
	FrmPreamble            = 0x00
	FrmStart1              = 0x00
	FrmStart2              = 0xFF
	FrmPostamble           = 0x00
	HostToPn532            = 0xD4
	Pn532ToHost            = 0xD5
)

var Ack = []byte{0x00, 0x00, 0xFF, 0x00, 0xFF, 0x00}

func WriteCmd(port serial.Port, cmd []byte) error {
	cmdLen := len(cmd)
	if cmdLen < 1 {
		return errors.New("cmd is empty")
	}
	if cmdLen > 255 {
		return errors.New("cmd is too long")
	}

	len := byte(cmdLen + 1)
	tfi := byte(HostToPn532)
	dcs := byte(^len + 1)

	data := append([]byte{FrmPreamble, FrmStart1, FrmStart2, len, tfi}, cmd...)
	data = append(data, dcs, FrmPostamble)

	_, err := port.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func ReadAck(port serial.Port) error {
	buf := make([]byte, 6)
	_, err := port.Read(buf)
	if err != nil {
		return err
	}

	if !bytes.Equal(buf, Ack) {
		return errors.New("invalid ack")
	}

	return nil
}

func ReadResponse(port serial.Port) ([]byte, error) {
	buf := make([]byte, 6)
	_, err := port.Read(buf)
	if err != nil {
		return nil, err
	}

	if len(buf) < 6 {
		return nil, errors.New("invalid response")
	}

	if buf[0] != FrmPreamble || buf[1] != FrmStart1 || buf[2] != FrmStart2 {
		return nil, errors.New("invalid response")
	}

	len := buf[3]
	// tfi := buf[4]
	dcs := buf[5]

	if dcs != ^byte(len+1) {
		return nil, errors.New("invalid dcs")
	}

	data := make([]byte, len)
	_, err = port.Read(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func SendCmd(port serial.Port, cmd []byte) ([]byte, error) {
	err := WriteCmd(port, cmd)
	if err != nil {
		return nil, err
	}

	err = ReadAck(port)
	if err != nil {
		return nil, err
	}

	data, err := ReadResponse(port)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func GetFirmwareVersion(port serial.Port) ([]byte, error) {
	return SendCmd(port, []byte{CmdGetFirmwareVersion})
}

func GetGeneralStatus(port serial.Port) ([]byte, error) {
	return SendCmd(port, []byte{CmdGetGeneralStatus})
}

func SamConfiguration(port serial.Port) ([]byte, error) {
	return SendCmd(port, []byte{CmdSamConfiguration, 0x01, 0x14, 0x01})
}
