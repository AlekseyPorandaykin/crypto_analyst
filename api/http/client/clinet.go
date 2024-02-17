package client

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"time"
)

type SymbolPrice struct {
	Exchange string    `json:"exchange"`
	Symbol   string    `json:"symbol"`
	Price    float64   `json:"price"`
	Date     time.Time `json:"date"`
}

type Client struct {
	client  *http.Client
	hostUrl *url.URL
}

func NewClient(client *http.Client, host string) (*Client, error) {
	hostUrl, err := url.Parse(host)
	if err != nil {
		return nil, errors.Wrap(err, "parse host")
	}
	return &Client{
		hostUrl: hostUrl,
		client:  client,
	}, nil
}

func DefaultClient() (*Client, error) {
	c, err := NewClient(http.DefaultClient, "http://localhost:8082/")
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) NewSymbol(ctx context.Context) ([]SymbolPrice, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.hostUrl.JoinPath("/api/price/new").String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "create request")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "execute request")
	}
	defer func() { _ = resp.Body.Close() }()
	var result []SymbolPrice
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.Wrap(err, "decode response")
	}
	return result, nil
}
