package recipe

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"wongnok/internal/reqctx"

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

	assertErrorCode(t, response, http.StatusBadRequest, "invalid_request")
}

func TestHandlerCreateRejectsMissingAuthenticatedUserWithoutCallingService(t *testing.T) {
	response := performCreateRequest(t, NewHandler(NewMockService(t)), validCreateRecipeJSON, uuid.Nil, false)

	assertErrorCode(t, response, http.StatusUnauthorized, "unauthorized")
}

func TestHandlerCreateMapsInvalidReferencesToInvalidRequest(t *testing.T) {
	service := NewMockService(t)
	creatorID := uuid.New()
	service.EXPECT().Create(mock.Anything, creatorID, mock.Anything).Return(nil, ErrInvalidReferenceData)

	response := performCreateRequest(t, NewHandler(service), validCreateRecipeJSON, creatorID, true)

	assertErrorCode(t, response, http.StatusBadRequest, "invalid_request")
}

func TestHandlerCreateMapsUnexpectedErrorToInternalError(t *testing.T) {
	service := NewMockService(t)
	creatorID := uuid.New()
	service.EXPECT().Create(mock.Anything, creatorID, mock.Anything).Return(nil, errors.New("database unavailable"))

	response := performCreateRequest(t, NewHandler(service), validCreateRecipeJSON, creatorID, true)

	assertErrorCode(t, response, http.StatusInternalServerError, "internal_error")
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

func assertErrorCode(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	assert.Equal(t, status, response.Code)
	var body map[string]string
	assert.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, code, body["code"])
}
