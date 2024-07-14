package acr122pcsc

import (
	"errors"
	"strings"
	"time"

	"github.com/ebfe/scard"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/readers"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
	"github.com/wizzomafizzo/tapto/pkg/utils"
)

type Acr122Pcsc struct {
	cfg     *config.UserConfig
	device  string
	polling bool
	ctx     *scard.Context
}

func NewAcr122Pcsc(cfg *config.UserConfig) *Acr122Pcsc {
	return &Acr122Pcsc{
		cfg: cfg,
	}
}

func (r *Acr122Pcsc) Ids() []string {
	return []string{"acr122_pcsc"}
}

func (r *Acr122Pcsc) Open(device string, iq chan<- readers.Scan) error {
	ps := strings.SplitN(device, ":", 2)
	if len(ps) != 2 {
		return errors.New("invalid device string: " + device)
	}

	if !utils.Contains(r.Ids(), ps[0]) {
		return errors.New("invalid reader id: " + ps[0])
	}

	if r.ctx != nil {
		ctx, err := scard.EstablishContext()
		if err != nil {
			return err
		}
		r.ctx = ctx
	}

	r.device = device
	r.polling = true

	go func() {
		for r.polling {
			time.Sleep(100 * time.Millisecond)
		}
	}()

	return nil
}

func (r *Acr122Pcsc) Close() error {
	r.polling = false
	if r.ctx != nil {
		r.ctx.Release()
	}
	return nil
}

func (r *Acr122Pcsc) Detect(connected []string) string {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return ""
	}
	defer ctx.Release()

	readers, err := ctx.ListReaders()
	if err != nil {
		log.Debug().Msgf("error listing pcsc readers: %s", err)
		return ""
	}

	log.Debug().Msgf("detected pcsc readers: %v", readers)

	return ""
}

func (r *Acr122Pcsc) Device() string {
	return r.device
}

func (r *Acr122Pcsc) Connected() bool {
	return r.polling
}

func (r *Acr122Pcsc) Info() string {
	return ""
}

func (r *Acr122Pcsc) Write(text string) (*tokens.Token, error) {
	return nil, errors.New("writing not supported on this reader")
}
