package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/olahol/melody"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/database"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/service/state"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

const (
	RequestTimeout = 30 * time.Second
)

// r.Get("/status", handleStatus(pl, cfg, st))
// r.Get("/version", handleVersion(pl))
// r.Post("/launch", handleLaunch(st, tq))
// r.Get("/launch/*", handleLaunchBasic(st, tq))
// r.Delete("/launch", HandleStopGame(pl))
// r.Post("/readers/0/write", handleReaderWrite(st))
// r.Get("/games", handleGames(pl, cfg))
// r.Get("/systems", handleSystems(pl))
// r.Get("/mappings", handleMappings(db))
// r.Post("/mappings", handleAddMapping(db))
// r.Delete("/mappings/{id}", handleDeleteMapping(db))
// r.Put("/mappings/{id}", handleUpdateMapping(db))
// r.Get("/history", handleHistory(db))
// r.Get("/settings", handleSettings(cfg, st))
// r.Get("/settings/log/download", handleSettingsDownloadLog(pl))
// r.Put("/settings", handleSettingsUpdate(cfg, st))
// r.Post("/settings/index/games", handleIndexGames(pl, cfg))

type RequestObject struct {
	Id        *string `json:"id"`
	Timestamp *int64  `json:"timestamp"`
	Method    string  `json:"method"`
	Params    *any    `json:"params"`
}

type ErrorObject struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ResponseObject struct {
	Id        string       `json:"id"`
	Timestamp *int64       `json:"timestamp"`
	Result    any          `json:"result"`
	Error     *ErrorObject `json:"error"`
}

func RunApiServer(
	pl platforms.Platform,
	cfg *config.UserConfig,
	st *state.State,
	tq *tokens.TokenQueue,
	db *database.Database,
) {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.NoCache)
	r.Use(middleware.Timeout(RequestTimeout))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*", "capacitor://*"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Accept"},
		ExposedHeaders: []string{},
	}))

	m := melody.New()
	m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		m.HandleRequest(w, r)
	})

	r.Get("/version", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(config.Version))
	})

	r.Get("/launch/*", handleLaunchBasic(st, tq))

	err := http.ListenAndServe(":"+cfg.Api.Port, r)
	if err != nil {
		log.Error().Err(err).Msg("error starting http server")
	}
}
