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

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sunchao1/hi-im-api/pkg/im/cmd"

	"github.com/sunchao1/hi-im-usrsvr/internal/config"
	"github.com/sunchao1/hi-im-usrsvr/internal/health"
	"github.com/sunchao1/hi-im-usrsvr/internal/hub"
	imhttp "github.com/sunchao1/hi-im-usrsvr/internal/http"
	"github.com/sunchao1/hi-im-usrsvr/internal/seqsvr"
	"github.com/sunchao1/hi-im-usrsvr/internal/session"
)

func main() {
	cfg, err := config.ConfigFromEnv()
	if err != nil {
		slog.Error("load config failed", "err", err)
		os.Exit(1)
	}

	log := newLogger(cfg.LogLevel)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Error("redis ping failed", "addr", cfg.RedisAddr, "err", err)
		os.Exit(1)
	}
	defer func() { _ = rdb.Close() }()

	seqClient, err := seqsvr.Dial(ctx, cfg.SeqsvrAddr)
	if err != nil {
		log.Error("seqsvr dial failed", "addr", cfg.SeqsvrAddr, "err", err)
		os.Exit(1)
	}
	defer func() { _ = seqClient.Close() }()

	sessStore := session.NewStore(rdb, cfg.ChatSidTTL)

	hubCfg, err := hub.ConfigFromEnv(cfg.NID, cfg.SubCmds, cfg.AuthUser, cfg.AuthPass, cfg.BackendAddr)
	if err != nil {
		log.Error("hub config failed", "err", err)
		os.Exit(1)
	}
	hubClient, err := hub.NewClient(*hubCfg)
	if err != nil {
		log.Error("hub client init failed", "err", err)
		os.Exit(1)
	}

	onlineH := hub.NewOnlineHandler(sessStore, seqClient, hubClient, cfg.M4SkipToken)
	offlineH := hub.NewOfflineHandler(sessStore)
	pingH := hub.NewPingHandler(sessStore, hubClient)

	_ = hubClient.RegisterHandler(cmd.CMD_ONLINE, onlineH.Handle)
	_ = hubClient.RegisterHandler(hub.CMDOffline, offlineH.Handle)
	_ = hubClient.RegisterHandler(cmd.CMD_PING, pingH.Handle)

	if err := hubClient.Start(ctx); err != nil {
		log.Error("hub start failed", "err", err)
		os.Exit(1)
	}
	defer func() { _ = hubClient.Close() }()

	readyCtx, readyCancel := context.WithTimeout(ctx, 30*time.Second)
	defer readyCancel()
	if err := hubClient.WaitReady(readyCtx); err != nil {
		log.Error("hub wait ready failed", "err", err)
		os.Exit(1)
	}
	log.Info("hub BACKEND ready", "nid", cfg.NID)

	ginEngine := imhttp.NewRouter(imhttp.RegisterDeps{Seq: seqClient})
	healthSrv := health.New(rdb, hubClient)
	ginEngine.GET("/healthz", func(c *gin.Context) {
		healthSrv.ServeHealthz(c.Writer, c.Request)
	})
	ginEngine.GET("/readyz", func(c *gin.Context) {
		healthSrv.ServeReadyz(c.Writer, c.Request)
	})

	httpSrv := &http.Server{
		Addr:    cfg.HTTPListen,
		Handler: ginEngine,
	}

	go func() {
		log.Info("http server started", "addr", cfg.HTTPListen)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http serve failed", "err", err)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}
