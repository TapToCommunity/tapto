package file

import (
	"encoding/hex"
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
)

const TokenType = "file"

type FileReader struct {
	cfg     *config.UserConfig
	device  string
	path    string
	polling bool
}

func NewReader(cfg *config.UserConfig) *FileReader {
	return &FileReader{
		cfg: cfg,
	}
}

func (r *FileReader) Ids() []string {
	return []string{"file"}
}

func (r *FileReader) Open(device string, iq chan<- readers.Scan) error {
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

	parent := filepath.Dir(path)
	if parent == "" {
		return errors.New("invalid device path")
	}

	if _, err := os.Stat(parent); err != nil {
		return err
	}

	if _, err := os.Stat(path); err != nil {
		// attempt to create empty file
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		_ = f.Close()
	}

	r.device = device
	r.path = path
	r.polling = true

	go func() {
		var token *tokens.Token

		for r.polling {
			time.Sleep(100 * time.Millisecond)

			contents, err := os.ReadFile(r.path)
			if err != nil {
				// TODO: have a max retries?
				iq <- readers.Scan{
					Source: r.device,
					Error:  err,
				}
				continue
			}

			text := strings.TrimSpace(string(contents))

			// "remove" the token if the file is now empty
			if text == "" && token != nil {
				log.Debug().Msg("file is empty, removing token")
				token = nil
				iq <- readers.Scan{
					Source: r.device,
					Token:  nil,
				}
				continue
			}

			if token != nil && token.Text == text {
				continue
			}

			if text == "" {
				continue
			}

			token = &tokens.Token{
				Type:     TokenType,
				Text:     text,
				Data:     hex.EncodeToString(contents),
				ScanTime: time.Now(),
			}

			log.Debug().Msgf("new token: %s", token.Text)
			iq <- readers.Scan{
				Source: r.device,
				Token:  token,
			}
		}
	}()

	return nil
}

func (r *FileReader) Close() error {
	r.polling = false
	return nil
}

func (r *FileReader) Detect(connected []string) string {
	return ""
}

func (r *FileReader) Device() string {
	return r.device
}

func (r *FileReader) Connected() bool {
	return r.polling
}

func (r *FileReader) Info() string {
	return r.path
}

func (r *FileReader) Write(text string) error {
	return errors.New("writing not supported on this reader")
}
