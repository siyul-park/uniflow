package main

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	err := execute(ctx, "memdb://", faker.UUIDHyphenated())
	assert.NoError(t, err)
}
