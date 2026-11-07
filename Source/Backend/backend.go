package Backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client calls the managed backend's Silicon API (API.md). Every call carries the shared API key;
// Silicon never touches the backend's database or files itself.
type Client struct {
	cfg  BackendConfig
	http *http.Client
}

func NewClient(cfg BackendConfig) *Client {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	return &Client{cfg: cfg, http: &http.Client{Timeout: timeout}}
}

func (c *Client) Name() string { return c.cfg.Name }

// Actor is the admin a call is made for. The backend uses it to name announcements.
type Actor struct {
	User string
	Role string
}

// Do sends method path (relative to /silicon/v1) with an optional JSON body and returns the raw
// answer, so the API layer can pass statuses such as 404/501 ("not supported") through.
func (c *Client) Do(ctx context.Context, method, path string, body any) (int, []byte, error) {
	return c.DoAs(ctx, Actor{}, method, path, body)
}

func (c *Client) DoAs(ctx context.Context, who Actor, method, path string, body any) (int, []byte, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		reader = bytes.NewReader(data)
	}
	url := strings.TrimRight(c.cfg.URL, "/") + "/silicon/v1" + path
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Accept", "application/json")
	if who.User != "" {
		req.Header.Set("X-Silicon-User", who.User)
		req.Header.Set("X-Silicon-Role", who.Role)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	return resp.StatusCode, data, err
}

// GetJSON decodes a 2xx answer into out.
func (c *Client) GetJSON(ctx context.Context, path string, out any) error {
	status, data, err := c.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	if status < 200 || status > 299 {
		return fmt.Errorf("%s", Explain(status, data))
	}
	return json.Unmarshal(data, out)
}

// Explain turns a failed backend answer into something an admin can act on.
func Explain(status int, data []byte) string {
	var answer struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(data, &answer)
	switch {
	case status == http.StatusUnauthorized:
		return "the backend rejected Silicon's API key: backend.apiKey in silicon.json must equal SILICON_API_KEY in the backend's .env"
	case status == http.StatusNotFound && strings.Contains(answer.Error, "SILICON_API_KEY"):
		return "the backend's Silicon API is off: set SILICON_API_KEY in the backend's .env to backend.apiKey from silicon.json, then restart the backend"
	case answer.Error != "":
		return fmt.Sprintf("backend answered %d: %s", status, answer.Error)
	case status == http.StatusNotFound:
		return "the backend has no Silicon API at this address (check backend.url in silicon.json)"
	default:
		return fmt.Sprintf("backend answered %d", status)
	}
}
