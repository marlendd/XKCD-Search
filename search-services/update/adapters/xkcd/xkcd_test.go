package xkcd_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"yadro.com/course/update/adapters/xkcd"
	"yadro.com/course/update/core"
)

type getResponse struct {
	Num        int    `json:"num"`
	Title      string `json:"title"`
	SafeTitle  string `json:"safe_title"`
	Transcript string `json:"transcript"`
	Alt        string `json:"alt"`
	Img        string `json:"img"`
}

func newServer(t *testing.T, handler http.HandlerFunc) (*xkcd.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client, err := xkcd.NewClient(srv.URL, time.Second, slog.Default())
	assert.NoError(t, err)
	return client, srv
}

func validResponse(id int) getResponse {
	return getResponse{
		Num:        id,
		Title:      "Test",
		SafeTitle:  "test",
		Transcript: "transcript",
		Alt:        "alt",
		Img:        "http://example.com/img.png",
	}
}

func TestNewClient(t *testing.T) {
	t.Run("empty url", func(t *testing.T) {
		_, err := xkcd.NewClient("", time.Second, slog.Default())
		assert.Error(t, err)
	})
}

func TestGet(t *testing.T) {
	t.Run("id=0 uses info.0.json", func(t *testing.T) {
		client, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/info.0.json", r.URL.Path)
			assert.NoError(t, json.NewEncoder(w).Encode(validResponse(100)))
		})

		info, err := client.Get(context.Background(), 0)
		assert.NoError(t, err)
		assert.Equal(t, 100, info.ID)
	})

	t.Run("id>0 uses id path", func(t *testing.T) {
		client, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/42/info.0.json", r.URL.Path)
			assert.NoError(t, json.NewEncoder(w).Encode(validResponse(42)))
		})

		info, err := client.Get(context.Background(), 42)
		assert.NoError(t, err)
		assert.Equal(t, 42, info.ID)
	})

	t.Run("maps fields correctly", func(t *testing.T) {
		client, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.NoError(t, json.NewEncoder(w).Encode(getResponse{
				Num:        1,
				Title:      "title",
				SafeTitle:  "safe",
				Transcript: "transcript",
				Alt:        "desc",
				Img:        "http://img.url",
			}))
		})

		info, err := client.Get(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, core.XKCDInfo{
			ID:          1,
			URL:         "http://img.url",
			Title:       "title",
			SafeTitle:   "safe",
			Transcript:  "transcript",
			Description: "desc",
		}, info)
	})

	t.Run("404 returns ErrNotFound", func(t *testing.T) {
		client, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		_, err := client.Get(context.Background(), 1)
		assert.ErrorIs(t, err, core.ErrNotFound)
	})

	t.Run("500 returns error", func(t *testing.T) {
		client, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		_, err := client.Get(context.Background(), 1)
		assert.Error(t, err)
	})

	t.Run("bad json returns error", func(t *testing.T) {
		client, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("not json"))
			assert.NoError(t, err)
		})

		_, err := client.Get(context.Background(), 1)
		assert.Error(t, err)
	})
}

func TestLastID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.NoError(t, json.NewEncoder(w).Encode(validResponse(999)))
		})

		id, err := client.LastID(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 999, id)
	})

	t.Run("error", func(t *testing.T) {
		client, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		_, err := client.LastID(context.Background())
		assert.Error(t, err)
	})
}