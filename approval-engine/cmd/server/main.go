// Command server 是审批流引擎的入口：加载配置、初始化数据库、（可选）写入
// 演示数据、组装路由、启动 HTTP 服务，并支持优雅关闭。
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/yorkyu/approval-engine/internal/auth"
	"github.com/yorkyu/approval-engine/internal/config"
	"github.com/yorkyu/approval-engine/internal/httpserver"
	"github.com/yorkyu/approval-engine/internal/seed"
	"github.com/yorkyu/approval-engine/internal/store"
)

func main() {
	cfg := config.Load()

	if dir := filepath.Dir(cfg.DBPath); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatalf("create db directory: %v", err)
		}
	}

	db, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if cfg.SeedOnBoot {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := seed.Run(ctx, db); err != nil {
			cancel()
			log.Fatalf("seed database: %v", err)
		}
		cancel()
	}

	signer := auth.NewSigner(cfg.JWTSecret)
	router := httpserver.NewRouter(db, signer)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("approval-engine listening on %s (db: %s)", cfg.Addr, cfg.DBPath)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	// 优雅关闭：收到 Ctrl+C / SIGTERM 后，给正在处理的请求最多 5 秒完成，
	// 而不是直接把连接砍断——这是任何生产服务都应该有的基本素养。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Println("shutting down ...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("bye.")
}
