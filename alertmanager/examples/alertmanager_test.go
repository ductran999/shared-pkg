package alertmanager_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ductran999/shared-pkg/alertmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SendAlert(t *testing.T) {
	var payload []struct {
		Labels      map[string]any `json:"labels"`
		Annotations map[string]any `json:"annotations"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v2/alerts", r.URL.Path)
		assert.Contains(t, r.Header.Get("Content-Type"), "application/json")
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	am, err := alertmanager.NewAlertManager(srv.URL)
	require.NoError(t, err)

	t.Run("sends alert with required labels", func(t *testing.T) {
		err := am.Send(
			alertmanager.WithLabels(alertmanager.Labels{
				"alertname": "TestAlert",
				"level":     "critical",
			}),
			alertmanager.WithAnnotations(alertmanager.Annotations{
				"summary":     "TestAlert",
				"description": "Example alert",
			}),
		)
		require.NoError(t, err)
		require.Len(t, payload, 1)
		assert.Equal(t, "TestAlert", payload[0].Labels["alertname"])
		assert.Equal(t, "critical", payload[0].Labels["level"])
		assert.Equal(t, "TestAlert", payload[0].Annotations["summary"])
	})

	t.Run("missing alertname fails", func(t *testing.T) {
		err := am.Send(alertmanager.WithLabels(alertmanager.Labels{
			"level": "critical",
		}))
		require.ErrorContains(t, err, "alertname label is required")
	})

	t.Run("no labels fails", func(t *testing.T) {
		require.Error(t, am.Send())
	})
}
