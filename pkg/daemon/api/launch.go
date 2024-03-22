package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
)

type LaunchRequestMetadata struct {
	ToyModel *string `json:"toyModel"`
}

type LaunchRequest struct {
	UID      string                 `json:"uid"`
	Text     string                 `json:"text"`
	Metadata *LaunchRequestMetadata `json:"metadata"`
}

func (lr *LaunchRequest) Bind(r *http.Request) error {
	return nil
}

func handleLaunch(
	st *state.State,
	tq *state.TokenQueue,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received launch request")

		var req LaunchRequest
		err := render.Bind(r, &req)
		if err != nil {
			log.Error().Msgf("error decoding request: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Info().Fields(req).Msgf("launching token")
		// TODO: how do we report back errors?

		t := state.Token{
			UID:      req.UID,
			Text:     req.Text,
			ScanTime: time.Now(),
		}

		st.SetActiveCard(t)
		tq.Enqueue(t)
	}
}

func handleLaunchBasic(
	st *state.State,
	tq *state.TokenQueue,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received basic launch request")

		text := chi.URLParam(r, "text")

		log.Info().Msgf("launching basic token: %s", text)

		t := state.Token{
			UID:      "",
			Text:     text,
			ScanTime: time.Now(),
		}

		st.SetActiveCard(t)
		tq.Enqueue(t)
	}
}
