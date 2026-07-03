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

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds usrsvr runtime settings loaded from environment variables.
type Config struct {
	HTTPListen   string
	BackendAddr  string
	NID          uint32
	AuthUser     string
	AuthPass     string
	SubCmds      string
	RedisAddr    string
	SeqsvrAddr   string
	LogLevel     string
	M4SkipToken  bool
	ChatSidTTL   int
}

// ConfigFromEnv loads configuration from HIIM_* environment variables.
func ConfigFromEnv() (Config, error) {
	cfg := Config{
		HTTPListen:  envOr("HIIM_HTTP_LISTEN", ":8081"),
		BackendAddr: os.Getenv("HIIM_BACKEND_ADDR"),
		AuthUser:    envOr("HIIM_AUTH_USER", "proxy"),
		AuthPass:    envOr("HIIM_AUTH_PASS", "proxy"),
		SubCmds:     envOr("HIIM_SUB_CMDS", "0x0101,0x0103,0x0105,0x0301,0x0305"),
		RedisAddr:   envOr("HIIM_REDIS_ADDR", "127.0.0.1:6379"),
		SeqsvrAddr:  envOr("HIIM_SEQSVR_ADDR", "127.0.0.1:50051"),
		LogLevel:    envOr("HIIM_LOG_LEVEL", "info"),
		M4SkipToken: envBoolOr("HIIM_M4_SKIP_TOKEN", true),
		ChatSidTTL:  envIntOr("HIIM_CHAT_SID_TTL", 300),
	}

	if v := os.Getenv("HIIM_NID"); v != "" {
		n, err := parseU32(v)
		if err != nil {
			return cfg, fmt.Errorf("HIIM_NID: %w", err)
		}
		cfg.NID = n
	} else {
		cfg.NID = 30001
	}

	if cfg.BackendAddr == "" {
		return cfg, fmt.Errorf("HIIM_BACKEND_ADDR is required")
	}
	if cfg.RedisAddr == "" {
		return cfg, fmt.Errorf("HIIM_REDIS_ADDR is required")
	}
	if cfg.SeqsvrAddr == "" {
		return cfg, fmt.Errorf("HIIM_SEQSVR_ADDR is required")
	}
	if cfg.ChatSidTTL <= 0 {
		return cfg, fmt.Errorf("HIIM_CHAT_SID_TTL must be positive")
	}
	return cfg, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envIntOr(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func envBoolOr(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

func parseU32(s string) (uint32, error) {
	v, err := strconv.ParseUint(strings.TrimSpace(s), 0, 32)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}
