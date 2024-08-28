package api

import (
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/database"
	"github.com/wizzomafizzo/tapto/pkg/utils"
)

type MappingResponse struct {
	database.Mapping
	Added string `json:"added"`
}

type AllMappingsResponse struct {
	Mappings []MappingResponse `json:"mappings"`
}

func handleMappings(env RequestEnv) error {
	log.Info().Msg("received mappings request")

	resp := AllMappingsResponse{
		Mappings: make([]MappingResponse, 0),
	}

	mappings, err := env.Database.GetAllMappings()
	if err != nil {
		log.Error().Err(err).Msg("error getting mappings")
		return env.SendError(env.Id, 1, "error getting mappings") // TODO: error code
	}

	mrs := make([]MappingResponse, 0)

	for _, m := range mappings {
		t := time.Unix(0, m.Added*int64(time.Millisecond))

		mr := MappingResponse{
			Mapping: m,
			Added:   t.Format(time.RFC3339),
		}

		mrs = append(mrs, mr)
	}

	resp.Mappings = mrs

	return env.SendResponse(env.Id, resp)
}

type AddMappingRequest struct {
	Label    string `json:"label"`
	Enabled  bool   `json:"enabled"`
	Type     string `json:"type"`
	Match    string `json:"match"`
	Pattern  string `json:"pattern"`
	Override string `json:"override"`
}

func (amr *AddMappingRequest) Bind(r *http.Request) error {
	if !utils.Contains(database.AllowedMappingTypes, amr.Type) {
		return errors.New("invalid type")
	}

	if !utils.Contains(database.AllowedMatchTypes, amr.Match) {
		return errors.New("invalid match")
	}

	if amr.Pattern == "" {
		return errors.New("missing pattern")
	}

	if amr.Match == database.MatchTypeRegex {
		_, err := regexp.Compile(amr.Pattern)
		if err != nil {
			return err
		}
	}

	return nil
}

func handleAddMapping(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received add mapping request")

		var req AddMappingRequest
		err := render.Bind(r, &req)
		if err != nil {
			log.Error().Err(err).Msg("error decoding request")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		m := database.Mapping{
			Label:    req.Label,
			Enabled:  req.Enabled,
			Type:     req.Type,
			Match:    req.Match,
			Pattern:  req.Pattern,
			Override: req.Override,
		}

		err = db.AddMapping(m)
		if err != nil {
			log.Error().Err(err).Msg("error adding mapping")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func handleDeleteMapping(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received delete mapping request")

		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		err := db.DeleteMapping(id)
		if err != nil {
			log.Error().Err(err).Msg("error deleting mapping")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type UpdateMappingRequest struct {
	Label    *string `json:"label"`
	Enabled  *bool   `json:"enabled"`
	Type     *string `json:"type"`
	Match    *string `json:"match"`
	Pattern  *string `json:"pattern"`
	Override *string `json:"override"`
}

func (umr *UpdateMappingRequest) Bind(r *http.Request) error {
	if umr.Label == nil && umr.Enabled == nil && umr.Type == nil && umr.Match == nil && umr.Pattern == nil && umr.Override == nil {
		return errors.New("missing fields")
	}

	if umr.Type != nil && !utils.Contains(database.AllowedMappingTypes, *umr.Type) {
		return errors.New("invalid type")
	}

	if umr.Match != nil && !utils.Contains(database.AllowedMatchTypes, *umr.Match) {
		return errors.New("invalid match")
	}

	if umr.Pattern != nil && *umr.Pattern == "" {
		return errors.New("missing pattern")
	}

	if umr.Match != nil && *umr.Match == database.MatchTypeRegex {
		_, err := regexp.Compile(*umr.Pattern)
		if err != nil {
			return err
		}
	}

	return nil
}

func handleUpdateMapping(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received update mapping request")

		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		var req UpdateMappingRequest
		err := render.Bind(r, &req)
		if err != nil {
			log.Error().Err(err).Msg("error decoding request")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		oldMapping, err := db.GetMapping(id)
		if err != nil {
			log.Error().Err(err).Msg("error getting mapping")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newMapping := oldMapping

		if req.Label != nil {
			newMapping.Label = *req.Label
		}

		if req.Enabled != nil {
			newMapping.Enabled = *req.Enabled
		}

		if req.Type != nil {
			newMapping.Type = *req.Type
		}

		if req.Match != nil {
			newMapping.Match = *req.Match
		}

		if req.Pattern != nil {
			newMapping.Pattern = *req.Pattern
		}

		if req.Override != nil {
			newMapping.Override = *req.Override
		}

		if oldMapping == newMapping {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		err = db.UpdateMapping(id, newMapping)
		if err != nil {
			log.Error().Err(err).Msg("error updating mapping")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
