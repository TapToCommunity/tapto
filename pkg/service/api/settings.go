package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/service/state"
)

type SettingsResponse struct {
	ConnectionString  string   `json:"connectionString"`
	AllowCommands     bool     `json:"allowCommands"`
	DisableSounds     bool     `json:"disableSounds"`
	ProbeDevice       bool     `json:"probeDevice"`
	ExitGame          bool     `json:"exitGame"`
	ExitGameDelay     int      `json:"exitGameDelay"`
	ExitGameBlocklist []string `json:"exitGameBlocklist"`
	Debug             bool     `json:"debug"`
	Launching         bool     `json:"launching"`
}

func handleSettings(env RequestEnv) error {
	log.Info().Msg("received settings request")

	resp := SettingsResponse{
		ConnectionString:  env.Config.GetConnectionString(),
		AllowCommands:     env.Config.GetAllowCommands(),
		DisableSounds:     env.Config.GetDisableSounds(),
		ProbeDevice:       env.Config.GetProbeDevice(),
		ExitGame:          env.Config.GetExitGame(),
		ExitGameDelay:     env.Config.GetExitGameDelay(),
		ExitGameBlocklist: make([]string, 0),
		Debug:             env.Config.GetDebug(),
		Launching:         !env.State.IsLauncherDisabled(),
	}

	resp.ExitGameBlocklist = append(
		resp.ExitGameBlocklist,
		env.Config.GetExitGameBlocklist()...,
	)

	return env.SendResponse(env.Id, resp)
}

type UpdateSettingsRequest struct {
	ConnectionString  *string   `json:"connectionString"`
	AllowCommands     *bool     `json:"allowCommands"`
	DisableSounds     *bool     `json:"disableSounds"`
	ProbeDevice       *bool     `json:"probeDevice"`
	ExitGame          *bool     `json:"exitGame"`
	ExitGameDelay     *int      `json:"exitGameDelay"`
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
				http.Error(w, "allow_commands can only be changed from localhost", http.StatusForbidden)
				log.Info().Str("remoteAddr", r.RemoteAddr).Bool("allowCommands", *req.AllowCommands).Msg("allow_commands can only be changed from localhost")
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
			log.Info().Int("exitGameDelay", *req.ExitGameDelay).Msg("updating exit game delay")
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
