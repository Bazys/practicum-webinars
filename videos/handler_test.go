package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jackc/pgx/v5/pgtype"
)

// mockVideoRepository — мок интерфейса VideoRepository для тестов хендлера.
//
// Обратите внимание: мы мокаем ДОМЕННЫЙ интерфейс VideoRepository, а не драйвер БД.
// Это корректный unit-тест: проверяет, что handler правильно вызывает репозиторий
// и корректно отображает ошибки в HTTP-статусы.
//
// Так тесты(handler) никак не зависят от SQL и pgx — они быстрые и стабильные.
type mockVideoRepository struct {
	mock.Mock
}

func (m *mockVideoRepository) List(ctx context.Context, limit int) ([]Video, error) {
	args := m.Called(ctx, limit)
	v, _ := args.Get(0).([]Video)
	return v, args.Error(1)
}

func (m *mockVideoRepository) Get(ctx context.Context, id string) (Video, error) {
	args := m.Called(ctx, id)
	v, _ := args.Get(0).(Video)
	return v, args.Error(1)
}

func (m *mockVideoRepository) Create(ctx context.Context, title string, metadata MetadataFields) (Video, error) {
	args := m.Called(ctx, title, metadata)
	v, _ := args.Get(0).(Video)
	return v, args.Error(1)
}

func (m *mockVideoRepository) AddView(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

// newTestAPI — собирает api с моком и готовый httptest.Server.
func newTestAPI(t *testing.T) (*httptest.Server, *mockVideoRepository) {
	t.Helper()
	repo := new(mockVideoRepository)
	srv := httptest.NewServer(NewAPI(repo).Routes())
	t.Cleanup(srv.Close)
	return srv, repo
}

func TestAPI_ListVideos(t *testing.T) {
	srv, repo := newTestAPI(t)

	expected := []Video{
		{ID: pgtype.UUID{Bytes: [16]byte{1}, Valid: true}, Title: "Go 101", Views: 10},
	}
	// Ожидаем, что handler вызовет List с дефолтным лимитом.
	repo.On("List", mock.Anything, defaultLimit).Return(expected, nil)

	resp, err := http.Get(srv.URL + "/videos")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got struct {
		Videos []Video `json:"videos"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Len(t, got.Videos, 1)
	assert.Equal(t, "Go 101", got.Videos[0].Title)

	repo.AssertExpectations(t)
}

// Тест проверяет, что доменная ошибка ErrVideoNotFound
// превращается в 404, а не в 500.
func TestAPI_GetVideo_NotFound(t *testing.T) {
	srv, repo := newTestAPI(t)
	repo.On("Get", mock.Anything, "missing").Return(Video{}, ErrVideoNotFound)

	resp, err := http.Get(srv.URL + "/videos/missing")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// Тест проверяет счастливый путь создания видео.
func TestAPI_CreateVideo(t *testing.T) {
	srv, repo := newTestAPI(t)

	created := Video{
		ID:    pgtype.UUID{Bytes: [16]byte{9}, Valid: true},
		Title: "Контекст в Go",
		Views: 0,
	}
	repo.On("Create", mock.Anything, "Контекст в Go", MetadataFields{Author: "Rafael"}).
		Return(created, nil)

	body, _ := json.Marshal(map[string]any{
		"title":    "Контекст в Go",
		"metadata": MetadataFields{Author: "Rafael"},
	})
	resp, err := http.Post(srv.URL+"/videos", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	repo.AssertExpectations(t)
}

// Техническая ошибка репозитория должна давать 500.
func TestAPI_GetVideo_InternalError(t *testing.T) {
	srv, repo := newTestAPI(t)
	repo.On("Get", mock.Anything, "boom").Return(Video{}, errors.New("db is down"))

	resp, err := http.Get(srv.URL + "/videos/boom")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
