package api_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type Comics struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type ComicsReply struct {
	Comics []Comics `json:"comics"`
	Total  int      `json:"total"`
}

func TestSearch(t *testing.T) {
	token := login(t)
	_, err := update(token)
	require.NoError(t, err, "could not run update")
	t.Run("no phrase", SearchNoPhrase)
	t.Run("bad limit minus", SearchBadLimitMinus)
	t.Run("bad limit alpha", SearchBadLimitAlpha)
	t.Run("search limit 2", SearchLimit2)
	t.Run("search limit default", SearchLimitDefault)
	t.Run("search phrases", SearchPhrases)
	t.Run("index search", IndexSearchPhrases)
}

func SearchNoPhrase(t *testing.T) {
	resp, err := client.Get(address + "/api/search")
	require.NoError(t, err, "failed to search")
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, "need bad request")
}

func SearchBadLimitMinus(t *testing.T) {
	resp, err := client.Get(address + "/api/search?limit=-1")
	require.NoError(t, err, "failed to search")
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, "need bad request")
}

func SearchBadLimitAlpha(t *testing.T) {
	resp, err := client.Get(address + "/api/search?limit=asdf")
	require.NoError(t, err, "failed to search")
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, "need bad request")
}

func SearchLimit2(t *testing.T) {
	resp, err := client.Get(address + "/api/search?limit=2&phrase=linux")
	require.NoError(t, err, "failed to search")
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "need OK status")
	var comics ComicsReply
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&comics), "decode failed")
	require.Equal(t, 2, comics.Total)
	require.Equal(t, 2, len(comics.Comics))
}

func SearchLimitDefault(t *testing.T) {

	resp, err := client.Get(address + "/api/search?phrase=linux")
	require.NoError(t, err, "failed to search")
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "need OK status")
	var comics ComicsReply
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&comics), "decode failed")
	require.Equal(t, 10, comics.Total)
	require.Equal(t, 10, len(comics.Comics))
}

func SearchPhrases(t *testing.T) {
	testCases := []struct {
		phrase string
		url    string
	}{
		{
			phrase: "linux+cpu+video+machine+русские+хакеры",
			url:    "https://imgs.xkcd.com/comics/supported_features.png",
		},
		{
			phrase: "Binary Christmas Tree",
			url:    "https://imgs.xkcd.com/comics/tree.png",
		},
		{
			phrase: "apple a day -> keeps doctors away",
			url:    "https://imgs.xkcd.com/comics/an_apple_a_day.png",
		},
		{
			phrase: "mines, captcha",
			url:    "https://imgs.xkcd.com/comics/mine_captcha.png",
		},
		{
			phrase: "newton apple's idea",
			url:    "https://imgs.xkcd.com/comics/inspiration.png",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.phrase, func(t *testing.T) {
			resp, err := client.Get(address + "/api/search?phrase=" + url.QueryEscape(tc.phrase))
			require.NoError(t, err, "failed to search")
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode, "need OK status")
			var comics ComicsReply
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&comics), "decode failed")
			urls := make([]string, 0, len(comics.Comics))
			for _, c := range comics.Comics {
				urls = append(urls, c.URL)
			}
			require.Containsf(t, urls, tc.url, "could not find %q", tc.phrase)
		})
	}
}

func IndexSearchPhrases(t *testing.T) {
	// clean DB and wait 30 sec for index update
	prepare(t)
	time.Sleep(30 * time.Second)
	resp, err := client.Get(address + "/api/isearch?phrase=linux")
	require.NoError(t, err, "failed to search")
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "need OK status")
	var comics ComicsReply
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&comics), "decode failed")
	require.Equal(t, 0, comics.Total)
	require.Equal(t, 0, len(comics.Comics))

	// update DB and wait 30 sec for index update
	token := login(t)
	_, err = update(token)
	require.NoError(t, err, "could not run update")
	time.Sleep(30 * time.Second)

	testCases := []struct {
		phrase string
		url    string
	}{
		{
			phrase: "linux+cpu+video+machine+русские+хакеры",
			url:    "https://imgs.xkcd.com/comics/supported_features.png",
		},
		{
			phrase: "Binary Christmas Tree",
			url:    "https://imgs.xkcd.com/comics/tree.png",
		},
		{
			phrase: "apple a day -> keeps doctors away",
			url:    "https://imgs.xkcd.com/comics/an_apple_a_day.png",
		},
		{
			phrase: "mines, captcha",
			url:    "https://imgs.xkcd.com/comics/mine_captcha.png",
		},
		{
			phrase: "newton apple's idea",
			url:    "https://imgs.xkcd.com/comics/inspiration.png",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.phrase, func(t *testing.T) {
			resp, err := client.Get(address + "/api/isearch?phrase=" + url.QueryEscape(tc.phrase))
			require.NoError(t, err, "failed to search")
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode, "need OK status")
			var comics ComicsReply
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&comics), "decode failed")
			urls := make([]string, 0, len(comics.Comics))
			for _, c := range comics.Comics {
				urls = append(urls, c.URL)
			}
			require.Containsf(t, urls, tc.url, "could not find %q", tc.phrase)
		})
	}
}
