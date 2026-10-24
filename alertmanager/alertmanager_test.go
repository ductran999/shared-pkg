package alertmanager_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ductran999/shared-pkg/alertmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewAlertManager(t *testing.T) {
	_, err := alertmanager.NewAlertManager("")
	require.Error(t, err)
	assert.ErrorIs(t, err, alertmanager.ErrEmptyHost)
}

func Test_AlertManager_Send(t *testing.T) {
	var (
		gotPath   string
		gotMethod string
		gotBody   []byte
		gotType   string
		status    = http.StatusOK
	)

	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(status)
	}))
	defer mock.Close()

	// The alert manager is a singleton, so it has to be created against the
	// mock server before any other test creates one.
	am, err := alertmanager.NewAlertManager(mock.URL)
	require.NoError(t, err)
	require.NotNil(t, am)

	t.Run("sends the alert", func(t *testing.T) {
		err := am.Send(
			alertmanager.WithLabels(alertmanager.Labels{"alertname": "HighCPU", "severity": "critical"}),
			alertmanager.WithAnnotations(alertmanager.Annotations{"summary": "cpu is high"}),
			alertmanager.WithDuration(time.Minute),
		)
		require.NoError(t, err)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/api/v2/alerts", gotPath)
		assert.Equal(t, "application/json", gotType)

		var payload []map[string]any
		require.NoError(t, json.Unmarshal(gotBody, &payload))
		require.Len(t, payload, 1)
		assert.Equal(t, "HighCPU", payload[0]["labels"].(map[string]any)["alertname"])
		assert.Equal(t, "cpu is high", payload[0]["annotations"].(map[string]any)["summary"])
		assert.NotEmpty(t, payload[0]["endsAt"])
	})

	t.Run("requires labels", func(t *testing.T) {
		assert.Error(t, am.Send())
	})

	t.Run("requires the alertname label", func(t *testing.T) {
		assert.Error(t, am.Send(alertmanager.WithLabels(alertmanager.Labels{"severity": "critical"})))
	})

	t.Run("fails on a non 2xx response", func(t *testing.T) {
		status = http.StatusInternalServerError
		t.Cleanup(func() { status = http.StatusOK })

		err := am.Send(alertmanager.WithLabels(alertmanager.Labels{"alertname": "HighCPU"}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "non-2xx status")
	})
}
