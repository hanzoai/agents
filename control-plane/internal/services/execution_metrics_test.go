package services

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecordQueueDepthClampsNegative(t *testing.T) {
	recordQueueDepth(-10)
	require.Equal(t, float64(0), queueDepthGauge.Get())
}

func TestRecordWaiterCountClampsNegative(t *testing.T) {
	recordWaiterCount(-5)
	require.Equal(t, float64(0), waiterInflightGauge.Get())
}

func TestNormalizeAgentLabel(t *testing.T) {
	require.Equal(t, "worker", normalizeAgentLabel(" worker "))
	require.Equal(t, "unknown", normalizeAgentLabel(""))
}
