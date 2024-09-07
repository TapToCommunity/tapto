package methods

import (
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/api"
	"github.com/wizzomafizzo/tapto/pkg/assets"
	"github.com/wizzomafizzo/tapto/pkg/database/gamesdb"
)

type System struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type SystemsResponse struct {
	Systems []System `json:"systems"`
}

func handleSystems(env api.RequestEnv) error {
	log.Info().Msg("received systems request")

	indexed, err := gamesdb.IndexedSystems(env.Platform)
	if err != nil {
		log.Error().Err(err).Msgf("error getting indexed systems")
		indexed = []string{}
	}

	if len(indexed) == 0 {
		log.Warn().Msg("no indexed systems found")
	}

	systems := make([]System, 0)

	for _, id := range indexed {
		sys, err := gamesdb.GetSystem(id)
		if err != nil {
			log.Error().Err(err).Msgf("error getting system: %s", id)
			continue
		}

		sr := System{
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

	return env.SendResponse(env.Id, SystemsResponse{
		Systems: systems,
	})
}
