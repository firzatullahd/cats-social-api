package handler

import (
	"context"

	"github.com/firzatullahd/cats-social-api/internal/model"
)

type IUsecase interface {
	Register(ctx context.Context, in *model.RegisterRequest) (*model.AuthResponse, error)
	Login(ctx context.Context, in *model.LoginRequest) (*model.AuthResponse, error)
}

type Handler struct {
	Usecase IUsecase
}

func NewHandler(usecase IUsecase) *Handler {
	return &Handler{
		Usecase: usecase,
	}
}
