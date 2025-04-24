package testing

import (
	"fmt"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func TestResult_Status(t *testing.T) {
	t.Run(StatusPass, func(t *testing.T) {
		result := Result{
			ID:        uuid.Must(uuid.NewV7()),
			Name:      "TestSuite",
			Error:     nil,
			StartTime: time.Now(),
			EndTime:   time.Now().Add(1 * time.Second),
		}
		require.Equal(t, StatusPass, result.Status())
	})

	t.Run(StatusFail, func(t *testing.T) {
		result := Result{
			ID:        uuid.Must(uuid.NewV7()),
			Name:      "TestSuite",
			Error:     fmt.Errorf("test error"),
			StartTime: time.Now(),
			EndTime:   time.Now().Add(1 * time.Second),
		}
		require.Equal(t, StatusFail, result.Status())
	})
}

func TestResult_Duration(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(2 * time.Second)
	result := Result{
		ID:        uuid.Must(uuid.NewV7()),
		Name:      "TestSuite",
		Error:     nil,
		StartTime: startTime,
		EndTime:   endTime,
	}
	require.Equal(t, 2*time.Second, result.Duration())
}
