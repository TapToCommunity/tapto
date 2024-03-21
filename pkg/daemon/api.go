package daemon

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/database"
)

type LaunchRequestMetadata struct {
	ToyModel *string `json:"toyModel"`
}

type LaunchRequest struct {
	UID      string                 `json:"uid"`
	Text     string                 `json:"text"`
	Metadata *LaunchRequestMetadata `json:"metadata"`
}

func handleLaunch(
	cfg *config.UserConfig,
	state *State,
	tq *TokenQueue,
	kbd input.Keyboard,
) http.HandlerFunc {
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
		// TODO: how do we report back errors?

		t := Token{
			UID:      req.UID,
			Text:     req.Text,
			ScanTime: time.Now(),
		}

		state.SetActiveCard(t)
		tq.Enqueue(t)
	}
}

func handleLaunchBasic(
	cfg *config.UserConfig,
	state *State,
	tq *TokenQueue,
	kbd input.Keyboard,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received basic launch request")

		vars := mux.Vars(r)
		text := vars["rest"]

		log.Info().Msgf("launching basic token: %s", text)

		t := Token{
			UID:      "",
			Text:     text,
			ScanTime: time.Now(),
		}

		state.SetActiveCard(t)
		tq.Enqueue(t)
	}
}

type HistoryReponseEntry struct {
	Time    time.Time `json:"time"`
	UID     string    `json:"uid"`
	Text    string    `json:"text"`
	Success bool      `json:"success"`
}

type HistoryResponse struct {
	Entries []HistoryReponseEntry `json:"entries"`
}

func handleHistory(
	db *database.Database,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received history request")

		entries, err := db.GetHistory()
		if err != nil {
			log.Error().Err(err).Msgf("error getting history")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := HistoryResponse{
			Entries: make([]HistoryReponseEntry, len(entries)),
		}

		for i, e := range entries {
			resp.Entries[i] = HistoryReponseEntry{
				Time:    e.Time,
				UID:     e.UID,
				Text:    e.Text,
				Success: e.Success,
			}
		}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Error().Err(err).Msgf("error encoding history response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func runApiServer(
	cfg *config.UserConfig,
	state *State,
	tq *TokenQueue,
	db *database.Database,
	kbd input.Keyboard,
) {
	r := mux.NewRouter()
	s := r.PathPrefix("/api/v1").Subrouter()

	s.Handle("/launch", handleLaunch(cfg, state, tq, kbd)).Methods(http.MethodPost)
	s.Handle("/launch/{rest:.*}", handleLaunchBasic(cfg, state, tq, kbd)).Methods(http.MethodGet)

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

	s.Handle("/history", handleHistory(db)).Methods(http.MethodGet)

	// GET /status

	// GET /settings

	// GET /settings/log

	http.Handle("/", r)

	err := http.ListenAndServe(":7497", nil) // TODO: move port to config
	if err != nil {
		log.Error().Msgf("error starting http server: %s", err)
	}
}
