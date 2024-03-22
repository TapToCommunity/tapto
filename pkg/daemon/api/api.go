package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
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
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))
		r.Use(middleware.Timeout(60 * time.Second))

		r.Get("/status", handleStatus(st, tr))

		r.Post("/launch", handleLaunch(st, tq))
		r.Get("/launch/{rest:.*}", handleLaunchBasic(st, tq))

		// GET /readers/0/read
		r.Post("/readers/0/write", handleReaderWrite(st))

		r.Get("/games", handleGames())
		r.Get("/systems", handleSystems())

		r.Get("/mappings", handleMappings(db))
		r.Post("/mappings", handleAddMapping(db))

		r.Get("/history", handleHistory(db))

		// PUT /settings
		r.Get("/settings", handleSettings(cfg))
		r.Get("/settings/log", handleSettingsDownloadLog())
		r.Post("/settings/index/games", handleIndexGames(cfg))
	})

	// events SSE

	err := http.ListenAndServe(":7497", r) // TODO: move port to config
	if err != nil {
		log.Error().Msgf("error starting http server: %s", err)
	}
}
