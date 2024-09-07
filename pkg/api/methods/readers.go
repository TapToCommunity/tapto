package methods

import (
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/api"
)

type ReaderWriteParams struct {
	Text string `json:"text"`
}

func HandleReaderWrite(env api.RequestEnv) error {
	log.Info().Msg("received reader write request")

	if len(env.Params) == 0 {
		return errors.New("missing params")
	}

	var params ReaderWriteParams
	err := json.Unmarshal(env.Params, &params)
	if err != nil {
		return errors.New("invalid params: " + err.Error())
	}

	rs := env.State.ListReaders()
	if len(rs) == 0 {
		return errors.New("no readers connected")
	}

	rid := rs[0]
	lt := env.State.GetLastScanned()

	if !lt.ScanTime.IsZero() && !lt.Remote {
		rid = lt.Source
	}

	reader, ok := env.State.GetReader(rid)
	if !ok || reader == nil {
		return errors.New("reader not connected: " + rs[0])
	}

	t, err := reader.Write(params.Text)
	if err != nil {
		return errors.New("error writing to reader")
	}

	if t != nil {
		env.State.SetWroteToken(t)
	}

	return env.SendResponse(env.Id, true)
}
