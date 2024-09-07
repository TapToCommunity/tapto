package api

import (
	"encoding/json"
	"errors"
	"github.com/wizzomafizzo/tapto/pkg/service/api/notifications"
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
	Indexing    bool   `json:"indexing"`
	TotalSteps  int    `json:"totalSteps"`
	CurrentStep int    `json:"currentStep"`
	CurrentDesc string `json:"currentDesc"`
	TotalFiles  int    `json:"totalFiles"`
}

func (s *Index) Exists(platform platforms.Platform) bool {
	return gamesdb.DbExists(platform)
}

func (s *Index) GenerateIndex(
	pl platforms.Platform,
	cfg *config.UserConfig,
	ns chan<- notifications.Notification,
) {
	if s.Indexing {
		return
	}

	s.mu.Lock()
	s.Indexing = true
	s.TotalFiles = 0

	log.Info().Msg("generating games index")
	ns <- notifications.Notification{
		Method: notifications.MediaIndexing,
		Params: s,
	}

	go func() {
		defer s.mu.Unlock()

		_, err := gamesdb.NewNamesIndex(pl, cfg, gamesdb.AllSystems(), func(status gamesdb.IndexStatus) {
			s.TotalSteps = status.Total
			s.CurrentStep = status.Step
			s.TotalFiles = status.Files
			if status.Step == 1 {
				s.CurrentDesc = "Finding games folders"
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
			ns <- notifications.Notification{
				Method: notifications.MediaIndexing,
				Params: s,
			}
		})
		if err != nil {
			log.Error().Err(err).Msg("error generating games index")
		}

		s.Indexing = false
		s.TotalSteps = 0
		s.CurrentStep = 0
		s.CurrentDesc = ""

		log.Info().Msg("finished generating games index")
		ns <- notifications.Notification{
			Method: notifications.MediaIndexing,
			Params: s,
		}
	}()
}

func NewIndex() *Index {
	return &Index{}
}

var IndexInstance = NewIndex()

func handleIndexGames(env RequestEnv) error {
	log.Info().Msg("received index games request")
	IndexInstance.GenerateIndex(env.Platform, env.Config, env.State.Notifications)
	return nil
}

type SearchResultGame struct {
	System System `json:"system"`
	Name   string `json:"name"`
	Path   string `json:"path"`
}

type SearchResults struct {
	Results []SearchResultGame `json:"results"`
	Total   int                `json:"total"`
}

type SearchParams struct {
	Query      string `json:"query"`
	System     string `json:"system"`
	MaxResults *int   `json:"maxResults"`
}

func handleGames(env RequestEnv) error {
	log.Info().Msg("received games search request")

	if len(env.Params) == 0 {
		return errors.New("missing params")
	}

	var params SearchParams
	err := json.Unmarshal(env.Params, &params)
	if err != nil {
		return errors.New("invalid params: " + err.Error())
	}

	maxResults := defaultMaxResults
	if params.MaxResults != nil && *params.MaxResults > 0 {
		maxResults = *params.MaxResults
	}

	if params.Query == "" && params.System == "" {
		return errors.New("query or system is required")
	}

	var results = make([]SearchResultGame, 0)
	var search []gamesdb.SearchResult
	system := params.System
	query := params.Query

	if system == "all" || system == "" {
		search, err = gamesdb.SearchNamesWords(env.Platform, gamesdb.AllSystems(), query)
		if err != nil {
			return errors.New("error searching all media: " + err.Error())
		}
	} else {
		system, err := gamesdb.GetSystem(system)
		if err != nil {
			return errors.New("error getting system: " + err.Error())
		}

		search, err = gamesdb.SearchNamesWords(env.Platform, []gamesdb.System{*system}, query)
		if err != nil {
			return errors.New("error searching " + system.Id + " media: " + err.Error())
		}
	}

	for _, result := range search {
		system, err := gamesdb.GetSystem(result.SystemId)
		if err != nil {
			continue
		}

		results = append(results, SearchResultGame{
			System: System{
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

	return env.SendResponse(env.Id, &SearchResults{
		Results: results,
		Total:   total,
	})
}
