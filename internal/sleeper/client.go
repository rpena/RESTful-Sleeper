package sleeper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Response struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) Get(ctx context.Context, path string, query string) (Response, error) {
	url := c.baseURL + "/" + strings.TrimLeft(path, "/")
	if query != "" {
		url += "?" + query
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Response{}, fmt.Errorf("create Sleeper request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("request Sleeper: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return Response{}, fmt.Errorf("read Sleeper response: %w", err)
	}
	return Response{StatusCode: resp.StatusCode, ContentType: resp.Header.Get("Content-Type"), Body: body}, nil
}
