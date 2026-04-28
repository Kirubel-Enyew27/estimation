package store

import (
	"context"
	"errors"

	"estimation/domain"
)

// ErrNotFound is returned when an entity is not present in the store.
var ErrNotFound = errors.New("not found")

// MaterialStore defines methods to access the material catalog.
// Implementations may use in-memory data, files, or external databases.
type MaterialStore interface {
    GetByCode(ctx context.Context, code string) (*domain.Material, error)
    List(ctx context.Context) ([]*domain.Material, error)
}
