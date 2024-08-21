package api

import (
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/database"
)

type HistoryReponseEntry struct {
	Time    time.Time `json:"time"`
	Type    string    `json:"type"`
	UID     string    `json:"uid"`
	Text    string    `json:"text"`
	Data    string    `json:"data"`
	Success bool      `json:"success"`
}

type HistoryResponse struct {
	Entries []HistoryReponseEntry `json:"entries"`
}

func (hr *HistoryResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
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
				Type:    e.Type,
				UID:     e.UID,
				Text:    e.Text,
				Data:    e.Data,
				Success: e.Success,
			}
		}

		err = render.Render(w, r, &resp)
		if err != nil {
			log.Error().Err(err).Msgf("error encoding history response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
