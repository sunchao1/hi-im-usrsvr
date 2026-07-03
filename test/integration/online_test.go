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

//go:build integration

package integration_test

import (
	"context"
	"net"
	"os"
	"testing"
	"time"
)

func TestOnlineAgainstHub(t *testing.T) {
	backendAddr := envOr("HIIM_BACKEND_ADDR", "127.0.0.1:28889")
	if err := waitTCP(backendAddr, 3*time.Second); err != nil {
		t.Skipf("hub backend not reachable at %s: %v", backendAddr, err)
	}
	t.Skip("full stack integration: start usrsvr + hub + redis + seqsvr, inject ONLINE via BACKEND")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func waitTCP(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return context.DeadlineExceeded
}
