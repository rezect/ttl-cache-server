package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rezect/go-interview/internal/server"
)

func TestHandlePost_Success(t *testing.T) {
	srv := server.NewServer(":0")
	t.Cleanup(func() { srv.Shutdown() })

	body := strings.NewReader(`{"value": "69", "ttl": "10s"}`)
	req := httptest.NewRequest("POST", "/cache/sixnine", body)
	w := httptest.NewRecorder()

	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status [200] 'OK', got: [%v]", w.Code)
	}

	var result map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("Can't parse output json. Err: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("expected status 'ok', got: '%v'", result["status"])
	}
}

func TestHandleGet(t *testing.T) {
	srv := server.NewServer(":0")
	t.Cleanup(func() { srv.Shutdown() })
	setKey := "one"
	setValue := "1"
	setTtl := "1h"
	setBody := strings.NewReader(fmt.Sprintf(`{"value": "%v", "ttl": "%v"}`, setValue, setTtl))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, httptest.NewRequest("POST", fmt.Sprintf("/cache/%v", setKey), setBody))
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
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/cache/%v", tt.key), nil)

			srv.Handler().ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Wrong status code. Expected [%v], got [%v]", tt.expectedStatus, w.Code)
			}
			var result map[string]string
			if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
				t.Errorf("Can't parse json, error: %v", err)
			}
			if result["value"] != tt.expectedValue && tt.expectedValue != nil {
				t.Errorf("Wrong return value. Expected '%v', got '%v'", tt.expectedValue, result["value"])
			}
		})
	}
}
