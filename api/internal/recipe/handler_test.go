package recipe

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"wongnok/internal/convutil"
	"wongnok/internal/reqctx"
	"wongnok/internal/user"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const validCreateRecipeJSON = `{"name":"Tom yum soup","description":"A bright, spicy Thai soup.","difficultyId":"medium","durationId":"30m","ingredients":[{"description":"2 cups stock"}],"instructions":[{"description":"Simmer the stock."}]}`

func TestHandlerCreateCreatesRecipeForAuthenticatedUser(t *testing.T) {
	service := NewMockService(t)
	creatorID := uuid.New()
	service.EXPECT().Create(mock.Anything, creatorID, Recipe{
		Name: "Tom yum soup", Description: "A bright, spicy Thai soup.", DifficultyID: "medium", DurationID: "30m",
		Ingredients: []RecipeIngredient{{Description: "2 cups stock"}}, Instructions: []RecipeInstruction{{Description: "Simmer the stock."}},
	}).Return(&Recipe{ID: 42}, nil)

	response := performCreateRequest(t, NewHandler(service), validCreateRecipeJSON, creatorID, true)

	assert.Equal(t, http.StatusCreated, response.Code)
	assert.JSONEq(t, `{"id":42}`, response.Body.String())
}

func TestHandlerCreateRejectsInvalidRequestWithoutCallingService(t *testing.T) {
	response := performCreateRequest(t, NewHandler(NewMockService(t)), `{`, uuid.New(), true)

	assertErrorMessage(t, response, http.StatusBadRequest, "invalid request")
}

func TestHandlerCreateRejectsMissingAuthenticatedUserWithoutCallingService(t *testing.T) {
	response := performCreateRequest(t, NewHandler(NewMockService(t)), validCreateRecipeJSON, uuid.Nil, false)

	assertErrorMessage(t, response, http.StatusUnauthorized, "unauthorized")
}

func TestHandlerCreateMapsInvalidReferencesToInvalidRequest(t *testing.T) {
	service := NewMockService(t)
	creatorID := uuid.New()
	service.EXPECT().Create(mock.Anything, creatorID, mock.Anything).Return(nil, ErrInvalidReferenceData)

	response := performCreateRequest(t, NewHandler(service), validCreateRecipeJSON, creatorID, true)

	assertErrorMessage(t, response, http.StatusBadRequest, "invalid request")
}

func TestHandlerCreateMapsUnexpectedErrorToInternalError(t *testing.T) {
	service := NewMockService(t)
	creatorID := uuid.New()
	service.EXPECT().Create(mock.Anything, creatorID, mock.Anything).Return(nil, errors.New("database unavailable"))

	response := performCreateRequest(t, NewHandler(service), validCreateRecipeJSON, creatorID, true)

	assertErrorMessage(t, response, http.StatusInternalServerError, "internal server error")
}

func TestHandlerGetRecipeReturnsRecipe(t *testing.T) {
	service := NewMockService(t)
	creatorID := uuid.New()
	createdAt := time.Now()
	recipe := Recipe{
		ID:          42,
		Name:        "Tom yum soup",
		Description: "A bright, spicy Thai soup.",
		Difficulty:  Difficulty{ID: "medium", Name: "Medium"},
		Duration:    Duration{ID: "30m", Name: "10 - 30 mins"},
		Creator:     user.User{ID: creatorID, Name: convutil.ToPointer("Somchai")},
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
	service.EXPECT().Get(mock.Anything, 42).Return(&recipe, nil)

	response := performGetRecipeRequest(t, NewHandler(service), "42")

	assert.Equal(t, http.StatusOK, response.Code)
	expected, err := json.Marshal(NewRecipeResponse(recipe))
	assert.NoError(t, err)
	assert.JSONEq(t, string(expected), response.Body.String())
}

func TestHandlerGetRecipeRejectsNonIntegerIDWithoutCallingService(t *testing.T) {
	response := performGetRecipeRequest(t, NewHandler(NewMockService(t)), "abc")

	assertErrorMessage(t, response, http.StatusBadRequest, "invalid request")
}

func TestHandlerGetRecipeMapsNotFoundToNotFoundResponse(t *testing.T) {
	service := NewMockService(t)
	service.EXPECT().Get(mock.Anything, 42).Return(nil, ErrRecipeNotFound)

	response := performGetRecipeRequest(t, NewHandler(service), "42")

	assertErrorMessage(t, response, http.StatusNotFound, "recipe not found")
}

func TestHandlerGetRecipeMapsUnexpectedErrorToInternalError(t *testing.T) {
	service := NewMockService(t)
	service.EXPECT().Get(mock.Anything, 42).Return(nil, errors.New("database unavailable"))

	response := performGetRecipeRequest(t, NewHandler(service), "42")

	assertErrorMessage(t, response, http.StatusInternalServerError, "internal server error")
}

func TestHandlerReplaceReturnsReplacedRecipe(t *testing.T) {
	service := NewMockService(t)
	userID := uuid.New()
	createdAt := time.Now()
	recipe := Recipe{
		ID:          42,
		Name:        "Tom yum soup",
		Description: "A bright, spicy Thai soup.",
		Difficulty:  Difficulty{ID: "medium", Name: "Medium"},
		Duration:    Duration{ID: "30m", Name: "10 - 30 mins"},
		Creator:     user.User{ID: userID, Name: convutil.ToPointer("Somchai")},
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
	service.EXPECT().Replace(mock.Anything, 42, userID, mock.Anything).Return(&recipe, nil)

	response := performReplaceRequest(t, NewHandler(service), "42", validCreateRecipeJSON, userID, true)

	assert.Equal(t, http.StatusOK, response.Code)
	expected, err := json.Marshal(NewRecipeResponse(recipe))
	assert.NoError(t, err)
	assert.JSONEq(t, string(expected), response.Body.String())
}

func TestHandlerReplaceRejectsMissingAuthenticatedUserWithoutCallingService(t *testing.T) {
	response := performReplaceRequest(t, NewHandler(NewMockService(t)), "42", validCreateRecipeJSON, uuid.Nil, false)

	assertErrorMessage(t, response, http.StatusUnauthorized, "unauthorized")
}

func TestHandlerReplaceRejectsNonIntegerIDWithoutCallingService(t *testing.T) {
	response := performReplaceRequest(t, NewHandler(NewMockService(t)), "abc", validCreateRecipeJSON, uuid.New(), true)

	assertErrorMessage(t, response, http.StatusBadRequest, "invalid request")
}

func TestHandlerReplaceRejectsInvalidRequestWithoutCallingService(t *testing.T) {
	response := performReplaceRequest(t, NewHandler(NewMockService(t)), "42", `{`, uuid.New(), true)

	assertErrorMessage(t, response, http.StatusBadRequest, "invalid request")
}

func TestHandlerReplaceMapsRecipeNotFoundToNotFoundResponse(t *testing.T) {
	service := NewMockService(t)
	service.EXPECT().Replace(mock.Anything, 42, mock.Anything, mock.Anything).Return(nil, ErrRecipeNotFound)

	response := performReplaceRequest(t, NewHandler(service), "42", validCreateRecipeJSON, uuid.New(), true)

	assertErrorMessage(t, response, http.StatusNotFound, "recipe not found")
}

func TestHandlerReplaceMapsReferenceDataUnavailableToNotFoundResponse(t *testing.T) {
	service := NewMockService(t)
	service.EXPECT().Replace(mock.Anything, 42, mock.Anything, mock.Anything).Return(nil, ErrReferenceDataUnavailable)

	response := performReplaceRequest(t, NewHandler(service), "42", validCreateRecipeJSON, uuid.New(), true)

	assertErrorMessage(t, response, http.StatusNotFound, "recipe not found")
}

func TestHandlerReplaceMapsForbiddenToForbiddenResponse(t *testing.T) {
	service := NewMockService(t)
	service.EXPECT().Replace(mock.Anything, 42, mock.Anything, mock.Anything).Return(nil, ErrForbidden)

	response := performReplaceRequest(t, NewHandler(service), "42", validCreateRecipeJSON, uuid.New(), true)

	assertErrorMessage(t, response, http.StatusForbidden, "forbidden")
}

func TestHandlerReplaceMapsUnexpectedErrorToInternalError(t *testing.T) {
	service := NewMockService(t)
	service.EXPECT().Replace(mock.Anything, 42, mock.Anything, mock.Anything).Return(nil, errors.New("database unavailable"))

	response := performReplaceRequest(t, NewHandler(service), "42", validCreateRecipeJSON, uuid.New(), true)

	assertErrorMessage(t, response, http.StatusInternalServerError, "internal server error")
}

func performReplaceRequest(t *testing.T, handler *handler, id string, body string, userID uuid.UUID, authenticated bool) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	request := httptest.NewRequest(http.MethodPut, "/recipes/"+id, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	if authenticated {
		request = request.WithContext(reqctx.WithUserID(request.Context(), userID))
	}
	ctx.Request = request
	ctx.Params = gin.Params{{Key: "id", Value: id}}
	handler.Replace(ctx)
	return response
}

func performGetRecipeRequest(t *testing.T, handler *handler, id string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/recipes/"+id, nil)
	ctx.Params = gin.Params{{Key: "id", Value: id}}
	handler.GetRecipe(ctx)
	return response
}

func performCreateRequest(t *testing.T, handler *handler, body string, userID uuid.UUID, authenticated bool) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	request := httptest.NewRequest(http.MethodPost, "/recipes", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	if authenticated {
		request = request.WithContext(reqctx.WithUserID(request.Context(), userID))
	}
	ctx.Request = request
	handler.Create(ctx)
	return response
}

func assertErrorMessage(t *testing.T, response *httptest.ResponseRecorder, status int, message string) {
	t.Helper()
	assert.Equal(t, status, response.Code)
	var body map[string]string
	assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, message, body["message"])
}
