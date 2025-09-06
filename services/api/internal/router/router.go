package router

import (
	"github.com/gin-gonic/gin"
	"github.com/tokane888/go-repository-template/services/api/internal/handler"
	"github.com/tokane888/go-repository-template/services/api/internal/router/middleware"
	"go.uber.org/zap"
)

type Config struct {
	Port int
}

type Router struct {
	config  *Config
	logger  *zap.Logger
	engine  *gin.Engine
	handler *handler.Handler
}

func NewRouter(config *Config, logger *zap.Logger, handler *handler.Handler) *Router {
	return &Router{
		config:  config,
		logger:  logger,
		engine:  gin.New(),
		handler: handler,
	}
}

func (r *Router) Setup() *gin.Engine {
	// グローバルミドルウェア
	r.engine.Use(gin.Recovery()) // handler内でpanic発生時に500を返す
	r.engine.Use(middleware.Logger(r.logger))

	// ヘルスチェック（認証不要）
	r.engine.GET("/health", r.handler.Health)

	// APIグループ（v1）
	v1 := r.engine.Group("/api/v1")
	{
		// API Key認証ミドルウェアを適用
		v1.Use(middleware.APIKeyAuth())

		// ユーザー管理エンドポイント
		users := v1.Group("/users")
		{
			users.POST("", r.handler.CreateUser)
			users.GET("", r.handler.ListUsers)
			users.DELETE("/:id", r.handler.DeleteUser)
		}
	}

	return r.engine
}
