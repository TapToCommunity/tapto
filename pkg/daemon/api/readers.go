package api

import (
	"errors"
	"net/http"
	"time"

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

		st.SetWriteRequest(req.Text)

		for st.GetWriteRequest() != "" {
			time.Sleep(100 * time.Millisecond)
		}

		if st.GetWriteError() != nil {
			http.Error(w, st.GetWriteError().Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
