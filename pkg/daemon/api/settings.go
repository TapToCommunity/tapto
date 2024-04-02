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
			ConnectionString:  cfg.GetConnectionString(),
			AllowCommands:     cfg.GetAllowCommands(),
			DisableSounds:     cfg.GetDisableSounds(),
			ProbeDevice:       cfg.GetProbeDevice(),
			ExitGame:          cfg.GetExitGame(),
			ExitGameBlocklist: make([]string, 0),
			Debug:             cfg.GetDebug(),
		}

		resp.ExitGameBlocklist = append(resp.ExitGameBlocklist, cfg.GetExitGameBlocklist()...)

		err := render.Render(w, r, &resp)
		if err != nil {
			log.Error().Err(err).Msg("error encoding settings response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type UpdateSettingsRequest struct {
	ConnectionString  *string   `json:"connectionString"`
	DisableSounds     *bool     `json:"disableSounds"`
	ProbeDevice       *bool     `json:"probeDevice"`
	ExitGame          *bool     `json:"exitGame"`
	ExitGameBlocklist *[]string `json:"exitGameBlocklist"`
	Debug             *bool     `json:"debug"`
}

func (usr *UpdateSettingsRequest) Bind(r *http.Request) error {
	return nil
}

func handleSettingsUpdate(cfg *config.UserConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received settings update request")

		var req UpdateSettingsRequest
		err := render.Bind(r, &req)
		if err != nil {
			log.Error().Err(err).Msg("error decoding request")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.ConnectionString != nil {
			log.Info().Str("connectionString", *req.ConnectionString).Msg("updating connection string")
			cfg.SetConnectionString(*req.ConnectionString)
		}

		if req.DisableSounds != nil {
			log.Info().Bool("disableSounds", *req.DisableSounds).Msg("updating disable sounds")
			cfg.SetDisableSounds(*req.DisableSounds)
		}

		if req.ProbeDevice != nil {
			log.Info().Bool("probeDevice", *req.ProbeDevice).Msg("updating probe device")
			cfg.SetProbeDevice(*req.ProbeDevice)
		}

		if req.ExitGame != nil {
			log.Info().Bool("exitGame", *req.ExitGame).Msg("updating exit game")
			cfg.SetExitGame(*req.ExitGame)
		}

		if req.ExitGameBlocklist != nil {
			log.Info().Strs("exitGameBlocklist", *req.ExitGameBlocklist).Msg("updating exit game blocklist")
			cfg.SetExitGameBlocklist(*req.ExitGameBlocklist)
		}

		if req.Debug != nil {
			log.Info().Bool("debug", *req.Debug).Msg("updating debug")
			cfg.SetDebug(*req.Debug)
		}

		err = cfg.SaveConfig()
		if err != nil {
			log.Error().Err(err).Msg("error saving config")
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
