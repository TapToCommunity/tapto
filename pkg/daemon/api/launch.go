package api

import (
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

type LaunchRequestMetadata struct {
	ToyModel *string `json:"toyModel"`
}

type LaunchRequest struct {
	Type     string                 `json:"type"`
	UID      string                 `json:"uid"`
	Text     string                 `json:"text"`
	Data     string                 `json:"data"`
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

		t := tokens.Token{
			UID:      req.UID,
			Text:     req.Text,
			ScanTime: time.Now(),
			FromApi:  true,
			Type:     req.Type,
			Data:     req.Data,
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

		text := chi.URLParam(r, "*")
		text, err := url.QueryUnescape(text)
		if err != nil {
			log.Error().Msgf("error decoding request: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Info().Msgf("launching basic token: %s", text)

		t := tokens.Token{
			UID:      "__api__",
			Text:     text,
			ScanTime: time.Now(),
			FromApi:  true,
		}

		st.SetActiveCard(t)
		tq.Enqueue(t)
	}
}

func HandleStopGame(platform platforms.Platform) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received stop game request")

		err := platform.KillLauncher()
		if err != nil {
			log.Error().Msgf("error launching menu: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
