package methods

import (
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/api/models"
	"github.com/wizzomafizzo/tapto/pkg/api/models/requests"
	"github.com/wizzomafizzo/tapto/pkg/assets"
	"github.com/wizzomafizzo/tapto/pkg/database/gamesdb"
)

func HandleSystems(env requests.RequestEnv) (any, error) {
	log.Info().Msg("received systems request")

	indexed, err := gamesdb.IndexedSystems(env.Platform)
	if err != nil {
		log.Error().Err(err).Msgf("error getting indexed systems")
		indexed = []string{}
	}

	if len(indexed) == 0 {
		log.Warn().Msg("no indexed systems found")
	}

	systems := make([]models.System, 0)

	for _, id := range indexed {
		sys, err := gamesdb.GetSystem(id)
		if err != nil {
			log.Error().Err(err).Msgf("error getting system: %s", id)
			continue
		}

		sr := models.System{
			Id: sys.Id,
		}

		sm, err := assets.GetSystemMetadata(id)
		if err != nil {
			log.Error().Err(err).Msgf("error getting system metadata: %s", id)
		}

		sr.Name = sm.Name
		sr.Category = sm.Category

		systems = append(systems, sr)
	}

	return models.SystemsResponse{
		Systems: systems,
	}, nil
}
