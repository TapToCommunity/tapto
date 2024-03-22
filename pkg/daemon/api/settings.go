package api

import (
	"io"
	"net/http"
	"os"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
)

type SettingsResponse struct {
	ConnectionString  string   `json:"connectionString"`
	AllowCommands     bool     `json:"allowCommands"`
	DisableSounds     bool     `json:"disableSounds"`
	ProbeDevice       bool     `json:"probeDevice"`
	ExitGame          bool     `json:"exitGame"`
	ExitGameBlocklist []string `json:"exitGameBlocklist"`
	Debug             bool     `json:"debug"`
}

func (sr *SettingsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func handleSettings(cfg *config.UserConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received settings request")

		resp := SettingsResponse{
			ConnectionString:  cfg.TapTo.ConnectionString,
			AllowCommands:     cfg.TapTo.AllowCommands,
			DisableSounds:     cfg.TapTo.DisableSounds,
			ProbeDevice:       cfg.TapTo.ProbeDevice,
			ExitGame:          cfg.TapTo.ExitGame,
			ExitGameBlocklist: make([]string, 0),
			Debug:             cfg.TapTo.Debug,
		}

		resp.ExitGameBlocklist = append(resp.ExitGameBlocklist, cfg.TapTo.ExitGameBlocklist...)

		err := render.Render(w, r, &resp)
		if err != nil {
			log.Error().Err(err).Msg("error encoding settings response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func handleSettingsDownloadLog() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received settings log request")

		file, err := os.Open(mister.LogFile)
		if err != nil {
			log.Error().Err(err).Msg("error opening log file")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		w.Header().Set("Content-Disposition", "attachment; filename=tapto.log")
		w.Header().Set("Content-Type", "text/plain")

		_, err = io.Copy(w, file)
		if err != nil {
			log.Error().Err(err).Msg("error copying log file")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
