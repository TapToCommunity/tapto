package api

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/daemon/api/websocket"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/database"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

const (
	SubStreamStatus = "status"
	SubStreamEvents = "events"
)

func setupWs(
	pl platforms.Platform,
	cfg *config.UserConfig,
	st *state.State,
) {
	send := func() {
		status, err := json.Marshal(newStatus(pl, cfg, st))
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

	trHook := func() {
		log.Debug().Msg("tracker update hook")
		if pl.GetActiveLauncher() == "" {
			st.SetSoftwareToken(nil)
		}
		send()
	}
	pl.SetEventHook(&trHook)

	idxHook := func(_ *Index) {
		log.Debug().Msg("index update hook")
		send()
	}
	IndexInstance.SetEventHook(&idxHook)
}

// https://github.com/ironstar-io/chizerolog/blob/master/main.go
func LoggerMiddleware(logger *zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			log := logger.With().Logger()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				t2 := time.Now()

				// Recover and record stack traces in case of a panic
				if rec := recover(); rec != nil {
					log.Error().
						Str("type", "error").
						Timestamp().
						Interface("recover_info", rec).
						Bytes("debug_stack", debug.Stack()).
						Msg("log system error")
					http.Error(ww, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}

				// log end request
				log.Debug().
					Str("type", "access").
					Timestamp().
					Fields(map[string]interface{}{
						"remote_ip":  r.RemoteAddr,
						"url":        r.URL.Path,
						"proto":      r.Proto,
						"method":     r.Method,
						"user_agent": r.Header.Get("User-Agent"),
						"status":     ww.Status(),
						"latency_ms": float64(t2.Sub(t1).Nanoseconds()) / 1000000.0,
						"bytes_in":   r.Header.Get("Content-Length"),
						"bytes_out":  ww.BytesWritten(),
					}).
					Msg("incoming_request")
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

func RunApiServer(
	pl platforms.Platform,
	cfg *config.UserConfig,
	st *state.State,
	tq *tokens.TokenQueue,
	db *database.Database,
) {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(LoggerMiddleware(&log.Logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*", "capacitor://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))

		r.Get("/status", handleStatus(pl, cfg, st))
		r.Get("/version", handleVersion(pl))

		r.Post("/launch", handleLaunch(st, tq))
		r.Get("/launch/*", handleLaunchBasic(st, tq))
		r.Delete("/launch", HandleStopGame(pl))

		// GET /readers/0/read
		r.Post("/readers/0/write", handleReaderWrite(st))

		r.Get("/games", handleGames(pl, cfg))
		r.Get("/systems", handleSystems(pl))

		r.Get("/mappings", handleMappings(db))
		r.Post("/mappings", handleAddMapping(db))
		r.Delete("/mappings/{id}", handleDeleteMapping(db))
		r.Put("/mappings/{id}", handleUpdateMapping(db))

		r.Get("/history", handleHistory(db))

		r.Get("/settings", handleSettings(cfg, st))
		r.Get("/settings/log/download", handleSettingsDownloadLog(pl))
		r.Put("/settings", handleSettingsUpdate(cfg, st))
		r.Post("/settings/index/games", handleIndexGames(pl, cfg))
	})

	setupWs(pl, cfg, st)
	r.HandleFunc("/api/v1/ws", websocket.Handle(
		func() []string {
			status, err := json.Marshal(newStatus(pl, cfg, st))
			if err != nil {
				log.Error().Err(err).Msg("error encoding status")
			}
			return []string{"STATUS " + string(status)}
		},
		func(msg string) string {
			return ""
		},
	))

	err := http.ListenAndServe(":"+string(cfg.Api.Port), r)
	if err != nil {
		log.Error().Msgf("error starting http server: %s", err)
	}
}
