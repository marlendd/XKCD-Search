package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

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

type Client struct {
	log    *slog.Logger
	client http.Client
	url    string
}

func NewClient(url string, timeout time.Duration, log *slog.Logger) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("empty base url specified")
	}
	return &Client{
		client: http.Client{Timeout: timeout},
		log:    log,
		url:    url,
	}, nil
}

func (c Client) Get(ctx context.Context, id int) (core.XKCDInfo, error) {
	var getUrl string
	if id == 0 {
		getUrl = c.url + `/info.0.json`
	} else {
		getUrl = c.url + `/` + strconv.Itoa(id) + `/info.0.json`
	}

	req, err := http.NewRequestWithContext(ctx, "GET", getUrl, nil)
	if err != nil {
		c.log.Error("failed to create GET request", "error", err)
		return core.XKCDInfo{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error("failed to GET", "error", err)
		return core.XKCDInfo{}, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Error("failed to close body", "error", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		c.log.Error("recieved 404")
		return core.XKCDInfo{}, core.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		c.log.Error("recieved unexpected status")
		return core.XKCDInfo{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	data := &getResponse{}
	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		c.log.Error("failed to unmarshall GET json")
		return core.XKCDInfo{}, err
	}

	return core.XKCDInfo{
		ID:          data.Num,
		URL:         data.Img,
		Title:       data.Title,
		SafeTitle:   data.SafeTitle,
		Transcript:  data.Transcript,
		Description: data.Alt,
	}, nil
}

func (c Client) LastID(ctx context.Context) (int, error) {
	info, err := c.Get(ctx, 0)
	if err != nil {
		c.log.Error("failed to get last id")
		return 0, err
	}
	return info.ID, nil
}
