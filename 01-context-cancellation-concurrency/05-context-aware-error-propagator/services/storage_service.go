package services

import (
	"context"
	"fmt"
	"net"
	"time"
)

type StorageQuotaError struct {
	Err       error
	IsTemp    bool
	IsTimeOut bool
}

func (e *StorageQuotaError) Unwrap() error {
	return e.Err
}

// Temporary implements [net.Error].
func (e *StorageQuotaError) Temporary() bool {
	panic("unimplemented")
}

// Timeout implements [net.Error].
func (e *StorageQuotaError) Timeout() bool {
	panic("unimplemented")
}

func (e *StorageQuotaError) Error() string {
	return e.Err.Error()
}

var _ net.Error = (*StorageQuotaError)(nil)

type StorageService struct {
	m    *MetaDataService
	auth *AuthService
}

func NewStorageService(m *MetaDataService, auth *AuthService) *StorageService {
	return &StorageService{
		m:    m,
		auth: auth,
	}
}

func (s *StorageService) Store(ctx context.Context) error {
	m, err := s.m.GetMetadata(ctx)
	if err != nil {
		return fmt.Errorf("getting metadata is failed: %w", err)
	}
	_ = m
	if err := s.blob(ctx); err != nil {
		return fmt.Errorf("storage error: %w", err)
	}
	return nil
}

func (s *StorageService) blob(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		return &StorageQuotaError{
			Err:       ctx.Err(),
			IsTimeOut: true,
			IsTemp:    false,
		}
	case <-time.After(3 * time.Second):
		return nil
	}
}
