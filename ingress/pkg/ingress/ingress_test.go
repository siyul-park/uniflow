package ingress

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewIngress(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	i := New(k8s)
	assert.NotNil(t, i)
}
