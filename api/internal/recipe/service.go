package recipe

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Repository interface {
	HasActiveReferences(ctx context.Context, difficultyID, durationID string) (bool, error)
	Create(ctx context.Context, recipe Recipe) (*Recipe, error)
	List(ctx context.Context, query GetRecipesQuery) ([]Recipe, int64, error)
	DifficultyExists(ctx context.Context, id string) (bool, error)
	FindByID(ctx context.Context, id int) (*Recipe, error)
	Replace(ctx context.Context, recipe Recipe) (*Recipe, error)
	Delete(ctx context.Context, id int) error
}

type service struct {
	repository Repository
}

func NewService(repo Repository) *service {
	return &service{
		repository: repo,
	}
}

func (svc *service) Create(ctx context.Context, creatorID uuid.UUID, recipe Recipe) (*Recipe, error) {
	refActive, err := svc.repository.HasActiveReferences(ctx, recipe.DifficultyID, recipe.DurationID)
	if err != nil {
		return nil, err
	}

	if !refActive {
		return nil, ErrInvalidReferenceData
	}

	recipe.CreatorID = creatorID

	return svc.repository.Create(ctx, recipe)
}

func (svc *service) List(ctx context.Context, query GetRecipesQuery) ([]Recipe, int64, error) {
	// Inject default parameter if blank
	query.Ensure()

	if query.Difficulty != "" {
		exists, err := svc.repository.DifficultyExists(ctx, query.Difficulty)
		if err != nil {
			return nil, 0, fmt.Errorf("list recipes: %w", err)

		}

		if !exists {
			return nil, 0, fmt.Errorf("%w: difficulty %q does not exist", ErrInvalidReferenceData, query.Difficulty)

		}
	}

	recipes, total, err := svc.repository.List(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("list recipes: %w", err)
	}

	return recipes, total, nil
}

func (svc *service) Get(ctx context.Context, id int) (*Recipe, error) {
	recipe, err := svc.repository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get recipe: %w", err)

	}

	return recipe, nil
}

func (svc *service) Replace(ctx context.Context, id int, userID uuid.UUID, recipe Recipe) (*Recipe, error) {
	existing, err := svc.repository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("replace recipe: %w", err)
	}

	if existing.CreatorID != userID {
		return nil, ErrForbidden
	}

	refActive, err := svc.repository.HasActiveReferences(ctx, recipe.DifficultyID, recipe.DurationID)
	if err != nil {
		return nil, fmt.Errorf("replace recipe: %w", err)
	}

	if !refActive {
		return nil, ErrReferenceDataUnavailable
	}

	recipe.ID = id
	recipe.CreatorID = existing.CreatorID

	replaced, err := svc.repository.Replace(ctx, recipe)
	if err != nil {
		return nil, fmt.Errorf("replace recipe: %w", err)
	}

	return replaced, nil
}

func (svc *service) Delete(ctx context.Context, id int, userID uuid.UUID) error {
	existing, err := svc.repository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("delete recipe: %w", err)
	}

	if existing.CreatorID != userID {
		return ErrForbidden
	}

	if err := svc.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete recipe: %w", err)
	}

	return nil
}
