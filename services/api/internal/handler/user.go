package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tokane888/go-repository-template/services/api/internal/domain"
	"github.com/tokane888/go-repository-template/services/api/internal/dto/query"
	"github.com/tokane888/go-repository-template/services/api/internal/dto/request"
	"github.com/tokane888/go-repository-template/services/api/internal/dto/response"
	"github.com/tokane888/go-repository-template/services/api/internal/usecase"
	"go.uber.org/zap"
)

func (h *Handler) CreateUser(c *gin.Context) {
	var req request.CreateUser
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewError("INVALID_REQUEST", "リクエストが不正です"))
		return
	}

	user, err := h.userUseCase.CreateUser(c.Request.Context(), &req)
	if err != nil {
		// ドメインバリデーションエラーを400エラーにマッピング
		if errors.Is(err, domain.ErrInvalidEmail) {
			c.JSON(http.StatusBadRequest, response.NewError("INVALID_EMAIL", "メールアドレスの形式が不正です"))
			return
		}
		if errors.Is(err, domain.ErrPasswordTooShort) {
			c.JSON(http.StatusBadRequest, response.NewError("PASSWORD_TOO_SHORT", "パスワードは8文字以上である必要があります"))
			return
		}
		if errors.Is(err, domain.ErrUsernameTooShort) {
			c.JSON(http.StatusBadRequest, response.NewError("USERNAME_TOO_SHORT", "ユーザー名は3文字以上である必要があります"))
			return
		}
		if errors.Is(err, domain.ErrUsernameTooLong) {
			c.JSON(http.StatusBadRequest, response.NewError("USERNAME_TOO_LONG", "ユーザー名は100文字以下である必要があります"))
			return
		}
		if errors.Is(err, domain.ErrInvalidPasswordFormat) {
			c.JSON(http.StatusBadRequest, response.NewError("INVALID_PASSWORD_FORMAT", "パスワードは英字と数字の両方を含む必要があります"))
			return
		}
		if errors.Is(err, usecase.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, response.NewError("USER_ALREADY_EXISTS", "ユーザーは既に存在します"))
			return
		}
		h.logger.Error("failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.NewError("INTERNAL_ERROR", "内部エラーが発生しました"))
		return
	}

	c.JSON(http.StatusCreated, response.NewUserFromDomain(user))
}

func (h *Handler) ListUsers(c *gin.Context) {
	var q query.ListUsers
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, response.NewError("INVALID_REQUEST", "リクエストが不正です"))
		return
	}

	users, total, err := h.userUseCase.ListUsers(c.Request.Context(), q.Limit, q.Offset)
	if err != nil {
		h.logger.Error("failed to list users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.NewError("INTERNAL_ERROR", "内部エラーが発生しました"))
		return
	}

	c.JSON(http.StatusOK, response.NewUserListFromDomain(users, total))
}

func (h *Handler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewError("INVALID_ID", "IDの形式が不正です"))
		return
	}

	err = h.userUseCase.DeleteUser(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, response.NewError("USER_NOT_FOUND", "ユーザーが見つかりません"))
			return
		}
		h.logger.Error("failed to delete user", zap.Error(err), zap.String("user_id", id.String()))
		c.JSON(http.StatusInternalServerError, response.NewError("INTERNAL_ERROR", "内部エラーが発生しました"))
		return
	}

	c.Status(http.StatusNoContent)
}
