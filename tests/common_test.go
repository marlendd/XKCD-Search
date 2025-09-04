package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const address = "http://localhost:28080"

var client = http.Client{
	Timeout: 10 * time.Minute,
}

func TestPreflight(t *testing.T) {
	require.Equal(t, true, true)
}

type PingResponse struct {
	Replies map[string]string `json:"replies"`
}

func TestPing(t *testing.T) {
	resp, err := client.Get(address + "/api/ping")
	require.NoError(t, err, "cannot ping")
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "wrong status")

	var reply PingResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&reply))
	require.Equal(t, "ok", reply.Replies["words"], "no words running")
	require.Equal(t, "ok", reply.Replies["update"], "no db running")
	require.Equal(t, "ok", reply.Replies["search"], "no search running")
}
