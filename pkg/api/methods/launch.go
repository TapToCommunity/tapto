package methods

import (
	"encoding/json"
	"errors"
	"github.com/wizzomafizzo/tapto/pkg/api/models"
	"github.com/wizzomafizzo/tapto/pkg/api/models/requests"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/service/state"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

var (
	ErrMissingParams = errors.New("missing params")
	ErrInvalidParams = errors.New("invalid params")
	ErrNotAllowed    = errors.New("not allowed")
)

func HandleLaunch(env requests.RequestEnv) (any, error) {
	log.Info().Msg("received launch request")

	if len(env.Params) == 0 {
		return nil, ErrMissingParams
	}

	var params models.LaunchParams
	err := json.Unmarshal(env.Params, &params)
	if err != nil {
		return nil, ErrInvalidParams
	}

	var t tokens.Token

	t.UID = params.UID
	t.Text = params.Text
	t.Type = params.Type
	t.Data = params.Data

	t.ScanTime = time.Now()
	t.Remote = true // TODO: check if this is still necessary after api update

	// TODO: how do we report back errors?
	env.State.SetActiveCard(t)
	env.TokenQueue.Enqueue(t)

	return nil, nil
}

// TODO: this is still insecure
func HandleLaunchBasic(
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
			Text:     text,
			ScanTime: time.Now(),
			Remote:   true,
		}

		st.SetActiveCard(t)
		tq.Enqueue(t)
	}
}

func HandleStop(env requests.RequestEnv) (any, error) {
	log.Info().Msg("received stop request")
	return nil, env.Platform.KillLauncher()
}
