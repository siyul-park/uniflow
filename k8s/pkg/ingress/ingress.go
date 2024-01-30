package ingress

import (
	"net/http"

	"k8s.io/client-go/kubernetes"
)

type Ingress struct {
	k8s kubernetes.Interface
}

var _ http.Handler = (*Ingress)(nil)

func New(k8s kubernetes.Interface) *Ingress {
	return &Ingress{k8s: k8s}
}

// ServeHTTP implements http.Handler.
func (*Ingress) ServeHTTP(http.ResponseWriter, *http.Request) {
	panic("unimplemented")
}
