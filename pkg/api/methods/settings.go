package methods

import (
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/api"
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

func handleSettings(env api.RequestEnv) error {
	log.Info().Msg("received settings request")

	resp := SettingsResponse{
		// TODO: this is very out of date
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

type UpdateSettingsParams struct {
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

func handleSettingsUpdate(env api.RequestEnv) error {
	log.Info().Msg("received settings update request")

	if len(env.Params) == 0 {
		return errors.New("missing params")
	}

	var params UpdateSettingsParams
	err := json.Unmarshal(env.Params, &params)
	if err != nil {
		return errors.New("invalid params: " + err.Error())
	}

	if params.ConnectionString != nil {
		log.Info().Str("connectionString", *params.ConnectionString).Msg("updating connection string")
		env.Config.SetConnectionString(*params.ConnectionString)
	}

	//if req.AllowCommands != nil {
	//	if !strings.HasPrefix(r.RemoteAddr, "127.0.0.1:") {
	//		http.Error(w, "allow_commands can only be changed from localhost", http.StatusForbidden)
	//		log.Info().Str("remoteAddr", r.RemoteAddr).Bool("allowCommands", *req.AllowCommands).Msg("allow_commands can only be changed from localhost")
	//	} else {
	//		log.Info().Bool("allowCommands", *req.AllowCommands).Msg("updating allow commands")
	//		cfg.SetAllowCommands(*req.AllowCommands)
	//	}
	//}

	if params.DisableSounds != nil {
		log.Info().Bool("disableSounds", *params.DisableSounds).Msg("updating disable sounds")
		env.Config.SetDisableSounds(*params.DisableSounds)
	}

	if params.ProbeDevice != nil {
		log.Info().Bool("probeDevice", *params.ProbeDevice).Msg("updating probe device")
		env.Config.SetProbeDevice(*params.ProbeDevice)
	}

	if params.ExitGameDelay != nil {
		log.Info().Int("exitGameDelay", *params.ExitGameDelay).Msg("updating exit game delay")
		env.Config.SetExitGameDelay(*params.ExitGameDelay)
	}

	if params.ExitGame != nil {
		log.Info().Bool("exitGame", *params.ExitGame).Msg("updating exit game")
		env.Config.SetExitGame(*params.ExitGame)
	}

	if params.ExitGameBlocklist != nil {
		log.Info().Strs("exitGameBlocklist", *params.ExitGameBlocklist).Msg("updating exit game blocklist")
		env.Config.SetExitGameBlocklist(*params.ExitGameBlocklist)
	}

	if params.Debug != nil {
		log.Info().Bool("debug", *params.Debug).Msg("updating debug")
		env.Config.SetDebug(*params.Debug)
	}

	if params.Launching != nil {
		log.Info().Bool("launching", *params.Launching).Msg("updating launching")
		if *params.Launching {
			env.State.EnableLauncher()
		} else {
			env.State.DisableLauncher()
		}
	}

	return env.Config.SaveConfig()
}
