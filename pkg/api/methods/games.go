package methods

import (
	"encoding/json"
	"errors"
	"github.com/wizzomafizzo/tapto/pkg/api/models"
	"github.com/wizzomafizzo/tapto/pkg/api/models/requests"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/tapto/pkg/assets"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/database/gamesdb"
	"github.com/wizzomafizzo/tapto/pkg/platforms"
)

const defaultMaxResults = 250

type Index struct {
	mu          sync.Mutex
	Indexing    bool
	TotalSteps  int
	CurrentStep int
	CurrentDesc string
	TotalFiles  int
}

func (s *Index) Exists(platform platforms.Platform) bool {
	return gamesdb.DbExists(platform)
}

func (s *Index) GenerateIndex(
	pl platforms.Platform,
	cfg *config.UserConfig,
	ns chan<- models.Notification,
) {
	if s.Indexing {
		return
	}

	s.mu.Lock()
	s.Indexing = true
	s.TotalFiles = 0

	log.Info().Msg("generating media index")
	ns <- models.Notification{
		Method: models.MediaIndexing,
		Params: models.IndexStatusResponse{
			Indexing:    true,
			TotalSteps:  0,
			CurrentStep: 0,
			CurrentDesc: "",
			TotalFiles:  0,
		},
	}

	go func() {
		defer s.mu.Unlock()

		_, err := gamesdb.NewNamesIndex(pl, cfg, gamesdb.AllSystems(), func(status gamesdb.IndexStatus) {
			s.TotalSteps = status.Total
			s.CurrentStep = status.Step
			s.TotalFiles = status.Files
			if status.Step == 1 {
				s.CurrentDesc = "Finding media folders"
			} else if status.Step == status.Total {
				s.CurrentDesc = "Writing database"
			} else {
				system, err := gamesdb.GetSystem(status.SystemId)
				if err != nil {
					s.CurrentDesc = status.SystemId
				} else {
					md, err := assets.GetSystemMetadata(system.Id)
					if err != nil {
						s.CurrentDesc = system.Id
					} else {
						s.CurrentDesc = md.Name
					}
				}
			}
			log.Info().Msgf("indexing status: %s", s.CurrentDesc)
			ns <- models.Notification{
				Method: models.MediaIndexing,
				Params: models.IndexStatusResponse{
					Indexing:    true,
					TotalSteps:  s.TotalSteps,
					CurrentStep: s.CurrentStep,
					CurrentDesc: s.CurrentDesc,
					TotalFiles:  s.TotalFiles,
				},
			}
		})
		if err != nil {
			log.Error().Err(err).Msg("error generating media index")
		}

		s.Indexing = false
		s.TotalSteps = 0
		s.CurrentStep = 0
		s.CurrentDesc = ""

		log.Info().Msg("finished generating media index")
		ns <- models.Notification{
			Method: models.MediaIndexing,
			Params: models.IndexStatusResponse{
				Indexing:    false,
				TotalSteps:  0,
				CurrentStep: 0,
				CurrentDesc: "",
				TotalFiles:  0,
			},
		}
	}()
}

func NewIndex() *Index {
	return &Index{}
}

var IndexInstance = NewIndex()

func HandleIndexMedia(env requests.RequestEnv) (any, error) {
	log.Info().Msg("received index media request")
	IndexInstance.GenerateIndex(env.Platform, env.Config, env.State.Notifications)
	return nil, nil
}

func HandleGames(env requests.RequestEnv) (any, error) {
	log.Info().Msg("received media search request")

	if len(env.Params) == 0 {
		return nil, ErrMissingParams
	}

	var params models.SearchParams
	err := json.Unmarshal(env.Params, &params)
	if err != nil {
		return nil, ErrInvalidParams
	}

	maxResults := defaultMaxResults
	if params.MaxResults != nil && *params.MaxResults > 0 {
		maxResults = *params.MaxResults
	}

	if params.Query == "" && (params.Systems == nil || len(*params.Systems) == 0) {
		return nil, errors.New("query or system is required")
	}

	var results = make([]models.SearchResultMedia, 0)
	var search []gamesdb.SearchResult
	system := params.Systems
	query := params.Query

	if system == nil || len(*system) == 0 {
		search, err = gamesdb.SearchNamesWords(env.Platform, gamesdb.AllSystems(), query)
		if err != nil {
			return nil, errors.New("error searching all media: " + err.Error())
		}
	} else {
		systems := make([]gamesdb.System, 0)
		for _, s := range *system {
			system, err := gamesdb.GetSystem(s)
			if err != nil {
				return nil, errors.New("error getting system: " + err.Error())
			}

			systems = append(systems, *system)
		}

		search, err = gamesdb.SearchNamesWords(env.Platform, systems, query)
		if err != nil {
			return nil, errors.New("error searching media: " + err.Error())
		}
	}

	for _, result := range search {
		system, err := gamesdb.GetSystem(result.SystemId)
		if err != nil {
			continue
		}

		results = append(results, models.SearchResultMedia{
			System: models.System{
				Id:   system.Id,
				Name: system.Id,
			},
			Name: result.Name,
			Path: env.Platform.NormalizePath(env.Config, result.Path),
		})
	}

	total := len(results)

	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return models.SearchResults{
		Results: results,
		Total:   total,
	}, nil
}
