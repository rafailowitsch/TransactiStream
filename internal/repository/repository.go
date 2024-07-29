package repository

import (
	"TransactiStream/internal/domain"
	"context"
)

type Database interface {
	Create(ctx context.Context, trans *domain.Transaction) (string, error)
	Read(ctx context.Context, id string) (*domain.Transaction, error)
	Update(ctx context.Context, trans *domain.Transaction) error
	ReadAll(ctx context.Context) ([]*domain.Transaction, error)
}

type Repository struct {
	DB Database
}

func NewRepository(db Database) *Repository {
	return &Repository{
		DB: db,
	}
}
