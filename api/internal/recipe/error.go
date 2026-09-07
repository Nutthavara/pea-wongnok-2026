package recipe

import "errors"

var (
	// Create
	ErrInvalidReferenceData = errors.New("invalid recipe reference data")

	// Get
	ErrRecipeNotFound = errors.New("recipe not found")

	// Update
	ErrReferenceDataUnavailable = errors.New("recipe reference data unavailable")
	ErrForbidden                = errors.New("recipe access forbidden")
)
