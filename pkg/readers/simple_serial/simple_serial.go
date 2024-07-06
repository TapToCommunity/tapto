package simple_serial

import (
	"errors"
	"os"
	"path/filepath"
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
	path      string
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
	return []string{"simple_serial"}
}

func (r *SimpleSerialReader) parseLine(line string) (*tokens.Token, error) {
	line = strings.TrimSpace(line)
	line = strings.Trim(line, "\r")

	if len(line) == 0 {
		return nil, nil
	}

	if !strings.HasPrefix(line, "SCAN\t") {
		return nil, nil
	}

	args := line[5:]
	if len(args) == 0 {
		return nil, nil
	}

	t := tokens.Token{
		Data:     line,
		ScanTime: time.Now(),
		Source:   r.device,
	}

	ps := strings.Split(args, "\t")
	for i := 0; i < len(ps); i++ {
		ps[i] = strings.TrimSpace(ps[i])
		if strings.HasPrefix(ps[i], "uid=") {
			t.UID = ps[i][4:]
		} else if strings.HasPrefix(ps[i], "text=") {
			t.Text = ps[i][5:]
		}
	}

	return &t, nil
}

func (r *SimpleSerialReader) Open(device string, iq chan<- readers.Scan) error {
	ps := strings.SplitN(device, ":", 2)
	if len(ps) != 2 {
		return errors.New("invalid device string: " + device)
	}

	if !utils.Contains(r.Ids(), ps[0]) {
		return errors.New("invalid reader id: " + ps[0])
	}

	path := ps[1]

	if !filepath.IsAbs(path) {
		return errors.New("invalid device path, must be absolute")
	}

	if _, err := os.Stat(path); err != nil {
		return err
	}

	port, err := serial.Open(path, &serial.Mode{
		BaudRate: 115200,
	})
	if err != nil {
		return err
	}

	err = port.SetReadTimeout(100 * time.Millisecond)
	if err != nil {
		return err
	}

	r.port = port
	r.device = device
	r.path = path
	r.polling = true

	go func() {
		var lineBuf []byte

		for r.polling {
			time.Sleep(100 * time.Millisecond)

			buf := make([]byte, 4096)
			n, err := r.port.Read(buf)
			if err != nil {
				log.Error().Err(err).Msg("failed to read from serial port")
				err = r.Close()
				if err != nil {
					log.Error().Err(err).Msg("failed to close serial port")
				}
				break
			}

			for i := 0; i < n; i++ {
				if buf[i] == '\n' {
					line := string(lineBuf)
					lineBuf = nil

					t, err := r.parseLine(line)
					if err != nil {
						log.Error().Err(err).Msg("failed to parse line")
						continue
					}

					if t != nil && !utils.TokensEqual(t, r.lastToken) {
						iq <- readers.Scan{
							Source: r.device,
							Token:  t,
						}
					}

					if t != nil {
						r.lastToken = t
					}
				} else {
					lineBuf = append(lineBuf, buf[i])
				}
			}

			if r.lastToken != nil && time.Since(r.lastToken.ScanTime) > 1*time.Second {
				iq <- readers.Scan{
					Source: r.device,
					Token:  nil,
				}
				r.lastToken = nil
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
	return ""
}

func (r *SimpleSerialReader) Device() string {
	return r.device
}

func (r *SimpleSerialReader) Connected() bool {
	return r.polling && r.port != nil
}

func (r *SimpleSerialReader) Info() string {
	return r.path
}

func (r *SimpleSerialReader) Write(text string) (*tokens.Token, error) {
	return nil, errors.New("writing not supported on this reader")
}
