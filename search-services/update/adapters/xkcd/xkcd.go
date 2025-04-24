package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"yadro.com/course/update/core"
)

type Client struct {
	log        *slog.Logger
	client     http.Client
	url        string
	missingIDs []int
	mu         sync.Mutex
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

func (c *Client) Get(ctx context.Context, id int) (core.XKCDInfo, error) {
	url := fmt.Sprintf("%s/%d/info.0.json", c.url, id)
	resp, err := c.client.Get(url)
	if err != nil {
		return core.XKCDInfo{}, fmt.Errorf("failed to get comic %d: %w", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.mu.Lock()
		c.missingIDs = append(c.missingIDs, id)
		c.mu.Unlock()
		return core.XKCDInfo{}, core.ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return core.XKCDInfo{}, fmt.Errorf("failed to get comic %d: status %d", id, resp.StatusCode)
	}
	info := struct {
		ID         int    `json:"num"`
		URL        string `json:"img"`
		Title      string `json:"title"`
		Transcript string `json:"transcript"`
		Alt        string `json:"alt"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return core.XKCDInfo{}, fmt.Errorf("failed to decode comics: %v", err)
	}

	return core.XKCDInfo{
		NUM:         info.ID,
		URL:         info.URL,
		Title:       info.Title,
		Description: info.Transcript + info.Alt + info.Title,
	}, nil
}

func (c *Client) LastID(ctx context.Context) (int, error) {
	resp, err := c.client.Get(c.url + "/info.0.json")
	if err != nil {
		return 0, fmt.Errorf("failed to get last comic: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get last comic: status %d", resp.StatusCode)
	}

	var info core.XKCDInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return 0, fmt.Errorf("failed to decode last comic: %w", err)
	}

	return info.NUM, nil
}

func (c *Client) MissingIds(ctx context.Context) []int {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.missingIDs == nil {
		return []int{}
	}

	missingIDsCopy := make([]int, len(c.missingIDs))
	copy(missingIDsCopy, c.missingIDs)

	return missingIDsCopy
}
