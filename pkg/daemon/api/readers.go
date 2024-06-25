package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
)

type ReaderWriteRequest struct {
	Text string `json:"text"`
}

func (rwr *ReaderWriteRequest) Bind(r *http.Request) error {
	if rwr.Text == "" {
		return errors.New("missing text")
	}

	return nil
}

func handleReaderWrite(st *state.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received reader write request")

		var req ReaderWriteRequest
		err := render.Bind(r, &req)
		if err != nil {
			log.Error().Err(err).Msg("error decoding request")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		reader := st.GetReader()

		if reader == nil {
			log.Error().Msg("no reader connected")
			http.Error(w, "no reader connected", http.StatusServiceUnavailable)
			return
		}

		err = reader.Write(req.Text)
		if err != nil {
			log.Error().Err(err).Msg("error writing to reader")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
