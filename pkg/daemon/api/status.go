package api

import (
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
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

// TODO: legacy, remove in v2
type ReaderStatusResponse struct {
	Connected bool   `json:"connected"`
	Type      string `json:"type"`
}

type ReaderResponse struct {
	// TODO: type
	Connected bool   `json:"connected"`
	Device    string `json:"device"`
	Info      string `json:"info"`
}

type PlayingResponse struct {
	System     string `json:"system"`
	SystemName string `json:"systemName"`
	Game       string `json:"game"`
	GameName   string `json:"gameName"`
	GamePath   string `json:"gamePath"`
}

type StatusResponse struct {
	Reader      ReaderStatusResponse `json:"reader"` // TODO: remove in v2
	Readers     []ReaderResponse     `json:"readers"`
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
	pl platforms.Platform,
	cfg *config.UserConfig,
	st *state.State,
) StatusResponse {
	active := st.GetActiveCard()
	last := st.GetLastScanned()

	readerConnected, readerType := false, ""

	rs := st.ListReaders()
	if len(rs) > 0 {
		// TODO: listing all readers will break API compatibility
		reader, ok := st.GetReader(rs[0])
		if ok && reader != nil {
			readerConnected = reader.Connected()
			readerType = reader.Info()
		}
	}

	readers := make([]ReaderResponse, 0)
	for _, device := range rs {
		reader, ok := st.GetReader(device)
		if ok && reader != nil {
			readers = append(readers, ReaderResponse{
				Connected: reader.Connected(),
				Device:    device,
				Info:      reader.Info(),
			})
		}
	}

	launcherDisabled := st.IsLauncherDisabled()

	return StatusResponse{
		Launching: !launcherDisabled,
		Readers:   readers,
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
			Exists:      IndexInstance.Exists(pl),
			Indexing:    IndexInstance.Indexing,
			TotalSteps:  IndexInstance.TotalSteps,
			CurrentStep: IndexInstance.CurrentStep,
			CurrentDesc: IndexInstance.CurrentDesc,
			TotalFiles:  IndexInstance.TotalFiles,
		},
		Playing: PlayingResponse{
			System:     pl.ActiveSystem(),
			SystemName: pl.ActiveSystem(),
			Game:       pl.ActiveGame(),
			GameName:   pl.ActiveGameName(),
			GamePath:   pl.NormalizePath(cfg, pl.ActiveGamePath()),
		},
	}
}

func handleStatus(
	pl platforms.Platform,
	cfg *config.UserConfig,
	st *state.State,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received status request")

		resp := newStatus(pl, cfg, st)

		err := render.Render(w, r, &resp)
		if err != nil {
			log.Error().Err(err).Msgf("error encoding status response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type VersionResponse struct {
	Version  string `json:"version"`
	Platform string `json:"platform"`
}

func (vr *VersionResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func handleVersion(
	pl platforms.Platform,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received version request")

		resp := VersionResponse{
			Version:  config.Version,
			Platform: pl.Id(),
		}

		err := render.Render(w, r, &resp)
		if err != nil {
			log.Error().Err(err).Msgf("error encoding version response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
