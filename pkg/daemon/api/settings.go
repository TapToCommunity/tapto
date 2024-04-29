package api

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/daemon/state"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
)

type SettingsResponse struct {
	ConnectionString  string   `json:"connectionString"`
	AllowCommands     bool     `json:"allowCommands"`
	DisableSounds     bool     `json:"disableSounds"`
	ProbeDevice       bool     `json:"probeDevice"`
	ExitGame          bool     `json:"exitGame"`
	ExitGameDelay     int8     `json:"exitGameDelay"`
	ExitGameBlocklist []string `json:"exitGameBlocklist"`
	Debug             bool     `json:"debug"`
	Launching         bool     `json:"launching"`
}

func (sr *SettingsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func handleSettings(cfg *config.UserConfig, st *state.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received settings request")

		resp := SettingsResponse{
			ConnectionString:  cfg.GetConnectionString(),
			AllowCommands:     cfg.GetAllowCommands(),
			DisableSounds:     cfg.GetDisableSounds(),
			ProbeDevice:       cfg.GetProbeDevice(),
			ExitGame:          cfg.GetExitGame(),
			ExitGameDelay:     cfg.GetExitGameDelay(),
			ExitGameBlocklist: make([]string, 0),
			Debug:             cfg.GetDebug(),
			Launching:         !st.IsLauncherDisabled(),
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
	AllowCommands     *bool     `json:"allowCommands"`
	DisableSounds     *bool     `json:"disableSounds"`
	ProbeDevice       *bool     `json:"probeDevice"`
	ExitGame          *bool     `json:"exitGame"`
	ExitGameDelay     *int8     `json:"exitGameDelay"`
	ExitGameBlocklist *[]string `json:"exitGameBlocklist"`
	Debug             *bool     `json:"debug"`
	Launching         *bool     `json:"launching"`
}

func (usr *UpdateSettingsRequest) Bind(r *http.Request) error {
	return nil
}

func handleSettingsUpdate(cfg *config.UserConfig, st *state.State) http.HandlerFunc {
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

		if req.AllowCommands != nil {
			if !strings.HasPrefix(r.RemoteAddr, "127.0.0.1:") {
				http.Error(w, "AllowCommands can only be changed from localhost", http.StatusForbidden)
				log.Info().Str("remoteAddr", r.RemoteAddr).Bool("allowCommands", *req.AllowCommands).Msg("AllowCommands can only be changed from localhost")
			} else {
				log.Info().Bool("allowCommands", *req.AllowCommands).Msg("updating allow commands")
				cfg.SetAllowCommands(*req.AllowCommands)
			}
		}

		if req.DisableSounds != nil {
			log.Info().Bool("disableSounds", *req.DisableSounds).Msg("updating disable sounds")
			cfg.SetDisableSounds(*req.DisableSounds)
		}

		if req.ProbeDevice != nil {
			log.Info().Bool("probeDevice", *req.ProbeDevice).Msg("updating probe device")
			cfg.SetProbeDevice(*req.ProbeDevice)
		}

		if req.ExitGameDelay != nil {
			log.Info().Int8("exitGameDelay", *req.ExitGameDelay).Msg("updating exit game delay")
			cfg.SetExitGameDelay(*req.ExitGameDelay)
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

		if req.Launching != nil {
			log.Info().Bool("launching", *req.Launching).Msg("updating launching")
			if *req.Launching {
				st.EnableLauncher()
			} else {
				st.DisableLauncher()
			}
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
