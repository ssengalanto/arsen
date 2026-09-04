package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"arsen/pkg/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPayload struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestJSON_WritesCorrectContentType(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/test", http.NoBody)

	response.JSON(w, r, http.StatusOK, testPayload{ID: 1, Name: "alice"})

	res := w.Result()
	defer res.Body.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestJSON_WritesCorrectStatusCode(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/test", http.NoBody)

	response.JSON(w, r, http.StatusOK, testPayload{ID: 1, Name: "alice"})

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJSON_MarshalsData(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/test", http.NoBody)

	response.JSON(w, r, http.StatusOK, testPayload{ID: 42, Name: "bob"})

	res := w.Result()
	defer res.Body.Close() //nolint:errcheck // test cleanup

	var body testPayload
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))

	assert.Equal(t, 42, body.ID)
	assert.Equal(t, "bob", body.Name)
}

func TestCreated_SetsLocationHeader(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/users", http.NoBody)

	response.Created(w, r, "/api/users/123", testPayload{ID: 123, Name: "charlie"})

	res := w.Result()
	defer res.Body.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, "/api/users/123", res.Header.Get("Location"))
}

func TestCreated_Sets201Status(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/users", http.NoBody)

	response.Created(w, r, "/api/users/123", testPayload{ID: 123, Name: "charlie"})

	assert.Equal(t, http.StatusCreated, w.Code)
}
