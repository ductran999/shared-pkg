package alertmanager_test

import (
	"testing"

	"github.com/ductran999/shared-pkg/alertmanager"
	"github.com/stretchr/testify/require"
)

func Test_SendAlert(t *testing.T) {
	am, err := alertmanager.NewAlertManager("http://localhost:9093")
	require.NoError(t, err)

	err = am.Send(
		alertmanager.WithLabels(alertmanager.Labels{
			"level": "critical",
		}),
		alertmanager.WithAnnotations(alertmanager.Annotations{
			"summary":     "TestAlert",
			"description": "Example alert",
		}),
	)

	require.NoError(t, err)
}
