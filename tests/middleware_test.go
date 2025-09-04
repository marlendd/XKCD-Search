package api_test

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// 200 tests in packs of 20, with concurrency 10. 100 reqs must be ok, the rest - 503
func TestSearchConcurrency(t *testing.T) {
	const numPacks = 10
	const packSize = 20
	const concurrency = 10
	token := login(t)
	_, err := update(token)
	require.NoError(t, err, "could not update")
	var countOK atomic.Int64
	var countBusy atomic.Int64
	for range numPacks {
		var wg sync.WaitGroup
		wg.Add(packSize)
		for range packSize {
			go func() {
				defer wg.Done()
				resp, err := client.Get(address + "/api/search?phrase=linux")
				require.NoError(t, err, "failed to search")
				defer resp.Body.Close()
				switch resp.StatusCode {
				case http.StatusServiceUnavailable:
					countBusy.Add(1)
				case http.StatusOK:
					countOK.Add(1)
				}
			}()
		}
		wg.Wait()
	}
	require.True(t, int64(concurrency*numPacks) <= countOK.Load(), "need some http ok")
	require.True(t, int64(0) < countBusy.Load(), "need at least some http busy")
	require.Equal(t, int64(numPacks*packSize), countOK.Load()+countBusy.Load(),
		"need only ok and busy statuses")
}

func TestSearchRateLong(t *testing.T) {
	const rate = 100
	const numReq = 1000
	token := login(t)
	_, err := update(token)
	require.NoError(t, err, "could not update")
	time.Sleep(30 * time.Second)
	var wg sync.WaitGroup
	wg.Add(numReq)
	start := time.Now()
	for range numReq {
		go func() {
			defer wg.Done()
			resp, err := client.Get(address + "/api/isearch?phrase=linux")
			require.NoError(t, err, "failed to search")
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)
		}()
	}
	wg.Wait()
	duration := time.Since(start)
	actualRate := numReq / duration.Seconds()

	require.InDelta(t, rate, actualRate, rate/10)
}

func TestBadLogin(t *testing.T) {
	data := bytes.NewBufferString(`{"name":"user", "password":""}`)
	req, err := http.NewRequest(http.MethodPost, address+"/api/login", data)
	require.NoError(t, err, "cannot make request")
	resp, err := client.Do(req)
	require.NoError(t, err, "could not send login command")
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestBadPassword(t *testing.T) {
	data := bytes.NewBufferString(`{"name":"admin", "password":""}`)
	req, err := http.NewRequest(http.MethodPost, address+"/api/login", data)
	require.NoError(t, err, "cannot make request")
	resp, err := client.Do(req)
	require.NoError(t, err, "could not send login command")
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGoodLogin(t *testing.T) {
	data := bytes.NewBufferString(`{"name":"admin", "password":"password"}`)
	req, err := http.NewRequest(http.MethodPost, address+"/api/login", data)
	require.NoError(t, err, "cannot make request")
	resp, err := client.Do(req)
	require.NoError(t, err, "could not send login command")
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	token, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.True(t, len(token) > 0)
}

func TestLoginExpiredVeryLong(t *testing.T) {
	token := login(t)
	time.Sleep(125 * time.Second)
	req, err := http.NewRequest(http.MethodPost, address+"/api/db/update", nil)
	require.NoError(t, err, "cannot make request")
	req.Header.Add("Authorization", "Token "+token)
	resp, err := client.Do(req)
	require.NoError(t, err, "could not send update command")
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestUpdateDbNoToken(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, address+"/api/db/update", nil)
	require.NoError(t, err, "cannot make request")
	resp, err := client.Do(req)
	require.NoError(t, err, "could not send update command")
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestDropDbNoToken(t *testing.T) {
	req, err := http.NewRequest(http.MethodDelete, address+"/api/db", nil)
	require.NoError(t, err, "cannot make request")
	resp, err := client.Do(req)
	require.NoError(t, err, "could not send drop command")
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
