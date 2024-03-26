package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/daemon/api/websocket"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/database"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
	"github.com/wizzomafizzo/tapto/pkg/utils"
)

const (
	SubStreamStatus = "status"
	SubStreamEvents = "events"
)

func setupWs(st *state.State, tr *mister.Tracker) {
	send := func() {
		status, err := json.Marshal(newStatus(st, tr))
		if err != nil {
			log.Error().Err(err).Msg("error encoding status")
			return
		}

		websocket.Broadcast("STATUS " + string(status))
	}

	stHook := func(_ *state.State) {
		log.Debug().Msg("state update hook")
		send()
	}
	st.SetUpdateHook(&stHook)

	trHook := func(_ *mister.Tracker, _ int, _ string) {
		log.Debug().Msg("tracker update hook")
		send()
	}
	tr.SetEventHook(&trHook)

	idxHook := func(_ *Index) {
		log.Debug().Msg("index update hook")
		send()
	}
	IndexInstance.SetEventHook(&idxHook)

	// give the ws module a logger that doesn't include itself
	websocket.SetLogger(log.Output(io.MultiWriter(utils.BaseLogWriters...)))
	// change the global logger to include the ws writer
	writers := append(utils.BaseLogWriters, &websocket.LogWriter{})
	log.Logger = log.Output(io.MultiWriter(writers...))
}

func RunApiServer(
	cfg *config.UserConfig,
	st *state.State,
	tq *state.TokenQueue,
	db *database.Database,
	tr *mister.Tracker,
) {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

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

	setupWs(st, tr)
	r.HandleFunc("/api/v1/ws", websocket.Handle(
		func() []string {
			return []string{}
		},
		func(msg string) string {
			return ""
		},
	))

	err := http.ListenAndServe(":7497", r) // TODO: move port to config
	if err != nil {
		log.Error().Msgf("error starting http server: %s", err)
	}
}
