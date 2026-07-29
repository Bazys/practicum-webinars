package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

// defaultLimit — сколько видео отдаём по умолчанию.
const defaultLimit = 30

// queryTimeout — общий таймаут на один запрос к БД из хендлера.
// Контекст с дедлайном НЕОБХОДИМ: иначе "зависший" запрос держит соединение
// из пула неограниченно долго и под нагрузкой пул истощается.
const queryTimeout = 3 * time.Second

// api — HTTP-слой. Зависит от ИНТЕРФЕЙСА VideoRepository, а не от конкретной
// реализации поверх pgx. Поэтому хендлер тестируется на моке без БД.
type api struct {
	repo VideoRepository
}

func NewAPI(repo VideoRepository) *api {
	return &api{repo: repo}
}

// Routes возвращает http.ServeMux с привязанными хендлерами.
// Вынесено отдельно, чтобы в тестах можно было поднять сервер на httptest.
func (a *api) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /videos", a.listVideos)
	mux.HandleFunc("POST /videos", a.createVideo)
	mux.HandleFunc("GET /videos/{id}", a.getVideo)
	mux.HandleFunc("POST /videos/{id}/view", a.addView)
	return mux
}

// listVideos: GET /videos?limit=N — список популярных видео.
//
// Демонстрируем таймаут: производный контекст с дедлайном передаётся в репозиторий,
// а оттуда — в pgx. Если БД не успела за queryTimeout, запрос отменится.
func (a *api) listVideos(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), queryTimeout)
	defer cancel()

	limit := defaultLimit

	videos, err := a.repo.List(ctx, limit)
	if err != nil {
		// Если клиент отвалился или таймаут — не пишем в ответ, просто логируем.
		log.Printf("list videos: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"videos": videos})
}

// getVideo: GET /videos/{id} — одно видео.
func (a *api) getVideo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), queryTimeout)
	defer cancel()

	id := r.PathValue("id")

	v, err := a.repo.Get(ctx, id)
	switch {
	case errors.Is(err, ErrVideoNotFound):
		writeError(w, http.StatusNotFound, "video not found")
		return
	case err != nil:
		log.Printf("get video %q: %v", id, err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, v)
}

// createVideo: POST /videos — создать видео.
func (a *api) createVideo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), queryTimeout)
	defer cancel()

	var req struct {
		Title    string         `json:"title"`
		Metadata MetadataFields `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	v, err := a.repo.Create(ctx, req.Title, req.Metadata)
	if err != nil {
		log.Printf("create video: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, v)
}

// addView: POST /videos/{id}/view — увеличить счётчик просмотров.
func (a *api) addView(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), queryTimeout)
	defer cancel()

	id := r.PathValue("id")

	err := a.repo.AddView(ctx, id)
	switch {
	case errors.Is(err, ErrVideoNotFound):
		writeError(w, http.StatusNotFound, "video not found")
		return
	case err != nil:
		log.Printf("add view %q: %v", id, err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("write json: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
