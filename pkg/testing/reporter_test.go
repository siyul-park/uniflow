package testing

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestReporters_Report(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	var reporters Reporters
	reporters = append(reporters, ReportFunc(func(_ context.Context, _ *Result) error {
		return nil
	}))
	result := &Result{Name: "foo", StartTime: time.Now(), EndTime: time.Now()}

	err := reporters.Report(ctx, result)
	assert.NoError(t, err)
}

func TestErrorReporter_Report(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	reporter := NewErrorReporter()
	result := &Result{Name: "foo", Error: errors.New(faker.Sentence()), StartTime: time.Now(), EndTime: time.Now()}

	err := reporter.Report(ctx, result)
	assert.NoError(t, err)

	err = reporter.Error()
	assert.Error(t, err)
}

func TestTextReporter_Report(t *testing.T) {
	t.Run(StatusPass, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		output := &bytes.Buffer{}
		reporter := NewTextReporter(output)
		result := &Result{Name: "foo", StartTime: time.Now(), EndTime: time.Now()}

		err := reporter.Report(ctx, result)
		assert.NoError(t, err)
		assert.Contains(t, output.String(), "PASS\tfoo")
	})

	t.Run(StatusFail, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		output := &bytes.Buffer{}
		reporter := NewTextReporter(output)
		result := &Result{Name: "foo", Error: fmt.Errorf("error"), StartTime: time.Now(), EndTime: time.Now()}

		err := reporter.Report(ctx, result)
		assert.NoError(t, err)
		assert.Contains(t, output.String(), "FAIL\tfoo")
	})
}
