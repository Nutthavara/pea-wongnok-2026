package recipe

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRecipeRequestBindsAndMapsCompleteWriteBody(t *testing.T) {
	request := bindCreateRecipeRequest(t, `{
		"name": "Tom yum soup",
		"description": "A bright, spicy Thai soup.",
		"imageUrl": "https://images.example.com/tom-yum.jpg",
		"difficultyId": "medium",
		"durationId": "30m",
		"ingredients": [{"description": "2 cups stock"}],
		"instructions": [{"description": "Bring the stock to a simmer."}]
	}`)

	recipe := request.ToRecipe()

	assert.Equal(t, "Tom yum soup", recipe.Name)
	assert.Equal(t, "A bright, spicy Thai soup.", recipe.Description)
	require.NotNil(t, recipe.ImageURL)
	assert.Equal(t, "https://images.example.com/tom-yum.jpg", *recipe.ImageURL)
	assert.Equal(t, "medium", recipe.DifficultyID)
	assert.Equal(t, "30m", recipe.DurationID)
	assert.Equal(t, []RecipeIngredient{{Description: "2 cups stock"}}, recipe.Ingredients)
	assert.Equal(t, []RecipeInstruction{{Description: "Bring the stock to a simmer."}}, recipe.Instructions)
	assert.Zero(t, recipe.ID)
	assert.Zero(t, recipe.CreatorID)
	assert.Zero(t, recipe.AverageRating)
}

func TestReplaceRecipeRequestBindsCompleteWriteBody(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("PUT", "/recipes/1", bytes.NewBufferString(`{
		"name": "Tom yum soup",
		"description": "A bright, spicy Thai soup.",
		"difficultyId": "medium",
		"durationId": "30m",
		"ingredients": [],
		"instructions": []
	}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	var request ReplaceRecipeRequest
	require.NoError(t, ctx.ShouldBindJSON(&request))
	assert.Empty(t, request.ToRecipe().Ingredients)
	assert.Empty(t, request.ToRecipe().Instructions)
}

func TestCreateRecipeRequestAllowsExplicitEmptyChildCollections(t *testing.T) {
	request := bindCreateRecipeRequest(t, `{
		"name": "Tom yum soup",
		"description": "A bright, spicy Thai soup.",
		"difficultyId": "medium",
		"durationId": "30m",
		"ingredients": [],
		"instructions": []
	}`)

	recipe := request.ToRecipe()
	require.NotNil(t, recipe.Ingredients)
	require.NotNil(t, recipe.Instructions)
	assert.Empty(t, recipe.Ingredients)
	assert.Empty(t, recipe.Instructions)
}

func TestCreateRecipeRequestRejectsInvalidWriteBodies(t *testing.T) {
	testCases := map[string]string{
		"missing name":                  `{"description":"Description","difficultyId":"easy","durationId":"10m","ingredients":[],"instructions":[]}`,
		"missing description":           `{"name":"Name","difficultyId":"easy","durationId":"10m","ingredients":[],"instructions":[]}`,
		"missing difficulty":            `{"name":"Name","description":"Description","durationId":"10m","ingredients":[],"instructions":[]}`,
		"missing duration":              `{"name":"Name","description":"Description","difficultyId":"easy","ingredients":[],"instructions":[]}`,
		"omitted ingredients":           `{"name":"Name","description":"Description","difficultyId":"easy","durationId":"10m","instructions":[]}`,
		"null ingredients":              `{"name":"Name","description":"Description","difficultyId":"easy","durationId":"10m","ingredients":null,"instructions":[]}`,
		"omitted instructions":          `{"name":"Name","description":"Description","difficultyId":"easy","durationId":"10m","ingredients":[]}`,
		"null instructions":             `{"name":"Name","description":"Description","difficultyId":"easy","durationId":"10m","ingredients":[],"instructions":null}`,
		"empty ingredient description":  `{"name":"Name","description":"Description","difficultyId":"easy","durationId":"10m","ingredients":[{"description":""}],"instructions":[]}`,
		"empty instruction description": `{"name":"Name","description":"Description","difficultyId":"easy","durationId":"10m","ingredients":[],"instructions":[{"description":""}]}`,
		"malformed image URL":           `{"name":"Name","description":"Description","imageUrl":"not-a-url","difficultyId":"easy","durationId":"10m","ingredients":[],"instructions":[]}`,
	}

	for name, body := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest("POST", "/recipes", bytes.NewBufferString(body))
			ctx.Request.Header.Set("Content-Type", "application/json")

			var request CreateRecipeRequest
			assert.Error(t, ctx.ShouldBindJSON(&request))
		})
	}
}

func TestRateRecipeRequestBindsRating(t *testing.T) {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("POST", "/recipes/1/rating", bytes.NewBufferString(`{"rating":5}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	var request RateRecipeRequest
	require.NoError(t, ctx.ShouldBindJSON(&request))
	assert.Equal(t, 5, request.Rating)
}

func TestRateRecipeRequestRejectsInvalidRatings(t *testing.T) {
	testCases := map[string]string{
		"missing rating":    `{}`,
		"zero rating":       `{"rating":0}`,
		"negative rating":   `{"rating":-1}`,
		"rating above five": `{"rating":6}`,
	}

	for name, body := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest("POST", "/recipes/1/rating", bytes.NewBufferString(body))
			ctx.Request.Header.Set("Content-Type", "application/json")

			var request RateRecipeRequest
			assert.Error(t, ctx.ShouldBindJSON(&request))
		})
	}
}

func TestGetRecipesQueryBindsTriStateFavoriteFilter(t *testing.T) {
	testCases := map[string]*bool{
		"":               nil,
		"favorite=true":  boolPtr(true),
		"favorite=false": boolPtr(false),
	}

	for rawQuery, expected := range testCases {
		t.Run(rawQuery, func(t *testing.T) {
			target := "/recipes"
			if rawQuery != "" {
				target += "?" + rawQuery
			}
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest("GET", target, nil)

			var query GetRecipesQuery
			require.NoError(t, ctx.ShouldBindQuery(&query))
			if expected == nil {
				assert.Nil(t, query.Favorite)
			} else {
				require.NotNil(t, query.Favorite)
				assert.Equal(t, *expected, *query.Favorite)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func bindCreateRecipeRequest(t *testing.T, body string) CreateRecipeRequest {
	t.Helper()
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("POST", "/recipes", bytes.NewBufferString(body))
	ctx.Request.Header.Set("Content-Type", "application/json")

	var request CreateRecipeRequest
	require.NoError(t, ctx.ShouldBindJSON(&request))
	return request
}
