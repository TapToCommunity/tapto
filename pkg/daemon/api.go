package daemon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/wizzomafizzo/mrext/cmd/remote/menu"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/gamesdb"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/tapto/pkg/config"
	"github.com/wizzomafizzo/tapto/pkg/database"
	"github.com/wizzomafizzo/tapto/pkg/platforms/mister"
)

type LaunchRequestMetadata struct {
	ToyModel *string `json:"toyModel"`
}

type LaunchRequest struct {
	UID      string                 `json:"uid"`
	Text     string                 `json:"text"`
	Metadata *LaunchRequestMetadata `json:"metadata"`
}

func handleLaunch(
	cfg *config.UserConfig,
	state *State,
	tq *TokenQueue,
	kbd input.Keyboard,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received launch request")

		var req LaunchRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error().Msgf("error decoding request: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Info().Fields(req).Msgf("launching token")
		// TODO: how do we report back errors?

		t := Token{
			UID:      req.UID,
			Text:     req.Text,
			ScanTime: time.Now(),
		}

		state.SetActiveCard(t)
		tq.Enqueue(t)
	}
}

func handleLaunchBasic(
	cfg *config.UserConfig,
	state *State,
	tq *TokenQueue,
	kbd input.Keyboard,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received basic launch request")

		vars := mux.Vars(r)
		text := vars["rest"]

		log.Info().Msgf("launching basic token: %s", text)

		t := Token{
			UID:      "",
			Text:     text,
			ScanTime: time.Now(),
		}

		state.SetActiveCard(t)
		tq.Enqueue(t)
	}
}

type HistoryReponseEntry struct {
	Time    time.Time `json:"time"`
	UID     string    `json:"uid"`
	Text    string    `json:"text"`
	Success bool      `json:"success"`
}

type HistoryResponse struct {
	Entries []HistoryReponseEntry `json:"entries"`
}

func handleHistory(
	db *database.Database,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received history request")

		entries, err := db.GetHistory()
		if err != nil {
			log.Error().Err(err).Msgf("error getting history")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := HistoryResponse{
			Entries: make([]HistoryReponseEntry, len(entries)),
		}

		for i, e := range entries {
			resp.Entries[i] = HistoryReponseEntry{
				Time:    e.Time,
				UID:     e.UID,
				Text:    e.Text,
				Success: e.Success,
			}
		}

		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Error().Err(err).Msgf("error encoding history response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type TokenResponse struct {
	Type     string    `json:"type"`
	UID      string    `json:"uid"`
	Text     string    `json:"text"`
	ScanTime time.Time `json:"scanTime"`
}

type IndexResponse struct {
	Indexing    bool   `json:"indexing"`
	TotalSteps  int    `json:"totalSteps"`
	CurrentStep int    `json:"currentStep"`
	CurrentDesc string `json:"currentDesc"`
}

type StatusResponse struct {
	ReaderConnected bool          `json:"readerConnected"`
	ReaderType      string        `json:"readerType"`
	ActiveCard      TokenResponse `json:"activeCard"`
	LastScanned     TokenResponse `json:"lastScanned"`
	DisableLauncher bool          `json:"disableLauncher"`
	GamesIndex      IndexResponse `json:"gamesIndex"`
}

func handleStatus(
	state *State,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received status request")

		active := state.GetActiveCard()
		last := state.GetLastScanned()

		resp := StatusResponse{
			ReaderConnected: state.readerConnected,
			ReaderType:      state.readerType,
			ActiveCard: TokenResponse{
				Type:     active.Type,
				UID:      active.UID,
				Text:     active.Text,
				ScanTime: active.ScanTime,
			},
			LastScanned: TokenResponse{
				Type:     last.Type,
				UID:      last.UID,
				Text:     last.Text,
				ScanTime: last.ScanTime,
			},
			DisableLauncher: state.disableLauncher,
			GamesIndex: IndexResponse{
				Indexing:    IndexInstance.Indexing,
				TotalSteps:  IndexInstance.TotalSteps,
				CurrentStep: IndexInstance.CurrentStep,
				CurrentDesc: IndexInstance.CurrentDesc,
			},
		}

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Error().Err(err).Msgf("error encoding status response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

const pageSize = 500

type System struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type SearchResultGame struct {
	System System `json:"system"`
	Name   string `json:"name"`
	Path   string `json:"path"`
}

type SearchResults struct {
	Data     []SearchResultGame `json:"data"`
	Total    int                `json:"total"`
	PageSize int                `json:"pageSize"`
	Page     int                `json:"page"`
}

type Index struct {
	mu          sync.Mutex
	Indexing    bool
	TotalSteps  int
	CurrentStep int
	CurrentDesc string
}

func GetIndexingStatus() string {
	status := "indexStatus:"

	if gamesdb.DbExists() {
		status += "y,"
	} else {
		status += "n,"
	}

	if IndexInstance.Indexing {
		status += "y,"
	} else {
		status += "n,"
	}

	status += fmt.Sprintf(
		"%d,%d,%s",
		IndexInstance.TotalSteps,
		IndexInstance.CurrentStep,
		IndexInstance.CurrentDesc,
	)

	return status
}

func (s *Index) GenerateIndex(cfg *config.UserConfig) {
	if s.Indexing {
		return
	}

	s.mu.Lock()
	s.Indexing = true

	log.Info().Msg("generating games index")
	// websocket.Broadcast(logger, GetIndexingStatus())

	go func() {
		defer s.mu.Unlock()

		_, err := gamesdb.NewNamesIndex(mister.UserConfigToMrext(cfg), games.AllSystems(), func(status gamesdb.IndexStatus) {
			s.TotalSteps = status.Total
			s.CurrentStep = status.Step
			if status.Step == 1 {
				s.CurrentDesc = "Finding games folders..."
			} else if status.Step == status.Total {
				s.CurrentDesc = "Writing database... (" + fmt.Sprint(status.Files) + " games)"
			} else {
				system, err := games.GetSystem(status.SystemId)
				if err != nil {
					s.CurrentDesc = "Indexing " + status.SystemId + "..."
				} else {
					s.CurrentDesc = "Indexing " + system.Name + "..."
				}
			}
			log.Info().Msgf("indexing status: %s", s.CurrentDesc)
			// websocket.Broadcast(logger, GetIndexingStatus())
		})
		if err != nil {
			log.Error().Err(err).Msg("error generating games index")
		}

		s.Indexing = false
		s.TotalSteps = 0
		s.CurrentStep = 0
		s.CurrentDesc = ""

		log.Info().Msg("finished generating games index")
		// websocket.Broadcast(logger, GetIndexingStatus())
	}()
}

func NewIndex() *Index {
	return &Index{}
}

var IndexInstance = NewIndex()

func handleIndexGames(
	cfg *config.UserConfig,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received index games request")

		IndexInstance.GenerateIndex(cfg)
	}
}

type SystemsResponse struct {
	Systems []System `json:"systems"`
}

func handleSystems() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		log.Info().Msg("received systems request")

		resp := SystemsResponse{
			Systems: make([]System, 0),
		}

		indexed, err := gamesdb.IndexedSystems()
		if err != nil {
			log.Error().Err(err).Msgf("error getting indexed systems")
			indexed = []string{}
		}

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

			resp.Systems = append(resp.Systems, System{
				Id:       id,
				Name:     name,
				Category: sysDef.Category,
			})
		}

		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			log.Error().Err(err).Msgf("error encoding systems response")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func handleGames() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("received games request")

		query := r.URL.Query().Get("query")
		system := r.URL.Query().Get("system")

		if query == "" && system == "" {
			http.Error(w, "query or system required", http.StatusBadRequest)
			return
		}

		var results = make([]SearchResultGame, 0)
		var search []gamesdb.SearchResult
		var err error

		if system == "all" || system == "" {
			search, err = gamesdb.SearchNamesWords(games.AllSystems(), query)
		} else {
			system, errSys := games.GetSystem(system)
			if errSys != nil {
				http.Error(w, errSys.Error(), http.StatusBadRequest)
				log.Error().Err(errSys).Msg("error getting system")
				return
			}
			search, err = gamesdb.SearchNamesWords([]games.System{*system}, query)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error().Err(err).Msg("error searching games")
			return
		}

		for _, result := range search {
			system, err := games.GetSystem(result.SystemId)
			if err != nil {
				continue
			}

			results = append(results, SearchResultGame{
				System: System{
					Id:       system.Id,
					Name:     system.Name,
					Category: system.Category,
				},
				Name: result.Name,
				Path: result.Path,
			})
		}

		total := len(results)

		if len(results) > pageSize {
			results = results[:pageSize]
		}

		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(&SearchResults{
			Data:     results,
			Total:    total,
			PageSize: pageSize,
			Page:     1,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error().Err(err).Msg("error encoding games response")
			return
		}
	}
}

func runApiServer(
	cfg *config.UserConfig,
	state *State,
	tq *TokenQueue,
	db *database.Database,
	kbd input.Keyboard,
) {
	r := mux.NewRouter()
	s := r.PathPrefix("/api/v1").Subrouter()

	s.Handle("/launch", handleLaunch(cfg, state, tq, kbd)).Methods(http.MethodPost)
	s.Handle("/launch/{rest:.*}", handleLaunchBasic(cfg, state, tq, kbd)).Methods(http.MethodGet)

	// GET /readers/0/read
	// POST /readers/0/write

	s.Handle("/games", handleGames()).Methods(http.MethodGet)
	s.Handle("/systems", handleSystems()).Methods(http.MethodGet)

	// GET /mappings
	// Return all current mappings, or filter based on query parameters
	// POST /mappings
	// Create a new mapping

	s.Handle("/history", handleHistory(db)).Methods(http.MethodGet)
	s.Handle("/status", handleStatus(state)).Methods(http.MethodGet)

	// GET /settings
	// GET /settings/log

	s.Handle("/settings/index/games", handleIndexGames(cfg)).Methods(http.MethodPost)

	// events SSE

	http.Handle("/", r)

	err := http.ListenAndServe(":7497", nil) // TODO: move port to config
	if err != nil {
		log.Error().Msgf("error starting http server: %s", err)
	}
}
