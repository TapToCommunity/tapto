package daemon

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/tapto/pkg/config"
)

type LaunchRequestMetadata struct {
	ToyModel *string `json:"toyModel"`
}

type LaunchRequest struct {
	UID      string                 `json:"uid"`
	Text     string                 `json:"text"`
	Metadata *LaunchRequestMetadata `json:"metadata"`
}

func handleLaunch(cfg *config.UserConfig, state *State, kbd input.Keyboard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received launch request")

		var req LaunchRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error().Msgf("error decoding request: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Info().Fields(req).Msgf("launching token")
		state.SetActiveCard(Token{
			UID:      req.UID,
			Text:     req.Text,
			ScanTime: time.Now(),
		})

		// TODO: is this necessary? does the poll loop handle it?
		err = launchCard(cfg, state, kbd)
		if err != nil {
			log.Error().Msgf("error launching token: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func handleLaunchBasic(cfg *config.UserConfig, state *State, kbd input.Keyboard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received basic launch request")

		vars := mux.Vars(r)
		text := vars["rest"]

		log.Info().Msgf("launching basic token: %s", text)

		state.SetActiveCard(Token{
			UID:      "",
			Text:     text,
			ScanTime: time.Now(),
		})

		// TODO: is this necessary? does the poll loop handle it?
		err := launchCard(cfg, state, kbd)
		if err != nil {
			log.Error().Msgf("error launching basic token: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func runApiServer(cfg *config.UserConfig, state *State, kbd input.Keyboard) {
	r := mux.NewRouter()
	s := r.PathPrefix("/api/v1").Subrouter()

	s.Handle("/launch", handleLaunch(cfg, state, kbd)).Methods(http.MethodPost)
	s.Handle("/launch/{rest:.*}", handleLaunchBasic(cfg, state, kbd)).Methods(http.MethodGet)

	// GET /readers
	// Return all attached NFC readers
	// GET /readers/{id}
	// Return information about a specific reader
	// GET /readers/{id}/read
	// Blocks until a token is read, then returns the token data or times out
	// POST /readers/{id}/write
	// Attempt to write text to a token, blocks until the operation is complete or times out

	// GET /games
	// Search games

	// POST /index/games
	// Regenerate the games index

	// GET /mappings
	// Return all current mappings, or filter based on query parameters
	// POST /mappings
	// Create a new mapping

	// GET /history
	// Return all scans

	http.Handle("/", r)

	err := http.ListenAndServe(":7497", nil) // TODO: move port to config
	if err != nil {
		log.Error().Msgf("error starting http server: %s", err)
	}
}
