package services

import (
	"context"
	"fmt"
)

type MetaDataService struct {
	Auth *AuthService
}

func (m *MetaDataService) GetMetadata(ctx context.Context) (map[string]any, error) {
	if err := m.Auth.Login(ctx, "abc"); err != nil {
		return nil, fmt.Errorf("auth failed: %w", err)
	}
	return nil, nil
}
