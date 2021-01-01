package config

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const configsPath = "/configs"

func TestConfigHandler_Get(t *testing.T) {
	t.Run("should propagate service error", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, configsPath, nil)
		w := httptest.NewRecorder()

		router, configService := createTestRouter()
		configService.On("GetAll").
			Return(Configs{}, assert.AnError).
			Once()

		router.ServeHTTP(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		configService.AssertExpectations(t)
	})

	t.Run("should return found configs", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, configsPath, nil)
		w := httptest.NewRecorder()

		router, configService := createTestRouter()
		expectedConfigs := Configs{Data: []Config{testCfg}}
		configService.On("GetAll").
			Return(expectedConfigs, nil).
			Once()

		router.ServeHTTP(w, r)

		expectedJSON, err := json.Marshal(expectedConfigs)
		require.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), w.Body.String())
		configService.AssertExpectations(t)
	})

	t.Run("should return empty array when no config found", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, configsPath, nil)
		w := httptest.NewRecorder()

		router, configService := createTestRouter()
		configService.On("GetAll").
			Return(Configs{Data: []Config{}}, nil).
			Once()

		router.ServeHTTP(w, r)

		assert.JSONEq(t, `{"data":[]}`, w.Body.String())
		assert.Equal(t, http.StatusOK, w.Code)
		configService.AssertExpectations(t)
	})
}

func TestConfigHandler_Create(t *testing.T) {
	testCfgBytes, err := json.Marshal(testCfg)
	require.NoError(t, err)

	t.Run("should return Bad Request on invalid config", func(t *testing.T) {
		r := httptest.NewRequest(
			http.MethodPost, configsPath, strings.NewReader("invalid payload"),
		)

		w := httptest.NewRecorder()
		router, _ := createTestRouter()

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should propagate service error", func(t *testing.T) {
		r := httptest.NewRequest(
			http.MethodPost, configsPath, bytes.NewReader(testCfgBytes),
		)
		w := httptest.NewRecorder()

		router, configService := createTestRouter()
		configService.On("Create", testCfg).
			Return(Config{}, assert.AnError).
			Once()

		router.ServeHTTP(w, r)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		configService.AssertExpectations(t)
	})

	t.Run("should return OK on success", func(t *testing.T) {
		r := httptest.NewRequest(
			http.MethodPost, configsPath, bytes.NewReader(testCfgBytes),
		)
		w := httptest.NewRecorder()

		router, configService := createTestRouter()
		configService.On("Create", testCfg).
			Return(testCfg, nil).
			Once()

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		configService.AssertExpectations(t)
	})
}

func createTestRouter() (http.Handler, *mockConfigService) {
	router := chi.NewRouter()
	svc := mockConfigService{}
	handler := NewHandler(&svc)
	router.Get(configsPath, handler.GetAll)
	router.Post(configsPath, handler.Create)

	return router, &svc
}
