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

func TestServiceGetReturnsRecipeWithFavoriteStatus(t *testing.T) {
	repo := NewMockRepository(t)
	userID := uuid.New()
	found := &Recipe{ID: 42, Name: "Tom yum soup"}
	repo.EXPECT().FindByID(mock.Anything, 42).Return(found, nil)
	repo.EXPECT().IsFavorite(mock.Anything, userID, 42).Return(true, nil)

	result, err := NewService(repo).Get(context.Background(), 42, userID)

	assert.NoError(t, err)
	assert.Equal(t, &Recipe{ID: 42, Name: "Tom yum soup", IsFavorite: true}, result)
}

func TestServiceGetPropagatesRecipeNotFound(t *testing.T) {
	repo := NewMockRepository(t)
	repo.EXPECT().FindByID(mock.Anything, 42).Return(nil, ErrRecipeNotFound)

	result, err := NewService(repo).Get(context.Background(), 42, uuid.New())

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrRecipeNotFound)
}

func TestServiceGetPropagatesFavoriteLookupFailure(t *testing.T) {
	repo := NewMockRepository(t)
	userID := uuid.New()
	repo.EXPECT().FindByID(mock.Anything, 42).Return(&Recipe{ID: 42}, nil)
	expected := errors.New("database unavailable")
	repo.EXPECT().IsFavorite(mock.Anything, userID, 42).Return(false, expected)

	result, err := NewService(repo).Get(context.Background(), 42, userID)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, expected)
}

func TestServiceListReturnsRecipesFromRepositoryWithDefaults(t *testing.T) {
	repo := NewMockRepository(t)
	userID := uuid.New()
	expected := []Recipe{{ID: 42, Name: "Tom yum soup"}}
	repo.EXPECT().List(mock.Anything, userID, GetRecipesQuery{
		Pagination: Pagination{Page: 1, Limit: 12},
		Sort:       DescendingSortDirection,
	}).Return(expected, int64(1), nil)

	result, total, err := NewService(repo).List(context.Background(), userID, GetRecipesQuery{})

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	assert.EqualValues(t, 1, total)
}

func TestServiceListPassesFavoriteTrueFilterThrough(t *testing.T) {
	repo := NewMockRepository(t)
	userID := uuid.New()
	repo.EXPECT().List(mock.Anything, userID, mock.MatchedBy(func(query GetRecipesQuery) bool {
		return query.Favorite != nil && *query.Favorite
	})).Return(nil, int64(0), nil)

	_, _, err := NewService(repo).List(context.Background(), userID, GetRecipesQuery{Favorite: boolPtr(true)})

	assert.NoError(t, err)
}

func TestServiceListPassesFavoriteFalseFilterThrough(t *testing.T) {
	repo := NewMockRepository(t)
	userID := uuid.New()
	repo.EXPECT().List(mock.Anything, userID, mock.MatchedBy(func(query GetRecipesQuery) bool {
		return query.Favorite != nil && !*query.Favorite
	})).Return(nil, int64(0), nil)

	_, _, err := NewService(repo).List(context.Background(), userID, GetRecipesQuery{Favorite: boolPtr(false)})

	assert.NoError(t, err)
}

func TestServiceListRejectsUnknownDifficultyWithoutQueryingRepository(t *testing.T) {
	repo := NewMockRepository(t)
	repo.EXPECT().DifficultyExists(mock.Anything, "missing").Return(false, nil)

	result, total, err := NewService(repo).List(context.Background(), uuid.New(), GetRecipesQuery{Difficulty: "missing"})

	assert.Nil(t, result)
	assert.Zero(t, total)
	assert.ErrorIs(t, err, ErrInvalidReferenceData)
}

func TestServiceListPropagatesRepositoryFailure(t *testing.T) {
	repo := NewMockRepository(t)
	expected := errors.New("database unavailable")
	repo.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(nil, int64(0), expected)

	result, total, err := NewService(repo).List(context.Background(), uuid.New(), GetRecipesQuery{})

	assert.Nil(t, result)
	assert.Zero(t, total)
	assert.ErrorIs(t, err, expected)
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

func TestServiceUnfavoriteRemovesRecipeFromCallerFavorites(t *testing.T) {
	repo := NewMockRepository(t)
	userID := uuid.New()
	repo.EXPECT().Unfavorite(mock.Anything, userID, 42).Return(nil)

	err := NewService(repo).Unfavorite(context.Background(), 42, userID)

	assert.NoError(t, err)
}

func TestServiceUnfavoritePropagatesRepositoryFailure(t *testing.T) {
	repo := NewMockRepository(t)
	userID := uuid.New()
	expected := errors.New("database unavailable")
	repo.EXPECT().Unfavorite(mock.Anything, userID, 42).Return(expected)

	err := NewService(repo).Unfavorite(context.Background(), 42, userID)

	assert.ErrorIs(t, err, expected)
}

func TestServiceRateRatesRecipeForCaller(t *testing.T) {
	repo := NewMockRepository(t)
	userID := uuid.New()
	repo.EXPECT().Rate(mock.Anything, userID, 42, 5.0).Return(nil)

	err := NewService(repo).Rate(context.Background(), 42, userID, 5)

	assert.NoError(t, err)
}

func TestServiceRatePropagatesRepositoryFailure(t *testing.T) {
	repo := NewMockRepository(t)
	userID := uuid.New()
	expected := errors.New("database unavailable")
	repo.EXPECT().Rate(mock.Anything, userID, 42, 5.0).Return(expected)

	err := NewService(repo).Rate(context.Background(), 42, userID, 5)

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
