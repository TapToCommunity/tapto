package pn532_uart

import (
	"errors"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/readers"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
	"github.com/wizzomafizzo/tapto/pkg/utils"

	"go.bug.st/serial"
)

type SimpleSerialReader struct {
	cfg       *config.UserConfig
	device    string
	name      string
	polling   bool
	port      serial.Port
	lastToken *tokens.Token
}

func NewReader(cfg *config.UserConfig) *SimpleSerialReader {
	return &SimpleSerialReader{
		cfg: cfg,
	}
}

func (r *SimpleSerialReader) Ids() []string {
	return []string{"pn532_uart"}
}

func (r *SimpleSerialReader) Open(device string, iq chan<- readers.Scan) error {
	ps := strings.SplitN(device, ":", 2)
	if len(ps) != 2 {
		return errors.New("invalid device string: " + device)
	}

	if !utils.Contains(r.Ids(), ps[0]) {
		return errors.New("invalid reader id: " + ps[0])
	}

	name := ps[1]

	if runtime.GOOS != "windows" {
		if _, err := os.Stat(name); err != nil {
			return err
		}
	}

	port, err := serial.Open(name, &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	})
	if err != nil {
		return err
	}

	err = port.SetReadTimeout(100 * time.Millisecond)
	if err != nil {
		return err
	}

	err = SamConfiguration(port)
	if err != nil {
		return err
	}

	fv, err := GetFirmwareVersion(port)
	if err != nil {
		return err
	}
	log.Info().Msgf("firmware version: %v", fv)

	gs, err := GetGeneralStatus(port)
	if err != nil {
		return err
	}
	log.Info().Msgf("general status: %v", gs)

	r.port = port
	r.device = device
	r.name = name
	r.polling = true

	exit := func() {
		r.polling = false
		err = r.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close serial port")
		}
	}

	go func() {
		for r.polling {
			buf := make([]byte, 1024)
			_, err := r.port.Read(buf)
			if err != nil {
				log.Error().Err(err).Msg("failed to read from serial port")
				exit()
				break
			}

		}
	}()

	return nil
}

func (r *SimpleSerialReader) Close() error {
	r.polling = false
	if r.port != nil {
		err := r.port.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *SimpleSerialReader) Detect(connected []string) string {
	_, err := serial.GetPortsList()
	if err != nil {
		log.Error().Err(err).Msg("failed to get serial ports")
	}

	// log.Debug().Msgf("detected serial ports: %v", ports)

	return ""
}

func (r *SimpleSerialReader) Device() string {
	return r.device
}

func (r *SimpleSerialReader) Connected() bool {
	return r.polling && r.port != nil
}

func (r *SimpleSerialReader) Info() string {
	return "PN532 UART (" + r.name + ")"
}

func (r *SimpleSerialReader) Write(text string) (*tokens.Token, error) {
	return nil, errors.New("writing not supported on this reader")
}
