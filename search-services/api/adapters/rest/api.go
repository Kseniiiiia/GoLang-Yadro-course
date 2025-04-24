package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

	"io"

	"yadro.com/course/api/core"
)

type Authenticator interface {
	Login(user, password string) (string, error)
}

func NewLoginHandler(log *slog.Logger, auth Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var credentials struct {
			Name     string `json:"name"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
			log.Debug("failed to decode login request", "error", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		token, err := auth.Login(credentials.Name, credentials.Password)
		if err != nil {
			log.Warn("login failed", "user", credentials.Name, "error", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte(token)); err != nil {
			log.Error("failed to write token")
			return
		}
	}
}

type PingResponse struct {
	Replies map[string]string `json:"replies"`
}

func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reply := PingResponse{
			Replies: make(map[string]string),
		}

		for name, pinger := range pingers {
			if err := pinger.Ping(r.Context()); err != nil {
				reply.Replies[name] = "unavailable"
				log.Error("one of services is not available", "service", name, "error", err.Error())
				continue
			}
			reply.Replies[name] = "ok"
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			log.Error("cannot encode reply", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

type WordsResponse struct {
	Words []string `json:"words"`
	Total int      `json:"total"`
}

func NewWordsHandler(log *slog.Logger, norm core.Normalizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")
		if phrase == "" {
			log.Error("missing phrase")
			http.Error(w, "missing phrase", http.StatusBadRequest)
			return
		}

		words, err := norm.Norm(r.Context(), phrase)
		if err != nil {
			log.Error("bad reply from normalizer", "error", err)
			if errors.Is(err, core.ErrBadArguments) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		reply := WordsResponse{
			Words: words,
			Total: len(words),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			log.Error("cannot encode reply", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

type UpdateStatsResponse struct {
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
	ComicsFetched int `json:"comics_fetched"`
	ComicsTotal   int `json:"comics_total"`
}

func NewUpdateStatsHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := updater.Stats(r.Context())
		if err != nil {
			log.Error("failed to get stats", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		reply := UpdateStatsResponse{
			WordsTotal:    stats.WordsTotal,
			WordsUnique:   stats.WordsUnique,
			ComicsFetched: stats.ComicsFetched,
			ComicsTotal:   stats.ComicsTotal,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			log.Error("cannot encode reply", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

type UpdateStatusResponse struct {
	Status string `json:"status"`
}

var (
	updates atomic.Bool
)

func NewUpdateStatusHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "idle"
		if updates.Load() {
			status = "running"
		}

		reply := UpdateStatusResponse{
			Status: status,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			log.Error("cannot encode reply", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func NewUpdateHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !updates.CompareAndSwap(false, true) {
			log.Info("update already running")
			w.WriteHeader(http.StatusAccepted)
			return
		}

		log.Info("update started")

		err := updater.Update(r.Context())
		if err != nil {
			if errors.Is(err, core.ErrAlreadyExists) {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			log.Error("failed to update", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		updates.Store(false)
		log.Info("update finished")

		w.WriteHeader(http.StatusOK)
	}
}

func NewDropHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Drop(r.Context()); err != nil {
			log.Error("failed to drop database", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

type SearchResponse struct {
	Comics []core.Comics `json:"comics"`
	Total  int32         `json:"total"`
}

func NewSearchHandler(log *slog.Logger, client core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		phrase := r.URL.Query().Get("phrase")
		limitStr := r.URL.Query().Get("limit")

		limit := 10
		if limitStr != "" {
			limitInt, err := strconv.Atoi(limitStr)
			if err != nil || limitInt < 1 {
				log.Warn("invalid limit", "limit", limitStr)
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}
			limit = limitInt
		}

		if phrase == "" {
			log.Warn("phrase is required")
			http.Error(w, "phrase is required", http.StatusBadRequest)
			return
		}

		comics, total, err := client.Search(ctx, phrase, int32(limit))
		if err != nil {
			if errors.Is(err, core.ErrBadArguments) {
				log.Warn("bad request", "error", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			log.Error("search failed", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		response := SearchResponse{
			Comics: comics,
			Total:  total,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("failed to encode response", "error", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}

type IndexSearchResponse struct {
	Comics []core.Comics `json:"comics"`
	Total  int32         `json:"total"`
}

func NewSearchIndexHandler(log *slog.Logger, client core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		phrase := r.URL.Query().Get("phrase")
		limitStr := r.URL.Query().Get("limit")

		limit := 10
		if limitStr != "" {
			limitInt, err := strconv.Atoi(limitStr)
			if err != nil || limitInt < 1 {
				log.Warn("invalid limit", "limit", limitStr)
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}
			limit = limitInt
		}

		if phrase == "" {
			log.Warn("phrase is required")
			http.Error(w, "phrase is required", http.StatusBadRequest)
			return
		}

		comics, total, err := client.IndexSearch(ctx, phrase, int32(limit))
		if err != nil {
			if errors.Is(err, core.ErrBadArguments) {
				log.Warn("bad request", "error", err)
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			log.Error("index search failed", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		response := IndexSearchResponse{
			Comics: comics,
			Total:  total,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("failed to encode response", "error", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}

type DetectHandler struct {
	log          *slog.Logger
	yoloClient   core.YoloDetector
	searchClient core.Searcher
}

func NewDetectHandler(
	log *slog.Logger,
	yoloClient core.YoloDetector,
	searchClient core.Searcher,
) *DetectHandler {
	return &DetectHandler{
		log:          log,
		yoloClient:   yoloClient,
		searchClient: searchClient,
	}
}

func (h *DetectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		h.log.Error("failed to get image", "error", err)
		http.Error(w, "Image required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imgData, err := io.ReadAll(file)
	if err != nil {
		h.log.Error("failed to read image", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	results, err := h.yoloClient.Detect(r.Context(), imgData)
	if err != nil {
		h.log.Error("yolo detection failed", "error", err)
		http.Error(w, "Detection failed", http.StatusInternalServerError)
		return
	}

	var labels []string
	for _, r := range results {
		labels = append(labels, r.Label)
	}
	phrase := strings.Join(labels, " ")

	comics, total, err := h.searchClient.Search(r.Context(), phrase, 10)
	if err != nil {
		h.log.Error("search failed", "error", err)
		if errors.Is(err, core.ErrBadArguments) {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := IndexSearchResponse{
		Comics: comics,
		Total:  total,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("failed to encode response", "error", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
