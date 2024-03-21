package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/database"
)

type HistoryReponseEntry struct {
	Time    time.Time `json:"time"`
	UID     string    `json:"uid"`
	Text    string    `json:"text"`
	Success bool      `json:"success"`
}

type HistoryResponse struct {
	Entries []HistoryReponseEntry `json:"entries"`
}

func handleHistory(
	db *database.Database,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received history request")

		entries, err := db.GetHistory()
		if err != nil {
			log.Error().Err(err).Msgf("error getting history")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := HistoryResponse{
			Entries: make([]HistoryReponseEntry, len(entries)),
		}

		for i, e := range entries {
			resp.Entries[i] = HistoryReponseEntry{
				Time:    e.Time,
				UID:     e.UID,
				Text:    e.Text,
				Success: e.Success,
			}
		}

		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Error().Err(err).Msgf("error encoding history response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
