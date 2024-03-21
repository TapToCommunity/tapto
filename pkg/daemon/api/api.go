package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/database"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
)

func RunApiServer(
	cfg *config.UserConfig,
	st *state.State,
	tq *state.TokenQueue,
	db *database.Database,
	kbd input.Keyboard,
	tr *mister.Tracker,
) {
	r := mux.NewRouter()
	s := r.PathPrefix("/api/v1").Subrouter()

	s.Handle("/launch", handleLaunch(st, tq)).Methods(http.MethodPost)
	s.Handle("/launch/{rest:.*}", handleLaunchBasic(st, tq)).Methods(http.MethodGet)

	// GET /readers/0/read

	s.Handle("/readers/0/write", handleReaderWrite(st)).Methods(http.MethodPost)
	s.Handle("/games", handleGames()).Methods(http.MethodGet)
	s.Handle("/systems", handleSystems()).Methods(http.MethodGet)
	s.Handle("/mappings", handleMappings(db)).Methods(http.MethodGet)
	s.Handle("/mappings", handleAddMapping(db)).Methods(http.MethodPost)
	s.Handle("/history", handleHistory(db)).Methods(http.MethodGet)
	s.Handle("/status", handleStatus(st, tr)).Methods(http.MethodGet)
	s.Handle("/settings", handleSettings(cfg)).Methods(http.MethodGet)

	// PUT /settings

	s.Handle("/settings/log", handleSettingsLog()).Methods(http.MethodGet)
	s.Handle("/settings/index/games", handleIndexGames(cfg)).Methods(http.MethodPost)

	// events SSE

	http.Handle("/", r)

	err := http.ListenAndServe(":7497", nil) // TODO: move port to config
	if err != nil {
		log.Error().Msgf("error starting http server: %s", err)
	}
}
