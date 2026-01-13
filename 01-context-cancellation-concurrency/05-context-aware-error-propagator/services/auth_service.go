package services

import (
	"context"
	"fmt"
	"slices"
)

type AuthError struct {
	Err    error
	ApiKey string
}

func (a *AuthError) Error() string {
	return a.Err.Error()
}

type AuthService struct {
	ApiKeys []string
}

func (a *AuthService) Login(ctx context.Context, apiKey string) error {
	if !slices.Contains(a.ApiKeys, apiKey) {
		return &AuthError{
			ApiKey: apiKey,
			Err:    fmt.Errorf("api key does not exists"),
		}
	}
	return nil
}
