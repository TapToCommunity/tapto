package methods

import (
	"encoding/json"
	"errors"
	"github.com/ZaparooProject/zaparoo-core/pkg/api/models"
	"github.com/ZaparooProject/zaparoo-core/pkg/api/models/requests"
	"regexp"
	"strconv"
	"time"

	"github.com/ZaparooProject/zaparoo-core/pkg/database"
	"github.com/ZaparooProject/zaparoo-core/pkg/utils"
	"github.com/rs/zerolog/log"
)

func HandleMappings(env requests.RequestEnv) (any, error) {
	log.Info().Msg("received mappings request")

	resp := models.AllMappingsResponse{
		Mappings: make([]models.MappingResponse, 0),
	}

	mappings, err := env.Database.GetAllMappings()
	if err != nil {
		log.Error().Err(err).Msg("error getting mappings")
		return nil, errors.New("error getting mappings")
	}

	mrs := make([]models.MappingResponse, 0)

	for _, m := range mappings {
		t := time.Unix(0, m.Added*int64(time.Millisecond))

		mr := models.MappingResponse{
			Id:       m.Id,
			Added:    t.Format(time.RFC3339),
			Label:    m.Label,
			Enabled:  m.Enabled,
			Type:     m.Type,
			Match:    m.Match,
			Pattern:  m.Pattern,
			Override: m.Override,
		}

		mrs = append(mrs, mr)
	}

	resp.Mappings = mrs

	return resp, nil
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

func HandleAddMapping(env requests.RequestEnv) (any, error) {
	log.Info().Msg("received add mapping request")

	if len(env.Params) == 0 {
		return nil, ErrMissingParams
	}

	var params models.AddMappingParams
	err := json.Unmarshal(env.Params, &params)
	if err != nil {
		return nil, ErrInvalidParams
	}

	err = validateAddMappingParams(&params)
	if err != nil {
		log.Error().Err(err).Msg("invalid params")
		return nil, ErrInvalidParams
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
		return nil, err
	}

	return nil, nil
}

func HandleDeleteMapping(env requests.RequestEnv) (any, error) {
	log.Info().Msg("received delete mapping request")

	if len(env.Params) == 0 {
		return nil, ErrMissingParams
	}

	var params models.DeleteMappingParams
	err := json.Unmarshal(env.Params, &params)
	if err != nil {
		return nil, ErrInvalidParams
	}

	err = env.Database.DeleteMapping(strconv.Itoa(params.Id))
	if err != nil {
		return nil, err
	}

	return nil, nil
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

func HandleUpdateMapping(env requests.RequestEnv) (any, error) {
	log.Info().Msg("received update mapping request")

	if len(env.Params) == 0 {
		return nil, ErrMissingParams
	}

	var params models.UpdateMappingParams
	err := json.Unmarshal(env.Params, &params)
	if err != nil {
		return nil, ErrInvalidParams
	}

	err = validateUpdateMappingParams(&params)
	if err != nil {
		log.Error().Err(err).Msg("invalid params")
		return nil, ErrInvalidParams
	}

	oldMapping, err := env.Database.GetMapping(strconv.Itoa(params.Id))
	if err != nil {
		return nil, err
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

	err = env.Database.UpdateMapping(strconv.Itoa(params.Id), newMapping)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
