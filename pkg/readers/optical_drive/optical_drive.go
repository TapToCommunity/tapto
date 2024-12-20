package optical_drive

import (
	"errors"
	"github.com/ZaparooProject/zaparoo-core/pkg/config"
	"github.com/ZaparooProject/zaparoo-core/pkg/readers"
	"github.com/ZaparooProject/zaparoo-core/pkg/service/tokens"
	"github.com/ZaparooProject/zaparoo-core/pkg/utils"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const TokenType = "disc"

type FileReader struct {
	cfg     *config.Instance
	device  string
	path    string
	polling bool
}

func NewReader(cfg *config.Instance) *FileReader {
	return &FileReader{
		cfg: cfg,
	}
}

func (r *FileReader) Ids() []string {
	return []string{"optical_drive"}
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

	r.device = device
	r.path = path
	r.polling = true

	go func() {
		var token *tokens.Token

		for r.polling {
			time.Sleep(1 * time.Second)

			rawUuid, err := exec.Command("blkid", "-o", "value", "-s", "UUID", r.path).Output()
			if err != nil {
				if token != nil {
					log.Debug().Err(err).Msg("error identifying optical media, removing token")
					token = nil
					iq <- readers.Scan{
						Source: r.device,
						Token:  nil,
					}
				} else {
					continue
				}
			}

			uuid := strings.TrimSpace(string(rawUuid))

			if uuid == "" && token != nil {
				log.Debug().Msg("id is empty, removing token")
				token = nil
				iq <- readers.Scan{
					Source: r.device,
					Token:  nil,
				}
				continue
			}

			if token != nil && token.UID == uuid {
				continue
			}

			if uuid == "" {
				continue
			}

			token = &tokens.Token{
				Type:     TokenType,
				ScanTime: time.Now(),
				UID:      uuid,
			}

			log.Debug().Msgf("new token: %s", token.UID)
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

func (r *FileReader) Write(text string) (*tokens.Token, error) {
	return nil, nil
}
