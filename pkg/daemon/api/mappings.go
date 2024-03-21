package api

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/database"
)

type MappingsResponse struct {
	Uids  map[string]string `json:"uids"`
	Texts map[string]string `json:"texts"`
}

func handleMappings(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received mappings request")

		resp := MappingsResponse{
			Uids:  make(map[string]string),
			Texts: make(map[string]string),
		}

		mappings, err := db.GetMappings()
		if err != nil {
			log.Error().Err(err).Msg("error getting mappings")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp.Uids = mappings.Uids
		resp.Texts = mappings.Texts

		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Error().Err(err).Msg("error encoding mappings response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type AddMappingRequest struct {
	MappingType string `json:"type"`
	Original    string `json:"original"`
	Text        string `json:"text"`
}

func handleAddMapping(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received add mapping request")

		var req AddMappingRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error().Err(err).Msg("error decoding request")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.MappingType == database.MappingTypeUID {
			err = db.AddUidMapping(req.Original, req.Text)
		} else if req.MappingType == database.MappingTypeText {
			err = db.AddTextMapping(req.Original, req.Text)
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
