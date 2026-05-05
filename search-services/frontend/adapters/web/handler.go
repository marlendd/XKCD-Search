package web

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"html/template"
	"time"

	"yadro.com/course/frontend/core"
)

type Handler struct {
	api      core.APIClient
	log      *slog.Logger
	tokenTTL time.Duration
}

func NewHandler(api core.APIClient, log *slog.Logger, tokenTTL time.Duration) *Handler {
	return &Handler{
		api:      api,
		log:      log,
		tokenTTL: tokenTTL,
	}
}

type searchPageData struct {
	Phrase     string
	Limit      int
	Mode       string
	Comics     []core.Comic
	Total      int
	HasResults bool
	Error      string
}
type loginPageData struct {
	Name  string
	Error string
}
type adminPageData struct {
	Status core.UpdateStatus
	Stats  core.UpdateStatsResponse
	Error  string
}

func (h *Handler) SearchPage(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, "search.html", searchPageData{
		Limit: 10,
		Mode:  "search",
	})
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.log.ErrorContext(r.Context(), "failed to parse search form", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	phrase := strings.TrimSpace(r.FormValue("phrase"))
	mode := strings.TrimSpace(r.FormValue("mode"))
	if mode == "" {
		mode = "search"
	}

	limit := 10
	if raw := strings.TrimSpace(r.FormValue("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			h.renderTemplate(w, "search.html", searchPageData{
				Phrase: phrase,
				Limit:  limit,
				Mode:   mode,
				Error:  "Некорректный limit",
			})
			return
		}
		limit = parsed
	}

	if phrase == "" {
		h.renderTemplate(w, "search.html", searchPageData{
			Limit: limit,
			Mode:  mode,
			Error: "Введите поисковую фразу",
		})
		return
	}

	var (
		resp core.SearchResponse
		err  error
	)

	if mode == "isearch" {
		resp, err = h.api.ISearch(r.Context(), phrase, limit)
	} else {
		resp, err = h.api.Search(r.Context(), phrase, limit)
	}

	if err != nil {
		h.log.ErrorContext(r.Context(), "search request failed", "error", err, "mode", mode)
		h.renderTemplate(w, "search.html", searchPageData{
			Phrase: phrase,
			Limit:  limit,
			Mode:   mode,
			Error:  "Не удалось выполнить поиск",
		})
		return
	}

	h.renderTemplate(w, "search.html", searchPageData{
		Phrase:     phrase,
		Limit:      limit,
		Mode:       mode,
		Comics:     resp.Comics,
		Total:      resp.Total,
		HasResults: true,
	})
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, "login.html", loginPageData{})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.log.ErrorContext(r.Context(), "failed to parse login form", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	password := r.FormValue("password")
	if name == "" || password == "" {
		h.renderTemplate(w, "login.html", loginPageData{
			Name:  name,
			Error: "Введите имя и пароль",
		})
		return
	}
	token, err := h.api.Login(r.Context(), name, password)
	if err != nil {
		h.log.ErrorContext(r.Context(), "failed to login", "error", err)
		h.renderTemplate(w, "login.html", loginPageData{
			Name:  name,
			Error: "Неверные учетные данные",
		})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     tokenCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(h.tokenTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) AdminPage(w http.ResponseWriter, r *http.Request) {
	if _, err := h.readToken(r); err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	statusResp, err := h.api.Status(r.Context())
	if err != nil {
		h.log.ErrorContext(r.Context(), "failed to load status", "error", err)
		h.renderTemplate(w, "admin.html", adminPageData{
			Error: "Не удалось получить статус",
		})
		return
	}
	statsResp, err := h.api.Stats(r.Context())
	if err != nil {
		h.log.ErrorContext(r.Context(), "failed to load stats", "error", err)
		h.renderTemplate(w, "admin.html", adminPageData{
			Status: statusResp.Status,
			Error:  "Не удалось получить статистику",
		})
		return
	}
	h.renderTemplate(w, "admin.html", adminPageData{
		Status: statusResp.Status,
		Stats:  statsResp,
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	token, err := h.readToken(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := h.api.Update(r.Context(), token); err != nil {
		h.log.ErrorContext(r.Context(), "failed to trigger update", "error", err)
		h.renderAdminWithError(w, r, "Не удалось запустить update")
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) Drop(w http.ResponseWriter, r *http.Request) {
	token, err := h.readToken(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := h.api.Drop(r.Context(), token); err != nil {
		h.log.ErrorContext(r.Context(), "failed to drop index", "error", err)
		h.renderAdminWithError(w, r, "Не удалось выполнить drop")
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) renderAdminWithError(w http.ResponseWriter, r *http.Request, errMsg string) {
	data := adminPageData{Error: errMsg}

	statusResp, err := h.api.Status(r.Context())
	if err != nil {
		h.log.ErrorContext(r.Context(), "failed to load status for error page", "error", err)
		h.renderTemplate(w, "admin.html", data)
		return
	}
	data.Status = statusResp.Status

	statsResp, err := h.api.Stats(r.Context())
	if err != nil {
		h.log.ErrorContext(r.Context(), "failed to load stats for error page", "error", err)
		h.renderTemplate(w, "admin.html", data)
		return
	}
	data.Stats = statsResp

	h.renderTemplate(w, "admin.html", data)
}

func (h *Handler) renderTemplate(w http.ResponseWriter, name string, data any) {
	tmplPath := filepath.Join("adapters", "web", "templates", name)
	tmpl, err := template.New(name).Funcs(template.FuncMap{
		"contains": strings.Contains,
	}).ParseFiles(tmplPath)
	if err != nil {
		h.log.Error("failed to parse template", "error", err, "template", name)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		h.log.Error("failed to execute template", "error", err, "template", name)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) readToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie(tokenCookieName)
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(cookie.Value)
	if token == "" {
		return "", http.ErrNoCookie
	}
	return token, nil
}