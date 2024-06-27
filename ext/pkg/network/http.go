package network

import (
	"net/http"
	"net/url"

	"github.com/siyul-park/uniflow/pkg/object"
)

// HTTPPayload is the payload structure for HTTP requests and responses.
type HTTPPayload struct {
	Method string        `map:"method,omitempty"`
	Scheme string        `map:"scheme,omitempty"`
	Host   string        `map:"host,omitempty"`
	Path   string        `map:"path,omitempty"`
	Query  url.Values    `map:"query,omitempty"`
	Proto  string        `map:"proto,omitempty"`
	Header http.Header   `map:"header,omitempty"`
	Body   object.Object `map:"body,omitempty"`
	Status int           `map:"status"`
}

const KeyHTTPRequest = "http.Request"
const KeyHTTPResponseWriter = "http.ResponseWriter"

// NewHTTPPayload creates a new HTTPPayload with the given HTTP status code and optional body.
func NewHTTPPayload(status int, body ...object.Object) *HTTPPayload {
	if len(body) == 0 {
		body = []object.Object{object.NewString(http.StatusText(status))}
	}
	return &HTTPPayload{
		Header: http.Header{},
		Body:   body[0],
		Status: status,
	}
}
