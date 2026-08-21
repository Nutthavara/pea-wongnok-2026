package recipe

import "context"

type Repository interface {
	Create(ctx context.Context, recipe Recipe) (*Recipe, error)
}

type service struct {
	repository Repository
}

func NewService(repo Repository) *service {
	return &service{
		repository: repo,
	}
}
