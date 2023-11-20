package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerAndRelease(t *testing.T) {
	server := Server()
	assert.NotNil(t, server)

	ReleaseServer(server)
}
