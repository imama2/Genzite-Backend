package repository

import "context"

type Repository interface {
	Ping(ctx context.Context) error
}

type InMemoryRepository struct{}

func New() *InMemoryRepository {
	return &InMemoryRepository{}
}

func (r *InMemoryRepository) Ping(_ context.Context) error {
	return nil
}
