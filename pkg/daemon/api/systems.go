package api

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/mrext/cmd/remote/menu"
	"github.com/wizzomafizzo/mrext/pkg/games"
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

func (sr *SystemsResponse) Render(w http.ResponseWriter, _ *http.Request) error {
	return nil
}

func handleSystems() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received systems request")

		indexed, err := gamesdb.IndexedSystems()
		if err != nil {
			log.Error().Err(err).Msgf("error getting indexed systems")
			indexed = []string{}
		}

		if len(indexed) == 0 {
			log.Warn().Msg("no indexed systems found")
		}

		systems := make([]System, 0)

		for _, system := range indexed {
			id := system
			sysDef, ok := games.Systems[id]
			if !ok {
				continue
			}

			name, _ := menu.GetNamesTxt(sysDef.Name, "")
			if name == "" {
				name = sysDef.Name
			}

			systems = append(systems, System{
				Id:       id,
				Name:     name,
				Category: sysDef.Category,
			})
		}

		err = render.Render(w, r, &SystemsResponse{
			Systems: systems,
		})
		if err != nil {
			log.Error().Err(err).Msgf("error encoding systems response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
