package api

import (
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
)

type TokenResponse struct {
	Type     string    `json:"type"`
	UID      string    `json:"uid"`
	Text     string    `json:"text"`
	Data     string    `json:"data"`
	ScanTime time.Time `json:"scanTime"`
}

type IndexResponse struct {
	Exists      bool   `json:"exists"`
	Indexing    bool   `json:"indexing"`
	TotalSteps  int    `json:"totalSteps"`
	CurrentStep int    `json:"currentStep"`
	CurrentDesc string `json:"currentDesc"`
	TotalFiles  int    `json:"totalFiles"`
}

type ReaderStatusResponse struct {
	Connected bool   `json:"connected"`
	Type      string `json:"type"`
}

type PlayingResponse struct {
	System     string `json:"system"`
	SystemName string `json:"systemName"`
	Game       string `json:"game"`
	GameName   string `json:"gameName"`
	GamePath   string `json:"gamePath"`
}

type StatusResponse struct {
	Reader      ReaderStatusResponse `json:"reader"`
	ActiveToken TokenResponse        `json:"activeToken"`
	LastToken   TokenResponse        `json:"lastToken"`
	Launching   bool                 `json:"launching"`
	GamesIndex  IndexResponse        `json:"gamesIndex"`
	Playing     PlayingResponse      `json:"playing"`
}

func (sr *StatusResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func newStatus(
	platform platforms.Platform,
	cfg *config.UserConfig,
	st *state.State,
	tr *mister.Tracker,
) StatusResponse {
	active := st.GetActiveCard()
	last := st.GetLastScanned()
	readerConnected, readerType := st.GetReaderStatus()
	launcherDisabled := st.IsLauncherDisabled()

	return StatusResponse{
		Launching: !launcherDisabled,
		Reader: ReaderStatusResponse{
			Connected: readerConnected,
			Type:      readerType,
		},
		ActiveToken: TokenResponse{
			Type:     active.Type,
			UID:      active.UID,
			Text:     active.Text,
			Data:     active.Data,
			ScanTime: active.ScanTime,
		},
		LastToken: TokenResponse{
			Type:     last.Type,
			UID:      last.UID,
			Text:     last.Text,
			Data:     last.Data,
			ScanTime: last.ScanTime,
		},
		GamesIndex: IndexResponse{
			Exists:      IndexInstance.Exists(platform),
			Indexing:    IndexInstance.Indexing,
			TotalSteps:  IndexInstance.TotalSteps,
			CurrentStep: IndexInstance.CurrentStep,
			CurrentDesc: IndexInstance.CurrentDesc,
			TotalFiles:  IndexInstance.TotalFiles,
		},
		Playing: PlayingResponse{
			System:     tr.ActiveSystem,
			SystemName: tr.ActiveSystemName,
			Game:       tr.ActiveGameId,
			GameName:   tr.ActiveGameName,
			GamePath:   platform.NormalizePath(cfg, tr.ActiveGamePath),
		},
	}
}

func handleStatus(
	platform platforms.Platform,
	cfg *config.UserConfig,
	st *state.State,
	tr *mister.Tracker,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received status request")

		resp := newStatus(platform, cfg, st, tr)

		err := render.Render(w, r, &resp)
		if err != nil {
			log.Error().Err(err).Msgf("error encoding status response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
