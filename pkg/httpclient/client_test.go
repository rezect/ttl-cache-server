package httpclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/rezect/go-interview/pkg/httpclient"
)

func returnOK(w http.ResponseWriter, t *testing.T) {
	w.WriteHeader(http.StatusOK)
	body := map[string]any{
		"login":        "torvalds",
		"name":         "Linus Torvalds",
		"public_repos": 12,
		"followers":    311034,
		"created_at":   "2011-09-03T15:26:22Z",
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		t.Error(err)
	}
	w.Write(bodyJSON)
}

func TestSuccess(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		returnOK(w, t)
	}))
	defer mockServer.Close()

	cfg := httpclient.DefaultConfig()
	cfg.BaseURL = mockServer.URL
	client := httpclient.NewClient(cfg)

	user, err := client.GetUser(context.Background(), "torvalds")
	if err != nil {
		t.Error(err)
	}

	cfg.Logger.Print(user)
}

func TestNotFound(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		body := map[string]any{
			"error": "user not found",
		}
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			t.Error(err)
		}
		w.Write(bodyJSON)
	}))
	defer mockServer.Close()

	cfg := httpclient.DefaultConfig()
	cfg.BaseURL = mockServer.URL
	client := httpclient.NewClient(cfg)

	user, err := client.GetUser(context.Background(), "torvalds")
	if err == nil {
		t.Error("Метод вернул без ошибки, а должен был http.StatusNotFound")
	}

	if user.Login != "" {
		t.Errorf("Должен был вернуть пустой User. А получил %v", user.Login)
	}
}

func TestTooManyRequests_RateLimitReset(t *testing.T) {
	serverValue := 0
	delayTimeout := 50 * time.Millisecond
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if serverValue == 0 {
			resetLimitsTime := time.Now().Add(delayTimeout)
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetLimitsTime.Unix(), 10))
			w.WriteHeader(http.StatusTooManyRequests)
			body := map[string]any{
				"error": "too many requests",
			}
			bodyJSON, err := json.Marshal(body)
			if err != nil {
				t.Error(err)
			}
			w.Write(bodyJSON)

			serverValue++
		} else {
			returnOK(w, t)
		}
	}))
	defer mockServer.Close()

	cfg := httpclient.DefaultConfig()
	cfg.BaseURL = mockServer.URL
	client := httpclient.NewClient(cfg)

	ch := make(chan httpclient.User)
	defer close(ch)
	go func(ch chan httpclient.User) {
		user, err := client.GetUser(context.Background(), "torvalds")
		if err != nil {
			t.Error(err)
		}
		ch <- user
	}(ch)

	select {
	case <-time.After(delayTimeout * 2):
		t.Errorf("Не пришел ответ от функции через нужное время")
		break
	case user := <-ch:
		if user.Login != "torvalds" {
			t.Errorf("Wrong login. Expected [%v], got [%v]", "torvalds", user.Login)
		}
		break
	}
}

func TestTooManyRequests_WithoutHeaders(t *testing.T) {
	serverValue := 0
	timeout := 50 * time.Millisecond
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if serverValue == 0 {
			w.WriteHeader(http.StatusTooManyRequests)
			serverValue++
		} else {
			returnOK(w, t)
		}
	}))
	defer mockServer.Close()

	cfg := httpclient.DefaultConfig()
	cfg.BaseURL = mockServer.URL
	cfg.MinDelay = timeout
	client := httpclient.NewClient(cfg)

	ch := make(chan httpclient.User)
	defer close(ch)
	go func(ch chan httpclient.User) {
		user, err := client.GetUser(context.Background(), "torvalds")
		if err != nil {
			t.Error(err)
		}

		ch <- user
	}(ch)

	select {
	case <-time.After(timeout * 2):
		t.Errorf("Не пришел ответ от функции через нужное время")
	case user := <-ch:
		if user.Login != "torvalds" {
			t.Errorf("Wrong login. Expected [%v], got [%v]", "torvalds", user.Login)
		}
	}
}

func TestInternalServerError(t *testing.T) {
	serverValue := 0
	timeout := 50 * time.Millisecond
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if serverValue == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			serverValue++
		} else {
			returnOK(w, t)
		}
	}))
	defer mockServer.Close()

	cfg := httpclient.DefaultConfig()
	cfg.BaseURL = mockServer.URL
	cfg.MinDelay = timeout
	client := httpclient.NewClient(cfg)

	ch := make(chan httpclient.User)
	defer close(ch)
	go func(ch chan httpclient.User) {
		user, err := client.GetUser(context.Background(), "torvalds")
		if err != nil {
			t.Error(err)
		}
		ch <- user
	}(ch)

	select {
	case <-time.After(timeout * 2):
		t.Errorf("Не пришел ответ от функции через нужное время")
	case user := <-ch:
		if user.Login != "torvalds" {
			t.Errorf("Wrong login. Expected [%v], got [%v]", "torvalds", user.Login)
		}
	}
}
