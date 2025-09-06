package handler

import (
	"github.com/tokane888/go-repository-template/services/api/internal/usecase"
	"go.uber.org/zap"
)

type Handler struct {
	logger      *zap.Logger
	userUseCase usecase.UserUseCase
}

func NewHandler(logger *zap.Logger, userUseCase usecase.UserUseCase) *Handler {
	return &Handler{
		logger:      logger,
		userUseCase: userUseCase,
	}
}
