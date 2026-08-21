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

func TestServiceGetReturnsRecipeFromRepository(t *testing.T) {
	repo := NewMockRepository(t)
	expected := &Recipe{ID: 42, Name: "Tom yum soup"}
	repo.EXPECT().FindByID(mock.Anything, 42).Return(expected, nil)

	result, err := NewService(repo).Get(context.Background(), 42)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestServiceGetPropagatesRecipeNotFound(t *testing.T) {
	repo := NewMockRepository(t)
	repo.EXPECT().FindByID(mock.Anything, 42).Return(nil, ErrRecipeNotFound)

	result, err := NewService(repo).Get(context.Background(), 42)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrRecipeNotFound)
}

func TestServiceReplaceUpdatesRecipeOwnedByCaller(t *testing.T) {
	repo := NewMockRepository(t)
	creatorID := uuid.New()
	existing := &Recipe{ID: 42, CreatorID: creatorID}
	input := Recipe{DifficultyID: "easy", DurationID: "10m"}
	persisted := input
	persisted.ID = 42
	persisted.CreatorID = creatorID
	replaced := persisted

	repo.EXPECT().FindByID(mock.Anything, 42).Return(existing, nil)
	repo.EXPECT().HasActiveReferences(mock.Anything, "easy", "10m").Return(true, nil)
	repo.EXPECT().Replace(mock.Anything, persisted).Return(&replaced, nil)

	result, err := NewService(repo).Replace(context.Background(), 42, creatorID, input)

	assert.NoError(t, err)
	assert.Equal(t, &replaced, result)
}

func TestServiceReplacePropagatesRecipeNotFound(t *testing.T) {
	repo := NewMockRepository(t)
	repo.EXPECT().FindByID(mock.Anything, 42).Return(nil, ErrRecipeNotFound)

	result, err := NewService(repo).Replace(context.Background(), 42, uuid.New(), Recipe{})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrRecipeNotFound)
}

func TestServiceReplaceRejectsAnotherUsersRecipe(t *testing.T) {
	repo := NewMockRepository(t)
	repo.EXPECT().FindByID(mock.Anything, 42).Return(&Recipe{ID: 42, CreatorID: uuid.New()}, nil)

	result, err := NewService(repo).Replace(context.Background(), 42, uuid.New(), Recipe{})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrForbidden)
}

func TestServiceReplaceRejectsInactiveReferencesWithoutPersisting(t *testing.T) {
	repo := NewMockRepository(t)
	creatorID := uuid.New()
	repo.EXPECT().FindByID(mock.Anything, 42).Return(&Recipe{ID: 42, CreatorID: creatorID}, nil)
	repo.EXPECT().HasActiveReferences(mock.Anything, "missing", "10m").Return(false, nil)

	result, err := NewService(repo).Replace(context.Background(), 42, creatorID, Recipe{
		DifficultyID: "missing",
		DurationID:   "10m",
	})

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrReferenceDataUnavailable)
}

func TestServiceDeleteRemovesRecipeOwnedByCaller(t *testing.T) {
	repo := NewMockRepository(t)
	creatorID := uuid.New()
	repo.EXPECT().FindByID(mock.Anything, 42).Return(&Recipe{ID: 42, CreatorID: creatorID}, nil)
	repo.EXPECT().Delete(mock.Anything, 42).Return(nil)

	err := NewService(repo).Delete(context.Background(), 42, creatorID)

	assert.NoError(t, err)
}

func TestServiceDeletePropagatesRecipeNotFound(t *testing.T) {
	repo := NewMockRepository(t)
	repo.EXPECT().FindByID(mock.Anything, 42).Return(nil, ErrRecipeNotFound)

	err := NewService(repo).Delete(context.Background(), 42, uuid.New())

	assert.ErrorIs(t, err, ErrRecipeNotFound)
}

func TestServiceDeleteRejectsAnotherUsersRecipe(t *testing.T) {
	repo := NewMockRepository(t)
	repo.EXPECT().FindByID(mock.Anything, 42).Return(&Recipe{ID: 42, CreatorID: uuid.New()}, nil)

	err := NewService(repo).Delete(context.Background(), 42, uuid.New())

	assert.ErrorIs(t, err, ErrForbidden)
}

func TestServiceFavoriteAddsRecipeToCallerFavorites(t *testing.T) {
	repo := NewMockRepository(t)
	userID := uuid.New()
	repo.EXPECT().Favorite(mock.Anything, userID, 42).Return(nil)

	err := NewService(repo).Favorite(context.Background(), 42, userID)

	assert.NoError(t, err)
}

func TestServiceFavoritePropagatesRepositoryFailure(t *testing.T) {
	repo := NewMockRepository(t)
	userID := uuid.New()
	expected := errors.New("database unavailable")
	repo.EXPECT().Favorite(mock.Anything, userID, 42).Return(expected)

	err := NewService(repo).Favorite(context.Background(), 42, userID)

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
