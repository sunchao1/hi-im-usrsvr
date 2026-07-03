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

package hub

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/sunchao1/hi-im-hubclient/pkg/hubclient"
)

// CMD constants not yet exported from hi-im-api (M4 lifecycle set).
const (
	CMDOffline = 0x0103
)

// Client wraps hubclient lifecycle and handler registration.
type Client struct {
	inner *hubclient.Client
	log   *slog.Logger
}

// NewClient builds a hub BACKEND client from environment and cfg fields.
func NewClient(cfg hubclient.Config) (*Client, error) {
	c, err := hubclient.New(&cfg)
	if err != nil {
		return nil, err
	}
	return &Client{inner: c, log: slog.Default()}, nil
}

// RegisterHandler registers a cmd handler before Start.
func (c *Client) RegisterHandler(cmd uint32, h hubclient.Handler) error {
	return c.inner.RegisterHandler(cmd, h)
}

// Start launches the hub client goroutines.
func (c *Client) Start(ctx context.Context) error {
	return c.inner.Start(ctx)
}

// WaitReady blocks until AUTH+SUB complete.
func (c *Client) WaitReady(ctx context.Context) error {
	return c.inner.WaitReady(ctx)
}

// Ready reports handshake completion.
func (c *Client) Ready() bool {
	return c.inner.Ready()
}

// AsyncSend enqueues a business frame to destNid.
func (c *Client) AsyncSend(cmd, destNid uint32, payload []byte) error {
	return c.inner.AsyncSend(cmd, destNid, payload)
}

// Close stops the hub client.
func (c *Client) Close() error {
	return c.inner.Close()
}

// ConfigFromEnv builds hubclient.Config from explicit app settings.
func ConfigFromEnv(nid uint32, subCmdsCSV, authUser, authPass, backendAddr string) (*hubclient.Config, error) {
	cfg := hubclient.DefaultConfig()
	cfg.Addr = backendAddr
	cfg.NID = nid
	cfg.User = authUser
	cfg.Pass = authPass
	cmds, err := parseHexList(subCmdsCSV)
	if err != nil {
		return nil, fmt.Errorf("HIIM_SUB_CMDS: %w", err)
	}
	cfg.Subscribe = cmds
	return cfg, nil
}

func parseHexList(s string) ([]uint32, error) {
	parts := strings.Split(s, ",")
	out := make([]uint32, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.ParseUint(p, 0, 32)
		if err != nil {
			return nil, err
		}
		out = append(out, uint32(v))
	}
	return out, nil
}
