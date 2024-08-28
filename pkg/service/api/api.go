package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/olahol/melody"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/database"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/service/state"
	"github.com/wizzomafizzo/tapto/pkg/tokens"
)

// TODO: should there be a TTL for request timestamps? what about offline misters?
// TODO: can we safely allow launch basic unrestricted for local accounts?
// TODO: should api launches from localhost require allowlist?
// TODO: download log file no longer works, need an alternative

const RequestTimeout = 30 * time.Second

/*
{
	"id": "123e4567-e89b-12d3-a456-426614174000",
	"timestamp": 1612345678901,
	"method": "version"
}
*/

// r.Post("/launch", handleLaunch(st, tq))
// r.Delete("/launch", HandleStopGame(pl))
// r.Post("/readers/0/write", handleReaderWrite(st))
// r.Post("/mappings", handleAddMapping(db))
// r.Delete("/mappings/{id}", handleDeleteMapping(db))
// r.Put("/mappings/{id}", handleUpdateMapping(db))
// r.Put("/settings", handleSettingsUpdate(cfg, st))

var methodMap = map[string]func(RequestEnv) error{
	"games.index":  handleIndexGames,
	"settings.get": handleSettings,
	"systems.get":  handleSystems,
	"history.get":  handleHistory,
	"mappings.get": handleMappings,
	"status":       handleStatus, // TODO: remove, convert to individual methods?
	"version":      handleVersion,
}

type RequestEnv struct {
	Platform     platforms.Platform
	Config       *config.UserConfig
	State        *state.State
	Database     *database.Database
	Id           uuid.UUID
	Params       *any
	SendResponse func(uuid.UUID, any) error
	SendError    func(uuid.UUID, int, string) error
}

type RequestObject struct {
	// no id means the request is a "notification" and requires no response
	Id        *uuid.UUID `json:"id,omitempty"` // UUID v1
	Timestamp int64      `json:"timestamp"`    // unix timestamp (ms)
	Method    string     `json:"method"`
	Params    *any       `json:"params,omitempty"`
}

type ErrorObject struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ResponseObject struct {
	Id        uuid.UUID    `json:"id"`        // UUID v1
	Timestamp int64        `json:"timestamp"` // unix timestamp (ms)
	Result    any          `json:"result"`
	Error     *ErrorObject `json:"error,omitempty"`
}

func handleRequest(env RequestEnv, req RequestObject) error {
	log.Debug().Interface("request", req).Msg("received request")

	fn, ok := methodMap[req.Method]
	if !ok {
		return errors.New("unknown method")
	}

	if req.Id == nil {
		log.Debug().Msg("request is a notification")
		return nil
	}

	env.Id = *req.Id
	env.Params = req.Params

	return fn(env)
}

func sendResponse(s *melody.Session) func(uuid.UUID, any) error {
	return func(id uuid.UUID, result any) error {
		log.Debug().Interface("result", result).Msg("sending response")

		resp := ResponseObject{
			Id:        id,
			Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
			Result:    result,
		}

		data, err := json.Marshal(resp)
		if err != nil {
			return err
		}

		return s.Write(data)
	}
}

func sendError(s *melody.Session) func(uuid.UUID, int, string) error {
	return func(id uuid.UUID, code int, message string) error {
		log.Debug().Int("code", code).Str("message", message).Msg("sending error")

		resp := ResponseObject{
			Id:        id,
			Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
			Error: &ErrorObject{
				Code:    code,
				Message: message,
			},
		}

		data, err := json.Marshal(resp)
		if err != nil {
			return err
		}

		return s.Write(data)
	}
}

func handleResponse(resp ResponseObject) error {
	log.Debug().Interface("response", resp).Msg("received response")
	return nil
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

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		if !json.Valid(msg) {
			log.Error().Msg("invalid json message")
			return
		}

		// try parse a request first, which has a method field
		var req RequestObject
		err := json.Unmarshal(msg, &req)
		if err == nil && req.Method != "" {
			env := RequestEnv{
				Platform:     pl,
				Config:       cfg,
				State:        st,
				Database:     db,
				SendResponse: sendResponse(s),
				SendError:    sendError(s),
			}

			err := handleRequest(env, req)
			if err != nil {
				log.Error().Err(err).Msg("error handling request")
			}
			return
		}

		// otherwise try parse a response, which has an id field
		var resp ResponseObject
		err = json.Unmarshal(msg, &resp)
		if err == nil && resp.Id != uuid.Nil {
			err := handleResponse(resp)
			if err != nil {
				log.Error().Err(err).Msg("error handling response")
			}
			return
		}

		// fall through for no matching type
		log.Error().Msg("invalid json message format")
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
