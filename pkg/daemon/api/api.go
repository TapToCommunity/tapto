package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/r3labs/sse/v2"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/database"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
)

const (
	SubStreamStatus = "status"
	SubStreamEvents = "events"
)

func setupSubscribe(sub *sse.Server, st *state.State, tr *mister.Tracker) {
	sub.CreateStream(SubStreamStatus)

	send := func() {
		status, err := json.Marshal(newStatus(st, tr))
		if err != nil {
			log.Error().Err(err).Msg("error encoding status")
			return
		}

		sub.Publish(SubStreamStatus, &sse.Event{
			Data: status,
		})
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
}

func RunApiServer(
	cfg *config.UserConfig,
	st *state.State,
	tq *state.TokenQueue,
	db *database.Database,
	tr *mister.Tracker,
) {
	sub := sse.New()
	// sub.AutoReplay = false
	sub.EventTTL = 1 * time.Second

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

	setupSubscribe(sub, st, tr)
	r.HandleFunc("/api/v1/subscribe", sub.ServeHTTP)

	err := http.ListenAndServe(":7497", r) // TODO: move port to config
	if err != nil {
		log.Error().Msgf("error starting http server: %s", err)
	}
}
