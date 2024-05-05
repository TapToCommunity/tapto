package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/database"
)

type MappingsResponse struct {
	Uids  map[string]string `json:"uids"`
	Texts map[string]string `json:"texts"`
	Data map[string]string `json:"data"`
}

func (mr *MappingsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func handleMappings(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received mappings request")

		resp := MappingsResponse{
			Uids:  make(map[string]string),
			Texts: make(map[string]string),
			Data:  make(map[string]string),
		}

		mappings, err := db.GetMappings()
		if err != nil {
			log.Error().Err(err).Msg("error getting mappings")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp.Uids = mappings.Uids
		resp.Texts = mappings.Texts
		resp.Data = mappings.Data

		err = render.Render(w, r, &resp)
		if err != nil {
			log.Error().Err(err).Msg("error encoding mappings response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type AddMappingRequest struct {
	Type  string `json:"type"`
	Match string `json:"match"`
	Text  string `json:"text"`
	Data  string `json:"data"`
}

func (amr *AddMappingRequest) Bind(r *http.Request) error {
	if amr.Type == "" {
		return errors.New("missing type")
	}

	if amr.Type != database.MappingTypeUID && amr.Type != database.MappingTypeText && amr.Type != database.MappingTypeData {
		return errors.New("invalid type: " + amr.Type)
	}

	if amr.Match == "" {
		return errors.New("missing match")
	}

	if amr.Text == "" {
		return errors.New("missing text")
	}

	if amr.Data == "" {
		return errors.New("missing data")
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

		if req.Type == database.MappingTypeUID {
			err = db.AddUidMapping(req.Match, req.Text)
		} else if req.Type == database.MappingTypeText {
			err = db.AddTextMapping(req.Match, req.Text)
		} else if req.Type == database.MappingTypeData {
			err = db.AddDataMapping(req.Match, req.Data)
		} else {
			http.Error(w, "invalid mapping type", http.StatusBadRequest)
			return
		}

		if err != nil {
			log.Error().Err(err).Msg("error adding mapping")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
