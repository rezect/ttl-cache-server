package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rezect/go-interview/internal/server"
	"github.com/stretchr/testify/assert"
)

func get(srv *server.Server, setKey string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, httptest.NewRequest("GET", fmt.Sprintf("/cache/%v", setKey), nil))

	return w
}

func set(srv *server.Server, setKey string, setValue string, setTtl string) *httptest.ResponseRecorder {
	setBody := strings.NewReader(fmt.Sprintf(`{"value": "%v", "ttl": "%v"}`, setValue, setTtl))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, httptest.NewRequest("POST", fmt.Sprintf("/cache/%v", setKey), setBody))

	return w
}

func deleteKey(srv *server.Server, setKey string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, httptest.NewRequest("DELETE", fmt.Sprintf("/cache/%v", setKey), nil))

	return w
}

func clear(srv *server.Server) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, httptest.NewRequest("POST", "/cache/clear", nil))

	return w
}

func parseJSON(r *bytes.Buffer) (map[string]string, error) {
	var result map[string]string
	if err := json.NewDecoder(r).Decode(&result); err != nil {
		return nil, fmt.Errorf("Can't parse output json. Err: %v", err)
	}

	return result, nil
}

func TestHandlePost_Success(t *testing.T) {
	srv := server.NewServer(":0")
	t.Cleanup(func() { srv.Shutdown() })
	setKey := "one"
	setValue := "1"
	setTtl := "1h"

	var w *httptest.ResponseRecorder
	w = set(srv, setKey, setValue, setTtl)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status [200] 'OK', got: [%v]", w.Code)
	}

	result, err := parseJSON(w.Body)
	if err != nil {
		t.Errorf("%v", err.Error())
	}

	if result["status"] != "ok" {
		t.Errorf("expected status 'ok', got: '%v'", result["status"])
	}
}

func TestHandlePost_WrongTtl(t *testing.T) {
	srv := server.NewServer(":0")
	t.Cleanup(func() { srv.Shutdown() })
	setKey := "one"
	setValue := "1"
	setTtl := "1hh"

	w := set(srv, setKey, setValue, setTtl)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status [400] 'Bad Request', got: [%v]", w.Code)
	}
}

func TestHandlePost_WrongBody(t *testing.T) {
	srv := server.NewServer(":0")
	t.Cleanup(func() { srv.Shutdown() })

	body := strings.NewReader("not a json format")
	req := httptest.NewRequest("POST", "/cache/sixnine", body)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status [400] 'Bad Request', got: [%v]", w.Code)
	}
}

func TestHandleGet(t *testing.T) {
	srv := server.NewServer(":0")
	t.Cleanup(func() { srv.Shutdown() })
	setKey := "one"
	setValue := "1"
	setTtl := "1h"

	w := set(srv, setKey, setValue, setTtl)

	if w.Code != http.StatusOK {
		t.Errorf("Wrong status code on POST. Expected [%v], got [%v]", http.StatusOK, w.Code)
	}
	tests := []struct {
		name           string
		key            string
		expectedValue  any
		expectedStatus int
	}{
		{"Get invalid value", "uno", nil, 404},
		{"Get valid value", setKey, setValue, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := get(srv, tt.key)
			if w.Code != tt.expectedStatus {
				t.Errorf("Wrong status code. Expected [%v], got [%v]", tt.expectedStatus, w.Code)
			}

			result, err := parseJSON(w.Body)
			if err != nil {
				t.Errorf("%v", err.Error())
			}

			if result["value"] != tt.expectedValue && tt.expectedValue != nil {
				t.Errorf("Wrong return value. Expected '%v', got '%v'", tt.expectedValue, result["value"])
			}
		})
	}
}

func TestHandleDelete(t *testing.T) {
	srv := server.NewServer(":0")
	t.Cleanup(func() { srv.Shutdown() })
	setKey := "one"
	setValue := "1"
	setTtl := "1h"

	w := set(srv, setKey, setValue, setTtl)

	w = deleteKey(srv, setKey)
	if w.Code != http.StatusNoContent {
		t.Errorf("Wrong status. Expected [%v], got [%v]", http.StatusNoContent, w.Code)
	}

	w = deleteKey(srv, setKey)
	if w.Code != http.StatusNotFound {
		t.Errorf("Wrong status. Expected [%v], got [%v]", http.StatusNoContent, w.Code)
	}
}

func TestHandleClear(t *testing.T) {
	srv := server.NewServer(":0")
	t.Cleanup(func() { srv.Shutdown() })
	setKey1 := "one"
	setValue1 := "1"
	setTtl1 := "1h"
	setKey2 := "two"
	setValue2 := "2"
	setTtl2 := "2h"

	var w *httptest.ResponseRecorder
	w = set(srv, setKey1, setValue1, setTtl1)
	w = set(srv, setKey2, setValue2, setTtl2)

	w = clear(srv)
	assert.Equal(t, http.StatusOK, w.Code, "Wrong status. Expected [%v], got [%v]", http.StatusOK, w.Code)

	w = get(srv, setKey1)
	assert.Equal(t, http.StatusNotFound, w.Code, "Wrong status. Expected [%v], got [%v]", http.StatusNotFound, w.Code)
	
	w = get(srv, setKey2)
	assert.Equal(t, http.StatusNotFound, w.Code, "Wrong status. Expected [%v], got [%v]", http.StatusNotFound, w.Code)
}
