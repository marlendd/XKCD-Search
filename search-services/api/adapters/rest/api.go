package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"

	"github.com/VictoriaMetrics/metrics"
	"yadro.com/course/api/core"
)

type pingResponse struct {
	Replies map[string]string `json:"replies"`
}

type statusResponse struct {
	Status core.UpdateStatus `json:"status"`
}

type updateStatsResponse struct {
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
	ComicsFetched int `json:"comics_fetched"`
	ComicsTotal   int `json:"comics_total"`
}

type comicResponse struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type searchResponse struct {
	Comics []comicResponse `json:"comics"`
	Total  int             `json:"total"`
}

type loginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func NewMetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics.WritePrometheus(w, true)
	}
}

//go:generate mockgen -source=api.go -destination=../../mocks/mock_api.go -package=mocks

type Authenticator interface {
	Login(user, password string) (string, error)
}

func NewLoginHandler(log *slog.Logger, auth Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := r.Body.Close(); err != nil {
				log.Error("failed to close request body", "error", err)
			}
		}()

		creds := loginRequest{}
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			log.Error("failed to decode json", "error", err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		token, err := auth.Login(creds.Name, creds.Password)
		if err != nil {
			if errors.Is(err, core.ErrNotAuthorized) {
				log.Info("unauthorized")
				http.Error(w, "user unauthorized", http.StatusUnauthorized)
				return
			}
			log.Error("failed to login", "error", err)
			http.Error(w, "failed to login", http.StatusInternalServerError)
			return
		}

		if _, err := w.Write([]byte(token)); err != nil {
			log.Error("failed to write response", "error", err)
		}
	}
}

func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			repliesMap = make(map[string]string, len(pingers))
			mu         sync.Mutex
			wg         sync.WaitGroup
		)

		for name, pinger := range pingers {
			wg.Go(func() {
				if err := pinger.Ping(r.Context()); err != nil {
					log.Warn("pinger unavailable", "name", name, "error", err)
					mu.Lock()
					repliesMap[name] = "unavailable"
					mu.Unlock()
				} else {
					mu.Lock()
					repliesMap[name] = "ok"
					mu.Unlock()
				}
			})
		}

		wg.Wait()

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(pingResponse{
			Replies: repliesMap,
		})
		if err != nil {
			log.Error("failed to encode json", "error", err)
		}
	}
}

func NewUpdateHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Update(r.Context()); err != nil {
			if errors.Is(err, core.ErrAlreadyExists) {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			log.Error("failed to update", "error", err)
			http.Error(w, "failed to update", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func NewUpdateStatsHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := updater.Stats(r.Context())
		if err != nil {
			log.Error("failed to get stats", "error", err)
			http.Error(w, "failed to get stats", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err = json.NewEncoder(w).Encode(updateStatsResponse{
			WordsTotal:    stats.WordsTotal,
			WordsUnique:   stats.WordsUnique,
			ComicsFetched: stats.ComicsFetched,
			ComicsTotal:   stats.ComicsTotal,
		}); err != nil {
			log.Error("failed to encode json", "error", err)
		}
	}
}

func NewUpdateStatusHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := updater.Status(r.Context())
		if err != nil {
			log.Error("failed to get status", "error", err)
			http.Error(w, "failed to get status", http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(
			statusResponse{Status: status},
		); err != nil {
			log.Error("failed to encode json", "error", err)
		}
	}
}

func NewDropHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := updater.Drop(r.Context())
		if err != nil {
			if errors.Is(err, core.ErrAlreadyExists) {
				log.Error("failed to drop db: update is running", "error", err)
				http.Error(w, "failed to drop db", http.StatusConflict)
				return
			}
			log.Error("failed to drop db", "error", err)
			http.Error(w, "failed to drop db", http.StatusInternalServerError)
		}
	}
}

func NewSearchHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			log.Error("invalid phrase")
			http.Error(w, fmt.Errorf("invalid phrase").Error(), http.StatusBadRequest)
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 10

		if limitStr != "" {
			var err error
			limit, err = strconv.Atoi(limitStr)
			if err != nil || limit <= 0 {
				log.Error("invalid limit value")
				http.Error(w, fmt.Errorf("invalid limit").Error(), http.StatusBadRequest)
				return
			}
		}

		reply, err := searcher.Search(r.Context(), phrase, limit)
		if err != nil {
			log.Error("failed to search", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		comics := make([]comicResponse, len(reply.Comics))
		for i, c := range reply.Comics {
			comics[i] = comicResponse{ID: c.ID, URL: c.URL}
		}
		if err := json.NewEncoder(w).Encode(searchResponse{
			Comics: comics,
			Total:  len(comics),
		}); err != nil {
			log.Error("failed to encode json", "error", err)
		}
	}
}

func NewISearchHandler(log *slog.Logger, isearcher core.ISearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			log.Error("invalid phrase")
			http.Error(w, fmt.Errorf("invalid phrase").Error(), http.StatusBadRequest)
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 10

		if limitStr != "" {
			var err error
			limit, err = strconv.Atoi(limitStr)
			if err != nil || limit <= 0 {
				log.Error("invalid limit value")
				http.Error(w, fmt.Errorf("invalid limit").Error(), http.StatusBadRequest)
				return
			}
		}

		reply, err := isearcher.ISearch(r.Context(), phrase, limit)
		if err != nil {
			log.Error("failed to search", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		comics := make([]comicResponse, len(reply.Comics))
		for i, c := range reply.Comics {
			comics[i] = comicResponse{ID: c.ID, URL: c.URL}
		}
		if err := json.NewEncoder(w).Encode(searchResponse{
			Comics: comics,
			Total:  len(comics),
		}); err != nil {
			log.Error("failed to encode json", "error", err)
		}
	}
}
