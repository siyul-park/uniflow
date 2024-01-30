package ingress

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewOperator(t *testing.T) {
	k8s := fake.NewSimpleClientset()
	o := NewOperator(k8s)
	assert.NotNil(t, o)
}
