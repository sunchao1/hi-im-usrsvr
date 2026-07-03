// Copyright 2026 sunchao1
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package health

import (
	"context"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// HubReady reports whether the hub BACKEND client has completed handshake.
type HubReady interface {
	Ready() bool
}

// Server exposes liveness and readiness HTTP endpoints.
type Server struct {
	redis *redis.Client
	hub   HubReady
}

// New returns a health check server.
func New(rdb *redis.Client, hub HubReady) *Server {
	return &Server{redis: rdb, hub: hub}
}

// Handler returns an http.Handler with /healthz and /readyz routes.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/readyz", s.handleReadyz)
	return mux
}

// ServeHealthz writes the liveness response.
func (s *Server) ServeHealthz(w http.ResponseWriter, r *http.Request) {
	s.handleHealthz(w, r)
}

// ServeReadyz writes the readiness response.
func (s *Server) ServeReadyz(w http.ResponseWriter, r *http.Request) {
	s.handleReadyz(w, r)
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleReadyz(w http.ResponseWriter, _ *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if s.redis != nil {
		if err := s.redis.Ping(ctx).Err(); err != nil {
			http.Error(w, "redis unavailable", http.StatusServiceUnavailable)
			return
		}
	}
	if s.hub != nil && !s.hub.Ready() {
		http.Error(w, "hub not ready", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
