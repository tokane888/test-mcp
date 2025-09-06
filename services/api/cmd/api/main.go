package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	// TODO: import元調整
	pkglogger "github.com/tokane888/go-repository-template/pkg/logger"
	"github.com/tokane888/go-repository-template/services/api/internal/config"
	"github.com/tokane888/go-repository-template/services/api/internal/db"
	"github.com/tokane888/go-repository-template/services/api/internal/handler"
	"github.com/tokane888/go-repository-template/services/api/internal/infrastructure/persistence"
	"github.com/tokane888/go-repository-template/services/api/internal/router"
	"github.com/tokane888/go-repository-template/services/api/internal/usecase"
	"go.uber.org/zap"
)

// アプリのversion。デフォルトは開発版。cloud上ではbuild時に-ldflagsフラグ経由でバージョンを埋め込む
var version = "dev"

func main() {
	cfg, err := config.LoadConfig(version)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logger := pkglogger.NewLogger(cfg.Logger)
	//nolint: errcheck
	defer logger.Sync()

	// データベース接続
	database, err := db.Connect(&cfg.DatabaseConfig)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer func() {
		if closeErr := database.Close(); closeErr != nil {
			logger.Error("failed to close database connection", zap.Error(closeErr))
		}
	}()

	// Repository層の初期化
	userRepository := persistence.NewUserRepository(database, logger)
	// UseCase層の初期化
	userUseCase := usecase.NewUserUseCase(userRepository, logger)
	// Handler層の初期化
	h := handler.NewHandler(logger, userUseCase)
	r := router.NewRouter(&cfg.RouterConfig, logger, h)
	engine := r.Setup()

	// シグナルハンドリングの設定
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.RouterConfig.Port),
		Handler: engine,
	}

	// サーバーをgoroutineで起動
	go func() {
		logger.Info("starting API server", zap.Int("port", cfg.RouterConfig.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen error", zap.Error(err))
		}
	}()

	// シグナル待機
	<-ctx.Done()
	logger.Info("shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownTimeout)*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited")
}
