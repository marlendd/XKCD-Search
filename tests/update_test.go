package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type UpdateStats struct {
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
	ComicsFetched int `json:"comics_fetched"`
	ComicsTotal   int `json:"comics_total"`
}

type UpdateStatus struct {
	Status string `json:"status"`
}

func TestEmptyDB(t *testing.T) {
	prepare(t)
}

func TestUpdate(t *testing.T) {
	prepare(t)
	var wg sync.WaitGroup
	wg.Add(3)
	var err1, err2, err3 error
	var res1, res2 int
	var res3 string
	token := login(t)
	go func() {
		res1, err1 = update(token)
		wg.Done()
	}()
	go func() {
		res2, err2 = update(token)
		wg.Done()
	}()
	go func() {
		time.Sleep(1 * time.Second)
		res3, err3 = status()
		wg.Done()
	}()
	wg.Wait()
	require.NoError(t, err1, "error from update")
	require.NoError(t, err2, "error from update")
	require.NoError(t, err3, "erorr from status")
	require.True(t,
		res1 == http.StatusOK && res2 == http.StatusAccepted ||
			res2 == http.StatusOK && res1 == http.StatusAccepted,
		"wrong statuses from concurrent updates, expect ok && accepted",
	)
	require.Contains(t, []string{"running", "idle"}, res3, "unexpected status value")
	waitForIdle(t, 8*time.Minute)
	st := stats(t)
	require.Equal(t, st.ComicsTotal, st.ComicsFetched)
	require.True(t, st.ComicsTotal > 3000, "there are more than 3000 comics in XKCD")
	require.True(t, 1000 < st.WordsTotal, "not enough total words in DB")
	require.True(t, 100 < st.WordsUnique, "not enough unique words in DB")

	prepare(t)
}

func login(t *testing.T) string {
	data := bytes.NewBufferString(`{"name":"admin", "password":"password"}`)
	req, err := http.NewRequest(http.MethodPost, address+"/api/login", data)
	require.NoError(t, err, "cannot make request")
	resp, err := client.Do(req)
	require.NoError(t, err, "could not send login command")
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	token, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(token)
}

func prepare(t *testing.T) {
	token := login(t)
	dropped := false
	for range 5 {
		req, err := http.NewRequest(http.MethodDelete, address+"/api/db", nil)
		require.NoError(t, err, "cannot make request")
		req.Header.Add("Authorization", "Token "+token)
		resp, err := client.Do(req)
		require.NoError(t, err, "could not send clean up command")
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			dropped = true
			break
		}
		require.Equal(t, http.StatusConflict, resp.StatusCode)
		waitForIdle(t, 8*time.Minute)
	}
	require.True(t, dropped, "failed to clean up DB")

	updateStats := stats(t)
	require.Equal(t, 0, updateStats.ComicsFetched)
	require.True(t, updateStats.ComicsTotal > 3000, "there are more than 3000 comics in XKCD")
	require.Equal(t, 0, updateStats.WordsTotal)
	require.Equal(t, 0, updateStats.WordsUnique)
	updateStatus, err := status()
	require.Equal(t, "idle", updateStatus, err)
}

func waitForIdle(t *testing.T, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		st, err := status()
		require.NoError(t, err, "error from status")
		if st == "idle" {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for idle status after %s", timeout)
}

// this must not contain t because it runs in a waited goroutine
func update(token string) (int, error) {
	req, err := http.NewRequest(http.MethodPost, address+"/api/db/update", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Add("Authorization", "Token "+token)
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}

// this must not contain t because it runs in a waited goroutine
func status() (string, error) {
	resp, err := client.Get(address + "/api/db/status")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if http.StatusOK != resp.StatusCode {
		return "", fmt.Errorf("http status: %v", resp.Status)
	}
	var status UpdateStatus
	if err = json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return "", fmt.Errorf("could not decode: %v", err)
	}
	return status.Status, nil
}

func stats(t *testing.T) UpdateStats {
	resp, err := client.Get(address + "/api/db/stats")
	require.NoError(t, err, "could not get stats")
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var stats UpdateStats
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&stats), "cannot decode")
	return stats
}
