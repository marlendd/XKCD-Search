package rest

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"yadro.com/course/api/core"
	"yadro.com/course/api/mocks"
)

func TestNewLoginHandler(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	log := slog.Default()
	mockAuth := mocks.NewMockAuthenticator(c)

	t.Run("successfull login", func(t *testing.T) {
		mockAuth.EXPECT().
			Login("admin", "secret").
			Return("jwt-token-123", nil)

		handler := NewLoginHandler(log, mockAuth)
		body := strings.NewReader(`{"name":"admin","password":"secret"}`)
		req := httptest.NewRequest(http.MethodPost, "/login", body)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "jwt-token-123", rec.Body.String())
	})

	t.Run("unauthorized", func(t *testing.T) {
		mockAuth.EXPECT().
			Login("admin", "wrong").
			Return("", core.ErrNotAuthorized)

		handler := NewLoginHandler(log, mockAuth)
		body := strings.NewReader(`{"name":"admin","password":"wrong"}`)
		req := httptest.NewRequest(http.MethodPost, "/login", body)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("bad request", func(t *testing.T) {

		handler := NewLoginHandler(log, mockAuth)
		body := strings.NewReader(`bad json`)
		req := httptest.NewRequest(http.MethodPost, "/login", body)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestNewPingHandler(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	log := slog.Default()
	mockPinger := mocks.NewMockPinger(c)

	t.Run("all services available", func(t *testing.T) {
		mockPinger.EXPECT().
			Ping(gomock.Any()).
			Return(nil)

		handler := NewPingHandler(log, map[string]core.Pinger{
			"db": mockPinger,
		})
		req := httptest.NewRequest(http.MethodPost, "/ping", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"ok"`)
	})

	t.Run("one service unavailable", func(t *testing.T) {
		mockPinger1 := mocks.NewMockPinger(c)
		mockPinger2 := mocks.NewMockPinger(c)

		mockPinger1.EXPECT().Ping(gomock.Any()).Return(nil)
		mockPinger2.EXPECT().Ping(gomock.Any()).Return(errors.New("connection refused"))

		handler := NewPingHandler(log, map[string]core.Pinger{
			"pinger1": mockPinger1,
			"pinger2": mockPinger2,
		})
		req := httptest.NewRequest(http.MethodPost, "/ping", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"ok"`)
		assert.Contains(t, rec.Body.String(), `"unavailable"`)
	})
}

func TestNewSearchHandler(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	log := slog.Default()
	mockSearcher := mocks.NewMockSearcher(c)

	cases := []struct {
		name         string
		query        string
		setupMock    func()
		expectedCode int
	}{
		{
			name:  "successful search",
			query: "?phrase=golang&limit=5",
			setupMock: func() {
				mockSearcher.EXPECT().
					Search(gomock.Any(), "golang", 5).
					Return(core.SearchResult{
						Comics: []core.Comic{
							{ID: 1, URL: "testing"},
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "empty phrase",
			query:        "?phrase=",
			setupMock:    func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid limit",
			query:        "?phrase=linux&limit=five",
			setupMock:    func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:  "search error",
			query: "?phrase=golang",
			setupMock: func() {
				mockSearcher.EXPECT().
					Search(gomock.Any(), "golang", 10).
					Return(core.SearchResult{}, errors.New("search error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			handler := NewSearchHandler(log, mockSearcher)
			req := httptest.NewRequest(http.MethodPost, "/search"+tc.query, nil)
			rec := httptest.NewRecorder()

			handler(rec, req)

			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestNewISearchHandler(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	log := slog.Default()
	mockISearcher := mocks.NewMockISearcher(c)

	cases := []struct {
		name         string
		query        string
		setupMock    func()
		expectedCode int
	}{
		{
			name:  "successful isearch",
			query: "?phrase=golang&limit=5",
			setupMock: func() {
				mockISearcher.EXPECT().
					ISearch(gomock.Any(), "golang", 5).
					Return(core.SearchResult{
						Comics: []core.Comic{
							{ID: 1, URL: "testing"},
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "empty phrase",
			query:        "?phrase=",
			setupMock:    func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid limit",
			query:        "?phrase=linux&limit=five",
			setupMock:    func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:  "search error",
			query: "?phrase=golang",
			setupMock: func() {
				mockISearcher.EXPECT().
					ISearch(gomock.Any(), "golang", 10).
					Return(core.SearchResult{}, errors.New("searcher error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			handler := NewISearchHandler(log, mockISearcher)
			req := httptest.NewRequest(http.MethodPost, "/search"+tc.query, nil)
			rec := httptest.NewRecorder()

			handler(rec, req)

			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestNewUpdateHandler(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	log := slog.Default()
	mockUpdater := mocks.NewMockUpdater(c)

	t.Run("successful async update start", func(t *testing.T) {
		mockUpdater.EXPECT().
			Status(gomock.Any()).
			Return(core.StatusUpdateIdle, nil)

		called := make(chan struct{})
		mockUpdater.EXPECT().
			Update(gomock.Any()).
			DoAndReturn(func(_ context.Context) error {
				close(called)
				return nil
			})

		handler := NewUpdateHandler(log, mockUpdater)
		req := httptest.NewRequest(http.MethodPost, "/update", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		select {
		case <-called:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("background update was not started")
		}
	})

	t.Run("failed status", func(t *testing.T) {
		mockUpdater.EXPECT().
			Status(gomock.Any()).
			Return(core.StatusUpdateUnknown, errors.New("status error"))

		handler := NewUpdateHandler(log, mockUpdater)
		req := httptest.NewRequest(http.MethodPost, "/update", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("update already running", func(t *testing.T) {
		mockUpdater.EXPECT().
			Status(gomock.Any()).
			Return(core.StatusUpdateRunning, nil)

		handler := NewUpdateHandler(log, mockUpdater)
		req := httptest.NewRequest(http.MethodPost, "/update", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusAccepted, rec.Code)
	})
}

func TestNewUpdateStatsHandler(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	log := slog.Default()
	mockUpdater := mocks.NewMockUpdater(c)

	t.Run("successfull stats", func(t *testing.T) {
		mockUpdater.EXPECT().
			Stats(gomock.Any()).
			Return(core.UpdateStats{
				WordsTotal:    100,
				WordsUnique:   50,
				ComicsFetched: 10,
				ComicsTotal:   20,
			}, nil)

		handler := NewUpdateStatsHandler(log, mockUpdater)
		req := httptest.NewRequest(http.MethodPost, "/stats", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		var resp updateStatsResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, 100, resp.WordsTotal)
		assert.Equal(t, 50, resp.WordsUnique)
		assert.Equal(t, 10, resp.ComicsFetched)
		assert.Equal(t, 20, resp.ComicsTotal)
	})
	t.Run("bad stats", func(t *testing.T) {
		mockUpdater.EXPECT().
			Stats(gomock.Any()).
			Return(core.UpdateStats{}, errors.New("updater error"))

		handler := NewUpdateStatsHandler(log, mockUpdater)
		req := httptest.NewRequest(http.MethodPost, "/stats", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestNewUpdateStatusHandler(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	log := slog.Default()
	mockUpdater := mocks.NewMockUpdater(c)

	t.Run("successfull status", func(t *testing.T) {
		mockUpdater.EXPECT().
			Status(gomock.Any()).
			Return(core.StatusUpdateIdle, nil)

		handler := NewUpdateStatusHandler(log, mockUpdater)
		req := httptest.NewRequest(http.MethodPost, "/status", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"idle"`)
		
	})
	t.Run("bad status", func(t *testing.T) {
		mockUpdater.EXPECT().
			Status(gomock.Any()).
			Return(core.StatusUpdateUnknown, errors.New("updater error"))

		handler := NewUpdateStatusHandler(log, mockUpdater)
		req := httptest.NewRequest(http.MethodPost, "/status", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestNewDropHandler(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	log := slog.Default()
	mockUpdater := mocks.NewMockUpdater(c)

	t.Run("successfull drop", func(t *testing.T) {
		mockUpdater.EXPECT().
			Drop(gomock.Any()).
			Return(nil)

		handler := NewDropHandler(log, mockUpdater)
		req := httptest.NewRequest(http.MethodPost, "/drop", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
	t.Run("failed drop", func(t *testing.T) {
		mockUpdater.EXPECT().
			Drop(gomock.Any()).
			Return(errors.New("updater error"))

		handler := NewDropHandler(log, mockUpdater)
		req := httptest.NewRequest(http.MethodPost, "/drop", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
	t.Run("failed drop due to update is running", func(t *testing.T) {
		mockUpdater.EXPECT().
			Drop(gomock.Any()).
			Return(core.ErrAlreadyExists)

		handler := NewDropHandler(log, mockUpdater)
		req := httptest.NewRequest(http.MethodPost, "/drop", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)
	})
}