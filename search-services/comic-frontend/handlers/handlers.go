package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

type Comic struct {
	ID    int    `json:"id"`
	URL   string `json:"url"`
	Score int    `json:"score"`
}

type UpdateStats struct {
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
	ComicsFetched int `json:"comics_fetched"`
	ComicsTotal   int `json:"comics_total"`
}

type UpdateStatus struct {
	Status string `json:"status"`
}

type Handler struct {
	log       *slog.Logger
	client    *http.Client
	apiURL    string
	templates *template.Template
}

func NewHandler(log *slog.Logger, client *http.Client, apiURL string) *Handler {
	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	return &Handler{
		log:       log,
		client:    client,
		apiURL:    apiURL,
		templates: tmpl,
	}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	data := struct {
		SearchQuery string
		FastSearch  bool
	}{
		SearchQuery: r.URL.Query().Get("q"),
		FastSearch:  r.URL.Query().Get("fast") == "on",
	}

	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		h.log.Error("failed to render index", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) ImageSearch(w http.ResponseWriter, r *http.Request) {
	data := struct {
		PageTitle string
	}{
		PageTitle: "Search by Image",
	}

	if err := h.templates.ExecuteTemplate(w, "image_search.html", data); err != nil {
		h.log.Error("failed to render image search page", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) Detect(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.log.Error("failed to parse multipart form", "error", err)
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		h.log.Error("failed to get image from request", "error", err)
		http.Error(w, "Bad request: no image provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", header.Filename)
	if err != nil {
		h.log.Error("failed to create form file", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(part, file); err != nil {
		h.log.Error("failed to copy file content", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writer.Close()

	apiURL := h.apiURL + "/api/detect"
	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		h.log.Error("failed to create API request", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := h.client.Do(req)
	if err != nil {
		h.log.Error("detect API call failed", "error", err)
		http.Error(w, "Detection service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		h.log.Error("detection failed", "status", resp.Status, "body", string(body))
		http.Error(w, "Detection failed", http.StatusInternalServerError)
		return
	}

	var result struct {
		Comics []struct {
			ID    int     `json:"id"`
			URL   string  `json:"url"`
			Score float64 `json:"score"`
		} `json:"comics"`
		Total int `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		h.log.Error("failed to decode detection results", "error", err)
		http.Error(w, "Invalid response format", http.StatusInternalServerError)
		return
	}

	searchTime := time.Since(startTime)
	data := struct {
		Phrase         string
		IsImageResults bool
		Total          int
		Comics         []Comic
		SearchTime     string
		Limit          string
		Fast           bool
	}{
		Phrase:         "Image search",
		IsImageResults: true,
		Total:          result.Total,
		Comics:         make([]Comic, len(result.Comics)),
		SearchTime:     fmt.Sprintf("%.2fms", float64(searchTime.Microseconds())/1000),
		Limit:          "10",
		Fast:           false,
	}

	for i, c := range result.Comics {
		data.Comics[i] = Comic{
			ID:    c.ID,
			URL:   c.URL,
			Score: int(c.Score * 100),
		}
	}

	if err := h.templates.ExecuteTemplate(w, "results.html", data); err != nil {
		h.log.Error("failed to render results", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	query := r.URL.Query().Get("phrase")
	if query == "" {
		query = r.URL.Query().Get("q")
		if query == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	limit := r.URL.Query().Get("limit")
	fastSearch := r.URL.Query().Get("fast") == "true"
	isImageResults := r.URL.Query().Get("image_results") == "true"

	endpoint := "/api/search"
	if fastSearch {
		endpoint = "/api/isearch"
	}

	apiURL := h.apiURL + endpoint + "?phrase=" + url.QueryEscape(query)
	if limit != "" {
		apiURL += "&limit=" + url.QueryEscape(limit)
	}

	resp, err := h.client.Get(apiURL)
	if err != nil {
		h.log.Error("search API call failed",
			"url", apiURL,
			"error", err)
		http.Error(w, "Search service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Comics []struct {
			ID    int     `json:"id"`
			URL   string  `json:"url"`
			Score float64 `json:"score"`
		} `json:"comics"`
		Total int `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		h.log.Error("failed to decode search results", "error", err)
		http.Error(w, "Invalid response format", http.StatusInternalServerError)
		return
	}

	searchTime := time.Since(startTime)

	data := struct {
		Phrase         string
		Limit          string
		Fast           bool
		Total          int
		Comics         []Comic
		SearchTime     string
		IsImageResults bool
	}{
		Phrase:         query,
		Limit:          limit,
		Fast:           fastSearch,
		Total:          result.Total,
		Comics:         make([]Comic, len(result.Comics)),
		SearchTime:     fmt.Sprintf("%.2fms", float64(searchTime.Microseconds())/1000),
		IsImageResults: isImageResults,
	}

	for i, c := range result.Comics {
		data.Comics[i] = Comic{
			ID:    c.ID,
			URL:   c.URL,
			Score: int(c.Score * 100),
		}
	}

	if err := h.templates.ExecuteTemplate(w, "results.html", data); err != nil {
		h.log.Error("failed to render results", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) Admin(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("admin_token")
	if err != nil || token.Value == "" {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	status, err := h.getStatus(token.Value)
	if err != nil {
		h.log.Error("failed to get status", "error", err)
		http.Error(w, "Failed to get status", http.StatusInternalServerError)
		return
	}

	stats, err := h.getStats(token.Value)
	if err != nil {
		h.log.Error("failed to get stats", "error", err)
		http.Error(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	data := struct {
		Success string
		Status  UpdateStatus
		Stats   UpdateStats
	}{
		Success: r.URL.Query().Get("success"),
		Status:  status,
		Stats:   stats,
	}

	if err := h.templates.ExecuteTemplate(w, "admin.html", data); err != nil {
		h.log.Error("failed to render admin panel", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) getStatus(token string) (UpdateStatus, error) {
	req, err := http.NewRequest("GET", h.apiURL+"/api/db/status", nil)
	if err != nil {
		return UpdateStatus{}, err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return UpdateStatus{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UpdateStatus{}, fmt.Errorf("status request failed with code %d", resp.StatusCode)
	}

	var status UpdateStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return UpdateStatus{}, err
	}

	return status, nil
}

func (h *Handler) getStats(token string) (UpdateStats, error) {
	req, err := http.NewRequest("GET", h.apiURL+"/api/db/stats", nil)
	if err != nil {
		return UpdateStats{}, err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return UpdateStats{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UpdateStats{}, fmt.Errorf("stats request failed with code %d", resp.StatusCode)
	}

	var stats UpdateStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return UpdateStats{}, err
	}

	return stats, nil
}

func (h *Handler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	err := h.templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
		"Error": r.URL.Query().Get("error") == "1",
	})
	if err != nil {
		h.log.Error("failed to render template", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) AdminLoginPost(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	type LoginRequest struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	loginData := LoginRequest{
		Name:     username,
		Password: password,
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		h.log.Error("Failed to encode login data", "error", err)
		http.Redirect(w, r, "/admin/login?error=1", http.StatusSeeOther)
		return
	}

	req, err := http.NewRequest("POST", h.apiURL+"/api/login", bytes.NewBuffer(jsonData))
	if err != nil {
		h.log.Error("Failed to create request", "error", err)
		http.Redirect(w, r, "/admin/login?error=1", http.StatusSeeOther)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		h.log.Error("Login API call failed", "error", err)
		http.Redirect(w, r, "/admin/login?error=1", http.StatusSeeOther)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.log.Warn("Login failed", "status", resp.StatusCode)
		http.Redirect(w, r, "/admin/login?error=1", http.StatusSeeOther)
		return
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		h.log.Error("Failed to read token", "error", err)
		http.Redirect(w, r, "/admin/login?error=1", http.StatusSeeOther)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "admin_token",
		Value:    string(token),
		Path:     "/admin",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   120,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("admin_token")
	if err != nil {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	req, err := http.NewRequest("POST", h.apiURL+"/api/db/update", nil)
	if err != nil {
		h.log.Error("failed to create update request", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	req.Header.Add("Authorization", "Token "+token.Value)

	resp, err := h.client.Do(req)
	if err != nil {
		h.log.Error("update API call failed", "error", err)
		http.Error(w, "Update service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		h.log.Error("update failed", "status", resp.Status, "body", string(body))
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin?success=update", http.StatusSeeOther)
}

func (h *Handler) AdminDrop(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("admin_token")
	if err != nil {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
		return
	}

	req, err := http.NewRequest("DELETE", h.apiURL+"/api/db", nil)
	if err != nil {
		h.log.Error("failed to create update request", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	req.Header.Add("Authorization", "Token "+token.Value)

	resp, err := h.client.Do(req)
	if err != nil {
		h.log.Error("drop API call failed", "error", err)
		http.Error(w, "drop service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		h.log.Error("drop failed", "status", resp.Status, "body", string(body))
		http.Error(w, "drop failed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin?success=drop", http.StatusSeeOther)
}
