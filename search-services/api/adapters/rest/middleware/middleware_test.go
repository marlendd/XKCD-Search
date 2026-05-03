package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"yadro.com/course/api/adapters/rest/middleware"
	"yadro.com/course/api/core"
	"yadro.com/course/api/mocks"
)

func TestAuth(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	mockVerifier := mocks.NewMockTokenVerifier(c)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

	cases := []struct{
		name string
		header string
		setupMock func()
		expectedCode int
	}{
		{
			name: "no header",
			header: "",
			setupMock: func() {},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "invalid token",
			header: "jwt-token-123",
			setupMock: func() {},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "valid token",
			header: "Token jwt-token-123",
			setupMock: func() {
				mockVerifier.EXPECT().
				Verify("jwt-token-123").
				Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "unauthorized token",
			header: "Token not-token",
			setupMock: func() {
				mockVerifier.EXPECT().
				Verify("not-token").
				Return(core.ErrNotAuthorized)
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "internal error",
			header: "Token not-token",
			setupMock: func() {
				mockVerifier.EXPECT().
				Verify("not-token").
				Return(errors.New("some error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			handler := middleware.Auth(next, mockVerifier)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()

			handler(rec, req)
			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}

func TestConcurrency(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	t.Run("request within limit", func(t *testing.T) {
		handler := middleware.Concurrency(next, 2)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("request exceeds limit", func(t *testing.T) {
		handler := middleware.Concurrency(next, 1)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			handler(rec, req)
		}()

		time.Sleep(10 * time.Millisecond)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		wg.Wait()
	})
}

func TestWithMetrics(t *testing.T) {
	t.Run("sets status code via WriteHeader", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		})

		handler := middleware.WithMetrics(next)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("sets status 200 implicitly via Write", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("hello"))
			assert.NoError(t, err)
		})

		handler := middleware.WithMetrics(next)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("double WriteHeader ignored", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.WriteHeader(http.StatusInternalServerError) // должен игнорироваться
		})

		handler := middleware.WithMetrics(next)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestRate(t *testing.T) {
	t.Run("requests succeed", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware.Rate(next, 1000)

		for range 3 {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			handler(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})

	t.Run("rate limiting adds delay", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware.Rate(next, 2)

		start := time.Now()
		for range 3 {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			handler(rec, req)
		}

		assert.GreaterOrEqual(t, time.Since(start), 900*time.Millisecond)
	})
}