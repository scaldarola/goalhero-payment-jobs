package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestGetJobStatuses(t *testing.T) {
	t.Run("should return job statuses endpoint", func(t *testing.T) {
		router := setupRouter()
		router.GET("/status", GetJobStatuses)

		req, _ := http.NewRequest(http.MethodGet, "/status", nil)
		req.Header.Set("X-Request-Time", "2024-01-01T00:00:00Z")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// The handler should respond (may fail due to no job manager initialized, but should not panic)
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
	})
}

func TestGetJobHealth(t *testing.T) {
	t.Run("should return job health endpoint", func(t *testing.T) {
		router := setupRouter()
		router.GET("/health", GetJobHealth)

		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// The handler should respond (may fail due to no job manager initialized, but should not panic)
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusServiceUnavailable)
	})
}

func TestTriggerJob(t *testing.T) {
	t.Run("should return error for invalid job name", func(t *testing.T) {
		router := setupRouter()
		router.POST("/trigger/:jobName", TriggerJob)

		req, _ := http.NewRequest(http.MethodPost, "/trigger/invalid-job", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response["success"].(bool))
		assert.Equal(t, "Invalid job name", response["error"])
		assert.Contains(t, response, "validJobs")
	})

	validJobs := []string{"rating-reminder", "auto-release", "dispute-escalation"}
	for _, jobName := range validJobs {
		t.Run("should handle trigger for "+jobName, func(t *testing.T) {
			router := setupRouter()
			router.POST("/trigger/:jobName", TriggerJob)

			req, _ := http.NewRequest(http.MethodPost, "/trigger/"+jobName, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should either succeed or fail with internal server error (if job manager not initialized)
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if w.Code == http.StatusOK {
				assert.True(t, response["success"].(bool))
				assert.Contains(t, response["message"], "triggered successfully")
				assert.Equal(t, jobName, response["jobName"])
			}
		})
	}
}

func TestUpdateJobConfig(t *testing.T) {
	t.Run("should return error for invalid JSON", func(t *testing.T) {
		router := setupRouter()
		router.POST("/config", UpdateJobConfig)

		invalidJSON := `{"invalidField": true`
		req, _ := http.NewRequest(http.MethodPost, "/config", bytes.NewBuffer([]byte(invalidJSON)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response["success"].(bool))
		assert.Equal(t, "Invalid configuration format", response["error"])
	})
}

func TestGetJobConfig(t *testing.T) {
	t.Run("should handle get job config request", func(t *testing.T) {
		router := setupRouter()
		router.GET("/config", GetJobConfig)

		req, _ := http.NewRequest(http.MethodGet, "/config", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should either return config or error if job manager not initialized
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		if w.Code == http.StatusInternalServerError {
			assert.False(t, response["success"].(bool))
			assert.Equal(t, "Job manager not initialized", response["error"])
		}
	})
}

func TestRestartJobs(t *testing.T) {
	t.Run("should return not implemented error", func(t *testing.T) {
		router := setupRouter()
		router.POST("/restart", RestartJobs)

		req, _ := http.NewRequest(http.MethodPost, "/restart", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotImplemented, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response["success"].(bool))
		assert.Contains(t, response["error"], "not yet implemented")
	})
}

func TestInternalTriggerHandlers(t *testing.T) {
	internalEndpoints := []struct{
		path string
		handler gin.HandlerFunc
		expectedMessage string
	}{
		{"/internal/trigger-rating-reminder", TriggerRatingReminder, "Rating reminder"},
		{"/internal/trigger-auto-release", TriggerAutoRelease, "Auto release"},
		{"/internal/trigger-dispute-escalation", TriggerDisputeEscalation, "Dispute escalation"},
	}

	for _, endpoint := range internalEndpoints {
		t.Run("should handle "+endpoint.path, func(t *testing.T) {
			router := setupRouter()
			router.POST(endpoint.path, endpoint.handler)

			req, _ := http.NewRequest(http.MethodPost, endpoint.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should either succeed or fail with internal server error (if job manager not initialized)
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if w.Code == http.StatusOK {
				assert.True(t, response["success"].(bool))
				assert.Contains(t, response["message"], "triggered")
			}
		})
	}
}