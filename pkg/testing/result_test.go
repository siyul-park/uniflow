package testing

import (
	"fmt"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

func TestResult_Status(t *testing.T) {
	t.Run("PASS", func(t *testing.T) {
		result := Result{
			ID:        uuid.Must(uuid.NewV4()),
			Name:      "TestSuite",
			Error:     nil,
			StartTime: time.Now(),
			EndTime:   time.Now().Add(1 * time.Second),
		}
		assert.Equal(t, "PASS", result.Status())
	})

	t.Run("FAIL", func(t *testing.T) {
		result := Result{
			ID:        uuid.Must(uuid.NewV4()),
			Name:      "TestSuite",
			Error:     fmt.Errorf("test error"),
			StartTime: time.Now(),
			EndTime:   time.Now().Add(1 * time.Second),
		}
		assert.Equal(t, "FAIL", result.Status())
	})
}

func TestResult_Duration(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(2 * time.Second)
	result := Result{
		ID:        uuid.Must(uuid.NewV4()),
		Name:      "TestSuite",
		Error:     nil,
		StartTime: startTime,
		EndTime:   endTime,
	}
	assert.Equal(t, 2*time.Second, result.Duration())
}
