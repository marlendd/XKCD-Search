package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"yadro.com/course/frontend/core"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	log        *slog.Logger
}

func NewClient(baseURL string, timeout time.Duration, log *slog.Logger) (*Client, error) {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		log: log,
	}, nil
}

type loginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (c *Client) Login(ctx context.Context, user, password string) (string, error) {
	data, _, err := c.doRaw(ctx, "POST", "/api/login", "", &loginRequest{
		Name:     user,
		Password: password,
	})
	if err != nil {
		c.log.ErrorContext(ctx, "failed to login", "error", err)
		return "", err
	}

	token := strings.TrimSpace(string(data))
	return token, nil
}

func (c *Client) Ping(ctx context.Context) (core.PingResponse, error) {
	resp := &core.PingResponse{
		Replies: make(map[string]string),
	}
	if err := c.doJSON(ctx, "GET", "/api/ping", "", nil, resp); err != nil {
		c.log.ErrorContext(ctx, "failed to ping", "error", err)
		return core.PingResponse{}, err
	}
	return *resp, nil
}

func (c *Client) Search(ctx context.Context, phrase string, limit int) (core.SearchResponse, error) {
	var resp core.SearchResponse

	q := url.Values{}
	q.Set("phrase", phrase)
	q.Set("limit", strconv.Itoa(limit))

	path := "/api/search?" + q.Encode()

	if err := c.doJSON(ctx, "GET", path, "", nil, &resp); err != nil {
		c.log.ErrorContext(ctx, "failed to search", "error", err)
		return core.SearchResponse{}, err
	}
	return resp, nil
}

func (c *Client) ISearch(ctx context.Context, phrase string, limit int) (core.SearchResponse, error) {
	var resp core.SearchResponse

	q := url.Values{}
	q.Set("phrase", phrase)
	q.Set("limit", strconv.Itoa(limit))

	path := "/api/isearch?" + q.Encode()

	if err := c.doJSON(ctx, "GET", path, "", nil, &resp); err != nil {
		c.log.ErrorContext(ctx, "failed to isearch", "error", err)
		return core.SearchResponse{}, err
	}
	return resp, nil
}

func (c *Client) Update(ctx context.Context, token string) error {
	// Frontend should not block on a long-running update job.
	if err := c.doJSON(ctx, "POST", "/api/db/update?async=1", token, nil, nil); err != nil {
		c.log.ErrorContext(ctx, "failed to update", "error", err)
		return err
	}
	return nil
}

func (c *Client) Stats(ctx context.Context) (core.UpdateStatsResponse, error) {
	var resp core.UpdateStatsResponse
	if err := c.doJSON(ctx, "GET", "/api/db/stats", "", nil, &resp); err != nil {
		c.log.ErrorContext(ctx, "failed to get stats", "error", err)
		return core.UpdateStatsResponse{}, err
	}
	return resp, nil
}

func (c *Client) Status(ctx context.Context) (core.StatusResponse, error) {
	var resp core.StatusResponse
	if err := c.doJSON(ctx, "GET", "/api/db/status", "", nil, &resp); err != nil {
		c.log.ErrorContext(ctx, "failed to get status", "error", err)
		return core.StatusResponse{}, err
	}
	return resp, nil
}

func (c *Client) Drop(ctx context.Context, token string) error {
	if err := c.doJSON(ctx, "DELETE", "/api/db", token, nil, nil); err != nil {
		c.log.ErrorContext(ctx, "failed to delete", "error", err)
		return err
	}
	return nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, token string, reqBody any, respBody any) error {
	var body io.Reader
	if reqBody != nil {
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(reqBody); err != nil {
			c.log.ErrorContext(ctx, "failed to encode json", "error", err)
			return err
		}
		body = buf
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		c.log.ErrorContext(ctx, "failed to create new request", "error", err)
		return err
	}

	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.ErrorContext(ctx, "failed to perform request", "error", err)
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.ErrorContext(ctx, "failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			if errors.Is(err, io.EOF) {
				c.log.DebugContext(ctx, "empty body", "error", err)
				return nil
			}
			c.log.ErrorContext(ctx, "failed to decode json", "error", err)
			return err
		}
	}

	return nil
}

func (c *Client) doRaw(ctx context.Context, method, path string, token string, reqBody any) ([]byte, int, error) {
	var body io.Reader
	if reqBody != nil {
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(reqBody); err != nil {
			c.log.ErrorContext(ctx, "failed to encode json", "error", err)
			return nil, 0, err
		}
		body = buf
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		c.log.ErrorContext(ctx, "failed to create new request", "error", err)
		return nil, 0, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.ErrorContext(ctx, "failed to perform request", "error", err)
		return nil, 0, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.ErrorContext(ctx, "failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.ErrorContext(ctx, "failed to read response body", "error", err)
		return nil, resp.StatusCode, err
	}

	return data, resp.StatusCode, nil
}