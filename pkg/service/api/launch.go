package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/service/state"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

type LaunchParams struct {
	Type string `json:"type"`
	UID  string `json:"uid"`
	Text string `json:"text"`
	Data string `json:"data"`
}

func handleLaunch(env RequestEnv) error {
	log.Info().Msg("received launch request")

	if len(env.Params) == 0 {
		return errors.New("missing params")
	}

	var params LaunchParams
	err := json.Unmarshal(env.Params, &params)
	if err != nil {
		return errors.New("invalid params: " + err.Error())
	}

	var t tokens.Token

	t.UID = params.UID
	t.Text = params.Text
	t.Type = params.Type
	t.Data = params.Data

	t.ScanTime = time.Now()
	t.Remote = true

	// TODO: how do we report back errors?
	env.State.SetActiveCard(t)
	env.TokenQueue.Enqueue(t)

	return nil
}

func handleLaunchBasic(
	st *state.State,
	tq *tokens.TokenQueue,
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
			Remote:   true,
		}

		st.SetActiveCard(t)
		tq.Enqueue(t)
	}
}

func handleStopGame(env RequestEnv) error {
	log.Info().Msg("received stop game request")
	return env.Platform.KillLauncher()
}
