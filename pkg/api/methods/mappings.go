package methods

import (
	"encoding/json"
	"errors"
	"github.com/wizzomafizzo/tapto/pkg/api/models"
	"github.com/wizzomafizzo/tapto/pkg/api/models/requests"
	"regexp"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/database"
	"github.com/wizzomafizzo/tapto/pkg/utils"
)

type MappingResponse struct {
	database.Mapping
	Added string `json:"added"`
}

type AllMappingsResponse struct {
	Mappings []MappingResponse `json:"mappings"`
}

func HandleMappings(env requests.RequestEnv) error {
	log.Info().Msg("received mappings request")

	resp := AllMappingsResponse{
		Mappings: make([]MappingResponse, 0),
	}

	mappings, err := env.Database.GetAllMappings()
	if err != nil {
		log.Error().Err(err).Msg("error getting mappings")
		return env.SendError(env.Id, 1, "error getting mappings") // TODO: error code
	}

	mrs := make([]MappingResponse, 0)

	for _, m := range mappings {
		t := time.Unix(0, m.Added*int64(time.Millisecond))

		mr := MappingResponse{
			Mapping: m,
			Added:   t.Format(time.RFC3339),
		}

		mrs = append(mrs, mr)
	}

	resp.Mappings = mrs

	return env.SendResponse(env.Id, resp)
}

func validateAddMappingParams(amr *models.AddMappingParams) error {
	if !utils.Contains(database.AllowedMappingTypes, amr.Type) {
		return errors.New("invalid type")
	}

	if !utils.Contains(database.AllowedMatchTypes, amr.Match) {
		return errors.New("invalid match")
	}

	if amr.Pattern == "" {
		return errors.New("missing pattern")
	}

	if amr.Match == database.MatchTypeRegex {
		_, err := regexp.Compile(amr.Pattern)
		if err != nil {
			return err
		}
	}

	return nil
}

func HandleAddMapping(env requests.RequestEnv) error {
	log.Info().Msg("received add mapping request")

	if len(env.Params) == 0 {
		return errors.New("missing params")
	}

	var params models.AddMappingParams
	err := json.Unmarshal(env.Params, &params)
	if err != nil {
		return errors.New("invalid params: " + err.Error())
	}

	err = validateAddMappingParams(&params)
	if err != nil {
		return errors.New("invalid params: " + err.Error())
	}

	m := database.Mapping{
		Label:    params.Label,
		Enabled:  params.Enabled,
		Type:     params.Type,
		Match:    params.Match,
		Pattern:  params.Pattern,
		Override: params.Override,
	}

	err = env.Database.AddMapping(m)
	if err != nil {
		return err
	}

	return nil
}

func HandleDeleteMapping(env requests.RequestEnv) error {
	log.Info().Msg("received delete mapping request")

	if len(env.Params) == 0 {
		return errors.New("missing params")
	}

	var params models.DeleteMappingParams
	err := json.Unmarshal(env.Params, &params)
	if err != nil {
		return errors.New("invalid params: " + err.Error())
	}

	err = env.Database.DeleteMapping(params.Id)
	if err != nil {
		return err
	}

	return nil
}

func validateUpdateMappingParams(umr *models.UpdateMappingParams) error {
	if umr.Label == nil && umr.Enabled == nil && umr.Type == nil && umr.Match == nil && umr.Pattern == nil && umr.Override == nil {
		return errors.New("missing fields")
	}

	if umr.Type != nil && !utils.Contains(database.AllowedMappingTypes, *umr.Type) {
		return errors.New("invalid type")
	}

	if umr.Match != nil && !utils.Contains(database.AllowedMatchTypes, *umr.Match) {
		return errors.New("invalid match")
	}

	if umr.Pattern != nil && *umr.Pattern == "" {
		return errors.New("missing pattern")
	}

	if umr.Match != nil && *umr.Match == database.MatchTypeRegex {
		_, err := regexp.Compile(*umr.Pattern)
		if err != nil {
			return err
		}
	}

	return nil
}

func HandleUpdateMapping(env requests.RequestEnv) error {
	log.Info().Msg("received update mapping request")

	if len(env.Params) == 0 {
		return errors.New("missing params")
	}

	var params models.UpdateMappingParams
	err := json.Unmarshal(env.Params, &params)
	if err != nil {
		return errors.New("invalid params: " + err.Error())
	}

	err = validateUpdateMappingParams(&params)
	if err != nil {
		return errors.New("invalid params: " + err.Error())
	}

	oldMapping, err := env.Database.GetMapping(params.Id)
	if err != nil {
		return err
	}

	newMapping := oldMapping

	if params.Label != nil {
		newMapping.Label = *params.Label
	}

	if params.Enabled != nil {
		newMapping.Enabled = *params.Enabled
	}

	if params.Type != nil {
		newMapping.Type = *params.Type
	}

	if params.Match != nil {
		newMapping.Match = *params.Match
	}

	if params.Pattern != nil {
		newMapping.Pattern = *params.Pattern
	}

	if params.Override != nil {
		newMapping.Override = *params.Override
	}

	err = env.Database.UpdateMapping(params.Id, newMapping)
	if err != nil {
		return err
	}

	return nil
}
