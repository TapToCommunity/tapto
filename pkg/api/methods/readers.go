package methods

import (
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/api/models"
	"github.com/wizzomafizzo/tapto/pkg/api/models/requests"
)

func HandleReaderWrite(env requests.RequestEnv) (any, error) {
	log.Info().Msg("received reader write request")

	if len(env.Params) == 0 {
		return nil, ErrMissingParams
	}

	var params models.ReaderWriteParams
	err := json.Unmarshal(env.Params, &params)
	if err != nil {
		return nil, ErrInvalidParams
	}

	rs := env.State.ListReaders()
	if len(rs) == 0 {
		return nil, errors.New("no readers connected")
	}

	rid := rs[0]
	lt := env.State.GetLastScanned()

	if !lt.ScanTime.IsZero() && !lt.Remote {
		rid = lt.Source
	}

	reader, ok := env.State.GetReader(rid)
	if !ok || reader == nil {
		return nil, errors.New("reader not connected: " + rs[0])
	}

	t, err := reader.Write(params.Text)
	if err != nil {
		log.Error().Err(err).Msg("error writing to reader")
		return nil, errors.New("error writing to reader")
	}

	if t != nil {
		env.State.SetWroteToken(t)
	}

	return nil, nil
}
