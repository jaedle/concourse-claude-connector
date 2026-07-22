package concourse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// The fly client credentials are a Concourse-wide constant used for the
// resource-owner password flow of local users.
const (
	flyClientID     = "fly"
	flyClientSecret = "Zmx5"
)

type Pipeline struct {
	Name     string `json:"name"`
	Team     string `json:"team_name"`
	Paused   bool   `json:"paused"`
	Public   bool   `json:"public"`
	Archived bool   `json:"archived"`
}

type Config struct {
	URL      string
	Username string
	Password string
}

type Client struct {
	config     Config
	httpClient *http.Client

	mutex sync.Mutex
	token string
}

func NewClient(config Config) *Client {
	return &Client{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Pipelines(ctx context.Context) ([]Pipeline, error) {
	body, err := c.get(ctx, "/api/v1/pipelines")
	if err != nil {
		return nil, err
	}

	var pipelines []Pipeline
	if err := json.Unmarshal(body, &pipelines); err != nil {
		return nil, fmt.Errorf("parse pipelines: %w", err)
	}
	return pipelines, nil
}

func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	token, err := c.currentToken(ctx)
	if err != nil {
		return nil, err
	}

	body, status, err := c.doGet(ctx, path, token)
	if err != nil {
		return nil, err
	}
	if status == http.StatusUnauthorized {
		token, err = c.login(ctx)
		if err != nil {
			return nil, err
		}
		body, status, err = c.doGet(ctx, path, token)
		if err != nil {
			return nil, err
		}
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("concourse responded with status %d on %s", status, path)
	}
	return body, nil
}

func (c *Client) doGet(ctx context.Context, path string, token string) ([]byte, int, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.URL+path, nil)
	if err != nil {
		return nil, 0, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, 0, fmt.Errorf("request %s: %w", path, err)
	}
	defer func() { _ = response.Body.Close() }()

	body, err := readAll(response)
	if err != nil {
		return nil, 0, err
	}
	return body, response.StatusCode, nil
}

func (c *Client) currentToken(ctx context.Context) (string, error) {
	c.mutex.Lock()
	token := c.token
	c.mutex.Unlock()

	if token != "" {
		return token, nil
	}
	return c.login(ctx)
}

func (c *Client) login(ctx context.Context) (string, error) {
	form := url.Values{
		"grant_type": {"password"},
		"username":   {c.config.Username},
		"password":   {c.config.Password},
		"scope":      {"openid profile email federated:id groups"},
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.config.URL+"/sky/issuer/token",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", err
	}
	request.SetBasicAuth(flyClientID, flyClientSecret)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed with status %d", response.StatusCode)
	}

	body, err := readAll(response)
	if err != nil {
		return "", err
	}

	var token struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &token); err != nil {
		return "", fmt.Errorf("parse token: %w", err)
	}
	if token.AccessToken == "" {
		return "", fmt.Errorf("login response contains no access token")
	}

	c.mutex.Lock()
	c.token = token.AccessToken
	c.mutex.Unlock()

	return token.AccessToken, nil
}
