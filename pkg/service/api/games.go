package api

import (
	"github.com/wizzomafizzo/tapto/pkg/service/notifications"
	"net/http"
	"strconv"
	"sync"

	"github.com/go-chi/render"
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

func (sr *SearchResults) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func handleGames(platform platforms.Platform, cfg *config.UserConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received games search request")

		query := r.URL.Query().Get("query")
		system := r.URL.Query().Get("system")
		maxResultsStr := r.URL.Query().Get("maxResults")

		maxResults := defaultMaxResults
		if maxResultsStr != "" {
			parsedMaxResults, err := strconv.Atoi(maxResultsStr)
			if err != nil {
				http.Error(w, "Invalid maxResults value", http.StatusBadRequest)
				return
			}
			maxResults = parsedMaxResults
		}

		if query == "" && system == "" {
			http.Error(w, "query or system required", http.StatusBadRequest)
			return
		}

		var results = make([]SearchResultGame, 0)
		var search []gamesdb.SearchResult
		var err error

		if system == "all" || system == "" {
			search, err = gamesdb.SearchNamesWords(platform, gamesdb.AllSystems(), query)
		} else {
			system, errSys := gamesdb.GetSystem(system)
			if errSys != nil {
				http.Error(w, errSys.Error(), http.StatusBadRequest)
				log.Error().Err(errSys).Msg("error getting system")
				return
			}
			search, err = gamesdb.SearchNamesWords(platform, []gamesdb.System{*system}, query)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error().Err(err).Msg("error searching games")
			return
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
				Path: platform.NormalizePath(cfg, result.Path),
			})
		}

		total := len(results)

		if maxResults == 0 {
		} else if len(results) > maxResults {
			results = results[:maxResults]
		}

		err = render.Render(w, r, &SearchResults{
			Results: results,
			Total:   total,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error().Err(err).Msg("error encoding games response")
			return
		}
	}
}
