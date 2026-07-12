package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type User struct {
	Login       string `json:"login"`
	Name        string `json:"name"`
	PublicRepos int    `json:"public_repos"`
	Followers   int    `json:"followers"`
	CreatedAt   string `json:"created_at"`
}

type client struct {
	cfg        Config
	httpClient *http.Client
}

func (c *client) GetUser(ctx context.Context, login string) (User, error) {
	c.cfg.Logger.Printf("Started GetUser() func.")
	req, err := http.NewRequestWithContext(ctx, "GET", c.cfg.BaseURL+"/"+login, nil)
	if err != nil {
		return User{}, err
	}

	retryTimeout := c.cfg.MinDelay

	for tries := 0; tries < c.cfg.MaxRetries; tries++ {
		if tries != 0 {
			retryTimeout = min(retryTimeout*2, c.cfg.MaxDelay)
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.cfg.Logger.Printf("Request error [%v]. Retrying after %v sec...", err, retryTimeout)
			select {
			case <-time.After(retryTimeout):
			case <-ctx.Done():
				return User{}, ctx.Err()
			}
			continue
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return User{}, err
		}

		switch {
		case resp.StatusCode == http.StatusOK:
			var user User
			err := json.Unmarshal(bodyBytes, &user)
			if err != nil {
				return User{}, err
			}
			return user, nil
		case resp.StatusCode == http.StatusTooManyRequests:
			xRatelimitReset := resp.Header.Get("X-RateLimit-Reset")
			retryAfter := resp.Header.Get("Retry-After")
			if xRatelimitReset != "" {
				timeResetInt, err := strconv.ParseInt(xRatelimitReset, 10, 64)
				if err != nil {
					return User{}, err
				}
				timeReset := time.Unix(timeResetInt, 0)
				select {
				case <-time.After(time.Until(timeReset)):
				case <-ctx.Done():
					return User{}, ctx.Err()
				}
			} else if retryAfter != "" {
				retryAfterMS, err := strconv.ParseInt(retryAfter, 10, 64)
				if err != nil {
					return User{}, err
				}
				select {
				case <-time.After(time.Duration(retryAfterMS)):
				case <-ctx.Done():
					return User{}, ctx.Err()
				}
			} else {
				select {
				case <-time.After(retryTimeout):
				case <-ctx.Done():
					return User{}, ctx.Err()
				}
			}
		case resp.StatusCode >= 500:
			select {
			case <-time.After(retryTimeout):
			case <-ctx.Done():
				return User{}, ctx.Err()
			}
		default:
			return User{}, fmt.Errorf("Статус ответа [%v].", resp.StatusCode)
		}
	}

	return User{}, fmt.Errorf("Max retries limit is hit")
}

type Client interface {
	GetUser(ctx context.Context, login string) (User, error)
}

func NewClient(cfg Config) Client {
	validateConfig(&cfg)
	return &client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}
