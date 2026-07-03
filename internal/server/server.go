package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rezect/go-interview/internal/cache"
	"github.com/rezect/go-interview/internal/middleware"
	"github.com/rezect/go-interview/internal/response"
)

type Server struct {
	cache  *cache.Cache
	httpServer *http.Server
}

type postRequest struct {
	Value any    `json:"value"`
	Ttl   string `json:"ttl"`
}

func NewServer(addr string) *Server {
	srv := &Server{}

	srv.cache = cache.CacheNew()
	srv.httpServer = &http.Server{
		Addr:         addr,
		Handler:      srv.Handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return srv
}

func (s *Server) Handler() http.Handler {
	mux := s.setupRoutes()
	recoverMiddleware := middleware.RecoverMiddleware(mux)
	logMiddleware := middleware.LoggingMiddleware(recoverMiddleware)
	return logMiddleware
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	err := s.httpServer.Shutdown(ctx)
	s.cache.Stop()

	return err
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]string{"health": "Ok"})
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	value, err := s.cache.Get(key)
	if err != nil {
		response.WriteJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"value": value})
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	var req postRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	defer r.Body.Close()

	ttl, err := time.ParseDuration(req.Ttl)
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "Invalid TTL format. Valid formats for TTL is: 'ns', 'us' (or 'µs'), 'ms', 's', 'm', 'h'"})
		return
	}
	err = s.cache.Set(key, req.Value, ttl)
	if err != nil {
		response.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	err := s.cache.Delete(key)
	if err != nil {
		response.WriteJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleClear(w http.ResponseWriter, r *http.Request) {
	s.cache.Clear()
	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.healthHandler)
	mux.HandleFunc("GET /cache/{key}", s.handleGet)
	mux.HandleFunc("POST /cache/{key}", s.handlePost)
	mux.HandleFunc("DELETE /cache/{key}", s.handleDelete)
	mux.HandleFunc("POST /cache/clear", s.handleClear)

	return mux
}
