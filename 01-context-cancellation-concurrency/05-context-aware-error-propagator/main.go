package main

import (
	"context"
	"context-aware-errors/services"
	"errors"
	"fmt"
	"log/slog"
)

func main() {
	logger := slog.Default()

	ctx := context.Background()
	authSvc := services.AuthService{
		ApiKeys: []string{"abc", "def"},
	}
	metSvc := services.MetaDataService{
		Auth: &authSvc,
	}

	storageSvc := services.NewStorageService(&metSvc, &authSvc)

	err := storageSvc.Store(ctx)
	if err != nil {
		newErr := fmt.Errorf("store failed: %w", err)
		logger.Error(fmt.Sprint(newErr))

		// var authErr *services.AuthError
		// if !errors.As(newErr, &authErr) {
		// 	logger.Error("auth error failed to parse", slog.Any("auth_err", authErr))
		// }

		if !errors.Is(err, context.DeadlineExceeded) {
			logger.Error("timeout condition failed")
		}

		logger.Error("timeout", slog.Any("_err", err))
	}
}
