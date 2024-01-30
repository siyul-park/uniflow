package ingress

import (
	"net/http"

	"k8s.io/client-go/kubernetes"
)

type Operator struct {
	k8s kubernetes.Interface
}

var _ http.Handler = (*Operator)(nil)

func NewOperator(k8s kubernetes.Interface) *Operator {
	return &Operator{k8s: k8s}
}

// ServeHTTP implements http.Handler.
func (*Operator) ServeHTTP(http.ResponseWriter, *http.Request) {
	panic("unimplemented")
}
