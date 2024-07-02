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

		rs := st.ListReaders()
		if len(rs) == 0 {
			log.Error().Msg("no readers connected")
			http.Error(w, "no readers connected", http.StatusServiceUnavailable)
			return
		}

		// TODO: this just picks one at random for now

		reader, ok := st.GetReader(rs[0])
		if !ok || reader == nil {
			log.Error().Msg("reader not connected: " + rs[0])
			http.Error(w, "reader not connected", http.StatusServiceUnavailable)
			return
		}

		t, err := reader.Write(req.Text)
		if err != nil {
			log.Error().Err(err).Msg("error writing to reader")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if t != nil {
			st.SetWroteToken(t)
		}

		w.WriteHeader(http.StatusOK)
	}
}
