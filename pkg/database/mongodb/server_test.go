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

func BenchmarkServerAndRelease(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			server := Server()
			ReleaseServer(server)
		}
	})
}
