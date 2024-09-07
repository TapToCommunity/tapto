package methods

import (
	"github.com/wizzomafizzo/tapto/pkg/api/models/requests"
	"time"

	"github.com/rs/zerolog/log"
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

func HandleHistory(env requests.RequestEnv) error {
	log.Info().Msg("received history request")

	entries, err := env.Database.GetHistory()
	if err != nil {
		log.Error().Err(err).Msgf("error getting history")
		return env.SendError(env.Id, 1, "error getting history") // TODO: error code
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

	return env.SendResponse(env.Id, resp)
}
