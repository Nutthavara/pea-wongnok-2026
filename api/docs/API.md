# Recipe API contract

This hand-authored contract covers recipe and reference-data endpoints being implemented. Generated Swagger files intentionally describe only implemented routes.

## Conventions

- Base path: `/api/v1`.
- All endpoints require `Authorization: Bearer <access-token>`.
- Request bodies use `Content-Type: application/json`; responses are JSON.
- Fields use camelCase; timestamps are RFC 3339 strings.
- Recipe, ingredient, and instruction IDs are JSON integers, never strings.
- Errors are `{ "message": "..." }`. The response type reserves a `code` field, but no endpoint populates it yet — treat the HTTP status as authoritative.

Unless otherwise specified: `400` is invalid input, `401` is a missing/invalid token, and `500` is an unexpected failure.

## Representations

Reference data:

```json
{ "id": "easy", "name": "Easy" }
```

Recipe:

```json
{
  "id": 42,
  "name": "Tom yum soup",
  "description": "A bright, spicy Thai soup.",
  "imageUrl": "https://images.example.com/tom-yum.jpg",
  "difficulty": { "id": "medium", "name": "Medium" },
  "duration": { "id": "30m", "name": "10 - 30 mins" },
  "ingredients": [{ "id": 101, "description": "2 cups stock" }],
  "instructions": [{ "id": 201, "description": "Bring the stock to a simmer." }],
  "creator": { "id": "3f0c1a7e-2b19-4c5e-9f3a-000000000000", "name": "Somchai" },
  "createdAt": "2026-08-20T10:00:00Z",
  "updatedAt": "2026-08-20T10:00:00Z"
}
```

`imageUrl` is `null` when absent. `creator.name` is a string. Only active ingredients and instructions are returned; no `deletedAt` field is exposed.

Create and replace use a complete write body. `name`, `description`, `difficultyId`, `durationId`, `ingredients`, and `instructions` are required; `imageUrl` is optional. Ingredient/instruction arrays may be empty, and each item has a required non-empty `description`.

```json
{
  "name": "Tom yum soup",
  "description": "A bright, spicy Thai soup.",
  "imageUrl": "https://images.example.com/tom-yum.jpg",
  "difficultyId": "medium",
  "durationId": "30m",
  "ingredients": [{ "description": "2 cups stock" }],
  "instructions": [{ "description": "Bring the stock to a simmer." }]
}
```

`difficultyId` and `durationId` must reference active master-data records; otherwise the response is `400 invalid_request` (not `404`, despite the name). The `ingredients`/`instructions` keys must be present in the body — an explicit `[]` is fine, but omitting the key or sending `null` fails validation. When present, `imageUrl` must be a valid URL.

## Reference data

### `GET /difficulties`

Returns active difficulties as `{ "total": 3, "results": [<reference-data>, ...] }`.

| Status   | Meaning            |
| -------- | ------------------ |
| 200      | Active master data |
| 401, 500 | Error body         |

Fixed values: `easy` (Easy), `medium` (Medium), `hard` (Hard).

### `GET /durations`

Returns active durations as `{ "total": 4, "results": [<reference-data>, ...] }`.

| Status   | Meaning            |
| -------- | ------------------ |
| 200      | Active master data |
| 401, 500 | Error body         |

Fixed values: `10m` (5 - 10 mins), `30m` (10 - 30 mins), `60m` (~1 Hour), and `long` (More than 1 hour).

## Recipes

### `POST /recipes`

Send the complete write body. The authenticated user becomes the creator. On success, the response body is **`{ "id": <int> }` only** — not the full recipe.

| Status | Meaning                                                                          |
| ------ | -------------------------------------------------------------------------------- |
| 201    | `{ "id": <int> }`                                                                |
| 400    | Missing/malformed body, or referenced `difficultyId`/`durationId` does not exist |
| 401    | Error body                                                                       |
| 500    | Error body                                                                       |

An invalid/malformed body and an unresolvable `difficultyId`/`durationId` both currently return `400 { "message": "invalid request" }` — there is no distinct `404` case for this endpoint despite reference-data lookups happening server-side.

### `GET /recipes/{recipeId}`

`recipeId` is a required integer.

| Status | Meaning                           |
| ------ | --------------------------------- |
| 200    | Complete recipe response          |
| 400    | Invalid `recipeId`                |
| 401    | Error body                        |
| 404    | Recipe is missing or soft-deleted |
| 500    | Error body                        |

### `GET /recipes`

Requires `Authorization: Bearer <access-token>` like all recipe routes — this list endpoint is not public. Returns only active recipes.

| Parameter    | Type            | Meaning                                                              |
| ------------ | --------------- | -------------------------------------------------------------------- |
| `name`       | string          | Optional case-insensitive substring filter on recipe name            |
| `difficulty` | string          | Optional difficulty ID filter; must reference an existing difficulty |
| `sort`       | `ASC` or `DESC` | Case-sensitive; orders by `createdAt`; default `DESC`                |
| `page`       | int             | Optional, minimum 1, default 1                                       |
| `limit`      | int             | Optional, 1-100, default 12                                          |

The success body is `{ "total": 1, "results": [<complete-recipe>, ...] }`; `total` is the count after all filters.

| Status | Meaning                                                                                        |
| ------ | ---------------------------------------------------------------------------------------------- |
| 200    | Filtered results                                                                               |
| 400    | Invalid query — malformed `page`/`limit`, unsupported `sort` value, or an unknown `difficulty` |
| 401    | Error body                                                                                     |
| 500    | Error body                                                                                     |

An unknown `difficulty` filter currently returns `400`, not `404`.

Known current limitation: `page` has no effect on the returned rows (a server-side pagination bug bypasses offsetting); every page currently returns the same first `limit` rows ordered by `sort`.

### `PUT /recipes/{recipeId}`

`recipeId` is a required integer. Send a complete write body: omitted required fields are not preserved. The response is the complete replacement recipe, including integer `id`.

| Status | Meaning                                                               |
| ------ | --------------------------------------------------------------------- |
| 200    | Complete recipe response                                              |
| 400    | Invalid path parameter or body                                        |
| 401    | Error body                                                            |
| 403    | Another user owns the active recipe                                   |
| 404    | Recipe is missing/soft-deleted, or difficulty/duration is unavailable |
| 500    | Error body                                                            |

### `DELETE /recipes/{recipeId}`

`recipeId` is a required integer. Deletion sets only `recipes.deleted_at`; that recipe is then omitted from detail and list results.

| Status | Meaning                             |
| ------ | ----------------------------------- |
| 204    | No response body                    |
| 400    | Invalid path parameter              |
| 401    | Error body                          |
| 403    | Another user owns the active recipe |
| 404    | Recipe is missing or soft-deleted   |
| 500    | Error body                          |

For update and delete, ownership is checked after locating an active recipe: another user's active recipe returns `403`; missing or soft-deleted recipes return `404`.
