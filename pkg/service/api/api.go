package api

import (
	"encoding/json"
	"errors"
	"github.com/wizzomafizzo/tapto/pkg/service/api/notifications"
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

var methodMap = map[string]func(RequestEnv) error{
	"launch":          handleLaunch,
	"stop":            handleStopGame,
	"media.index":     handleIndexGames,
	"media.search":    handleGames,
	"settings":        handleSettings,
	"settings.update": handleSettingsUpdate,
	"systems":         handleSystems,
	"history":         handleHistory,
	"mappings":        handleMappings,
	"mappings.new":    handleAddMapping,
	"mappings.delete": handleDeleteMapping,
	"mappings.update": handleUpdateMapping,
	"readers.write":   handleReaderWrite,
	"status":          handleStatus, // TODO: remove, convert to individual methods?
	"version":         handleVersion,
}

type RequestEnv struct {
	Platform     platforms.Platform
	Config       *config.UserConfig
	State        *state.State
	Database     *database.Database
	TokenQueue   *tokens.TokenQueue
	Id           uuid.UUID
	Params       []byte
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
	Result    any          `json:"result,omitempty"`
	Error     *ErrorObject `json:"error,omitempty"`
}

// TODO: request function should return a response and error, not be
// handed the sendResponse and sendError functions
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

	var params []byte
	if req.Params != nil {
		var err error
		// double unmarshal to use json decode on params later
		params, err = json.Marshal(req.Params)
		if err != nil {
			return err
		}
	}

	env.Id = *req.Id
	env.Params = params

	err := fn(env)
	if err != nil {
		err := env.SendError(env.Id, 0, err.Error())
		if err != nil {
			log.Error().Err(err).Msg("problem sending error response")
		}
	}

	return err
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

func Start(
	pl platforms.Platform,
	cfg *config.UserConfig,
	st *state.State,
	tq *tokens.TokenQueue,
	db *database.Database,
	ns <-chan notifications.Notification,
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

	// consume and broadcast notifications
	go func(ns <-chan notifications.Notification) {
		for !st.ShouldStopService() {
			select {
			case n := <-ns:
				ro := RequestObject{
					Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
					Method:    n.Method,
					Params:    &n.Params,
				}

				data, err := json.Marshal(ro)
				if err != nil {
					log.Error().Err(err).Msg("marshalling notification request")
					continue
				}

				err = m.Broadcast(data)
				if err != nil {
					log.Error().Err(err).Msg("broadcasting notification")
				}
			case <-time.After(500 * time.Millisecond):
				continue
			}
		}
	}(ns)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := m.HandleRequest(w, r)
		if err != nil {
			log.Error().Err(err).Msg("handling websocket request")
		}
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		if !json.Valid(msg) {
			log.Error().Msg("data not valid json")
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
				TokenQueue:   tq,
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

		log.Error().Err(err).Msg("message does not match known types")
	})

	r.Get("/version", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte(config.Version))
		if err != nil {
			log.Error().Err(err).Msg("error writing version request")
		}
	})

	r.Get("/launch/*", handleLaunchBasic(st, tq))

	err := http.ListenAndServe(":"+cfg.Api.Port, r)
	if err != nil {
		log.Error().Err(err).Msg("error starting http server")
	}
}
