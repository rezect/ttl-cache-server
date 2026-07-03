package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rezect/go-interview/internal/server"
)

func TestHandlePost_Success(t *testing.T) {
	srv := server.NewServer(":0")
	t.Cleanup(func() {srv.Shutdown()})

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
