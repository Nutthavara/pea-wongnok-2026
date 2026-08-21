package recipe

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServiceCreatePersistsRecipeWithActiveReferences(t *testing.T) {
	repo := NewMockRepository(t)
	creatorID := uuid.New()
	input := Recipe{DifficultyID: "easy", DurationID: "10m"}
	persisted := input
	persisted.CreatorID = creatorID
	created := persisted
	created.ID = 42

	repo.EXPECT().HasActiveReferences(mock.Anything, "easy", "10m").Return(true, nil)
	repo.EXPECT().Create(mock.Anything, persisted).Return(&created, nil)

	result, err := NewService(repo).Create(context.Background(), creatorID, input)

	assert.NoError(t, err)
	assert.Equal(t, &created, result)
}

func TestServiceCreateRejectsInactiveReferencesWithoutPersisting(t *testing.T) {
	repo := NewMockRepository(t)
	repo.EXPECT().HasActiveReferences(mock.Anything, "missing", "10m").Return(false, nil)

	result, err := NewService(repo).Create(context.Background(), uuid.New(), Recipe{
		DifficultyID: "missing",
		DurationID:   "10m",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidReferenceData)
}

func TestServiceCreateReturnsReferenceValidationFailure(t *testing.T) {
	repo := NewMockRepository(t)
	expected := errors.New("database unavailable")
	repo.EXPECT().HasActiveReferences(mock.Anything, "easy", "10m").Return(false, expected)

	result, err := NewService(repo).Create(context.Background(), uuid.New(), Recipe{
		DifficultyID: "easy",
		DurationID:   "10m",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, expected)
}

func TestServiceCreateAssignsCreatorBeforePersistence(t *testing.T) {
	repo := NewMockRepository(t)
	creatorID := uuid.New()
	repo.EXPECT().HasActiveReferences(mock.Anything, "easy", "10m").Return(true, nil)
	repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(recipe Recipe) bool {
		return recipe.CreatorID == creatorID
	})).Return(&Recipe{}, nil)

	_, err := NewService(repo).Create(context.Background(), creatorID, Recipe{
		DifficultyID: "easy",
		DurationID:   "10m",
	})

	assert.NoError(t, err)
}
