package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
)

type TokenResponse struct {
	Type     string    `json:"type"`
	UID      string    `json:"uid"`
	Text     string    `json:"text"`
	ScanTime time.Time `json:"scanTime"`
}

type IndexResponse struct {
	Indexing    bool   `json:"indexing"`
	TotalSteps  int    `json:"totalSteps"`
	CurrentStep int    `json:"currentStep"`
	CurrentDesc string `json:"currentDesc"`
}

type ReaderStatusResponse struct {
	ReaderConnected bool   `json:"readerConnected"`
	ReaderType      string `json:"readerType"`
}

type PlayingPayload struct {
	System     string `json:"system"`
	SystemName string `json:"systemName"`
	Game       string `json:"game"`
	GameName   string `json:"gameName"`
}

type StatusResponse struct {
	Reader      ReaderStatusResponse `json:"reader"`
	ActiveToken TokenResponse        `json:"activeToken"`
	LastToken   TokenResponse        `json:"lastToken"`
	Launching   bool                 `json:"launching"`
	GamesIndex  IndexResponse        `json:"gamesIndex"`
	Playing     PlayingPayload       `json:"playing"`
}

func handleStatus(
	st *state.State,
	tr *mister.Tracker,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received status request")

		active := st.GetActiveCard()
		last := st.GetLastScanned()
		readerConnected, readerType := st.GetReaderStatus()
		launcherDisabled := st.IsLauncherDisabled()

		resp := StatusResponse{
			Launching: !launcherDisabled,
			Reader: ReaderStatusResponse{
				ReaderConnected: readerConnected,
				ReaderType:      readerType,
			},
			ActiveToken: TokenResponse{
				Type:     active.Type,
				UID:      active.UID,
				Text:     active.Text,
				ScanTime: active.ScanTime,
			},
			LastToken: TokenResponse{
				Type:     last.Type,
				UID:      last.UID,
				Text:     last.Text,
				ScanTime: last.ScanTime,
			},
			GamesIndex: IndexResponse{
				Indexing:    IndexInstance.Indexing,
				TotalSteps:  IndexInstance.TotalSteps,
				CurrentStep: IndexInstance.CurrentStep,
				CurrentDesc: IndexInstance.CurrentDesc,
			},
			Playing: PlayingPayload{
				System:     tr.ActiveSystem,
				SystemName: tr.ActiveSystemName,
				Game:       tr.ActiveGame,
				GameName:   tr.ActiveGameName,
			},
		}

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Error().Err(err).Msgf("error encoding status response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
