package store

import (
	"context"
	"errors"
	"estimation/domain"
)


var ErrNotFound = errors.New("not found")

type MaterialStore interface {
	GetByCode(ctx context.Context, code string) (*domain.Material, error) 
	List(ctx context.Context) ([]*domain.Material, error)
}

