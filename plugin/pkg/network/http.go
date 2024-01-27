package network

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
)

type HTTPNode struct {
	server   *http.Server
	listener net.Listener
	ioPort   *port.Port
	inPort   *port.Port
	outPort  *port.Port
	errPort  *port.Port
	mu       sync.RWMutex
}

type HTTPPayload struct {
	Proto   string          `map:"proto,omitempty"`
	Path    string          `map:"path,omitempty"`
	Method  string          `map:"method,omitempty"`
	Header  http.Header     `map:"header,omitempty"`
	Query   url.Values      `map:"query,omitempty"`
	Cookies []*http.Cookie  `map:"cookies,omitempty"`
	Body    primitive.Value `map:"body,omitempty"`
	Status  int             `map:"status"`
}

var (
	PayloadBadRequest                    = NewHTTPPayload(http.StatusBadRequest)                    // HTTP 400 Bad Request
	PayloadUnauthorized                  = NewHTTPPayload(http.StatusUnauthorized)                  // HTTP 401 Unauthorized
	PayloadPaymentRequired               = NewHTTPPayload(http.StatusPaymentRequired)               // HTTP 402 Payment Required
	PayloadForbidden                     = NewHTTPPayload(http.StatusForbidden)                     // HTTP 403 Forbidden
	PayloadNotFound                      = NewHTTPPayload(http.StatusNotFound)                      // HTTP 404 Not Found
	PayloadMethodNotAllowed              = NewHTTPPayload(http.StatusMethodNotAllowed)              // HTTP 405 Method Not Allowed
	PayloadNotAcceptable                 = NewHTTPPayload(http.StatusNotAcceptable)                 // HTTP 406 Not Acceptable
	PayloadProxyAuthRequired             = NewHTTPPayload(http.StatusProxyAuthRequired)             // HTTP 407 Proxy AuthRequired
	PayloadRequestTimeout                = NewHTTPPayload(http.StatusRequestTimeout)                // HTTP 408 Request Timeout
	PayloadConflict                      = NewHTTPPayload(http.StatusConflict)                      // HTTP 409 Conflict
	PayloadGone                          = NewHTTPPayload(http.StatusGone)                          // HTTP 410 Gone
	PayloadLengthRequired                = NewHTTPPayload(http.StatusLengthRequired)                // HTTP 411 Length Required
	PayloadPreconditionFailed            = NewHTTPPayload(http.StatusPreconditionFailed)            // HTTP 412 Precondition Failed
	PayloadRequestEntityTooLarge         = NewHTTPPayload(http.StatusRequestEntityTooLarge)         // HTTP 413 Payload Too Large
	PayloadRequestURITooLong             = NewHTTPPayload(http.StatusRequestURITooLong)             // HTTP 414 URI Too Long
	PayloadUnsupportedMediaType          = NewHTTPPayload(http.StatusUnsupportedMediaType)          // HTTP 415 Unsupported Media Type
	PayloadRequestedRangeNotSatisfiable  = NewHTTPPayload(http.StatusRequestedRangeNotSatisfiable)  // HTTP 416 Range Not Satisfiable
	PayloadExpectationFailed             = NewHTTPPayload(http.StatusExpectationFailed)             // HTTP 417 Expectation Failed
	PayloadTeapot                        = NewHTTPPayload(http.StatusTeapot)                        // HTTP 418 I'm a teapot
	PayloadMisdirectedRequest            = NewHTTPPayload(http.StatusMisdirectedRequest)            // HTTP 421 Misdirected Request
	PayloadUnprocessableEntity           = NewHTTPPayload(http.StatusUnprocessableEntity)           // HTTP 422 Unprocessable Entity
	PayloadLocked                        = NewHTTPPayload(http.StatusLocked)                        // HTTP 423 Locked
	PayloadFailedDependency              = NewHTTPPayload(http.StatusFailedDependency)              // HTTP 424 Failed Dependency
	PayloadTooEarly                      = NewHTTPPayload(http.StatusTooEarly)                      // HTTP 425 Too Early
	PayloadUpgradeRequired               = NewHTTPPayload(http.StatusUpgradeRequired)               // HTTP 426 Upgrade Required
	PayloadPreconditionRequired          = NewHTTPPayload(http.StatusPreconditionRequired)          // HTTP 428 Precondition Required
	PayloadTooManyRequests               = NewHTTPPayload(http.StatusTooManyRequests)               // HTTP 429 Too Many Requests
	PayloadRequestHeaderFieldsTooLarge   = NewHTTPPayload(http.StatusRequestHeaderFieldsTooLarge)   // HTTP 431 Request Header Fields Too Large
	PayloadUnavailableForLegalReasons    = NewHTTPPayload(http.StatusUnavailableForLegalReasons)    // HTTP 451 Unavailable For Legal Reasons
	PayloadInternalServerError           = NewHTTPPayload(http.StatusInternalServerError)           // HTTP 500 Internal Server Error
	PayloadNotImplemented                = NewHTTPPayload(http.StatusNotImplemented)                // HTTP 501 Not Implemented
	PayloadBadGateway                    = NewHTTPPayload(http.StatusBadGateway)                    // HTTP 502 Bad Gateway
	PayloadServiceUnavailable            = NewHTTPPayload(http.StatusServiceUnavailable)            // HTTP 503 Service Unavailable
	PayloadGatewayTimeout                = NewHTTPPayload(http.StatusGatewayTimeout)                // HTTP 504 Gateway Timeout
	PayloadHTTPVersionNotSupported       = NewHTTPPayload(http.StatusHTTPVersionNotSupported)       // HTTP 505 HTTP Version Not Supported
	PayloadVariantAlsoNegotiates         = NewHTTPPayload(http.StatusVariantAlsoNegotiates)         // HTTP 506 Variant Also Negotiates
	PayloadInsufficientStorage           = NewHTTPPayload(http.StatusInsufficientStorage)           // HTTP 507 Insufficient Storage
	PayloadLoopDetected                  = NewHTTPPayload(http.StatusLoopDetected)                  // HTTP 508 Loop Detected
	PayloadNotExtended                   = NewHTTPPayload(http.StatusNotExtended)                   // HTTP 510 Not Extended
	PayloadNetworkAuthenticationRequired = NewHTTPPayload(http.StatusNetworkAuthenticationRequired) // HTTP 511 Network Authentication Required
)

var _ node.Node = (*HTTPNode)(nil)
var _ http.Handler = (*HTTPNode)(nil)

func NewHTTPPayload(status int, body ...primitive.Value) HTTPPayload {
	if len(body) == 0 {
		body = []primitive.Value{primitive.String(http.StatusText(status))}
	}
	return HTTPPayload{
		Body:   body[0],
		Status: status,
	}
}

func NewHTTPNode(address string) *HTTPNode {
	n := &HTTPNode{
		ioPort:  port.New(),
		inPort:  port.New(),
		outPort: port.New(),
		errPort: port.New(),
	}

	s := new(http.Server)
	s.Addr = address
	s.Handler = n
	n.server = s

	return n
}

func (n *HTTPNode) Port(name string) (*port.Port, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIO:
		return n.ioPort, true
	case node.PortIn:
		return n.inPort, true
	case node.PortOut:
		return n.outPort, true
	case node.PortErr:
		return n.errPort, true
	default:
	}

	return nil, false
}

func (n *HTTPNode) Address() net.Addr {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.listener == nil {
		return nil
	}
	return n.listener.Addr()
}

func (n *HTTPNode) WaitForListen(errChan <-chan error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if addr := n.Address(); addr != nil {
				return nil
			}
		case err := <-errChan:
			if err == http.ErrServerClosed {
				return nil
			}
			return err
		}
	}
}

func (n *HTTPNode) Listen() error {
	if err := func() error {
		n.mu.Lock()
		defer n.mu.Unlock()

		if n.listener != nil {
			return nil
		}
		if l, err := newTCPKeepAliveListener(n.server.Addr, "tcp"); err != nil {
			return err
		} else {
			n.listener = l
		}
		return nil
	}(); err != nil {
		return err
	}

	return n.server.Serve(n.listener)
}

func (n *HTTPNode) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	proc := process.New()
	defer proc.Exit(nil)

	if err := n.action(proc, w, r); err != nil {
		errPayload := n.newErrorPayload(proc, err)
		n.write(w, errPayload)
		proc.Exit(err)
	}
}

func (n *HTTPNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()

	return n.server.Close()
}

func (n *HTTPNode) action(proc *process.Process, w http.ResponseWriter, r *http.Request) error {
	ioStream := n.ioPort.Open(proc)
	inStream := n.inPort.Open(proc)
	outStream := n.outPort.Open(proc)

	if ioStream.Links()+outStream.Links() == 0 {
		return nil
	}

	req, err := n.read(r)
	if err != nil {
		return err
	}

	outPayload, err := primitive.MarshalBinary(req)
	if err != nil {
		return err
	}
	outPck := packet.New(outPayload)

	ioStream.Send(outPck)
	outStream.Send(outPck)

	if ioStream.Links()+inStream.Links() == 0 {
		return nil
	}

	var inPck *packet.Packet
	var ok bool
	select {
	case inPck, ok = <-ioStream.Receive():
	case inPck, ok = <-inStream.Receive():
	}
	if !ok {
		return n.write(w, PayloadServiceUnavailable)
	}

	if err, ok := packet.AsError(inPck); ok {
		return err
	}

	var res HTTPPayload
	inPayload := inPck.Payload()
	if err := primitive.Unmarshal(inPayload, &res); err != nil {
		res.Body = inPayload
	}

	return n.write(w, res)
}

func (n *HTTPNode) newErrorPayload(proc *process.Process, err error) HTTPPayload {
	errStream := n.errPort.Open(proc)
	if errStream.Links() == 0 {
		return PayloadInternalServerError
	}

	errPck := packet.WithError(err, nil)
	errStream.Send(errPck)

	inPck, ok := <-errStream.Receive()
	if !ok {
		return PayloadInternalServerError
	}

	if _, ok := packet.AsError(inPck); ok {
		return PayloadInternalServerError
	}

	var res HTTPPayload
	inPayload := inPck.Payload()
	if err := primitive.Unmarshal(inPayload, &res); err != nil {
		res.Body = inPayload
	}
	return res
}

func (n *HTTPNode) read(r *http.Request) (HTTPPayload, error) {
	contentType := r.Header.Get(HeaderContentType)

	if b, err := io.ReadAll(r.Body); err != nil {
		return HTTPPayload{}, err
	} else if b, err := UnmarshalMIME(b, &contentType); err != nil {
		return HTTPPayload{}, err
	} else {
		r.Header.Set(HeaderContentType, contentType)
		return HTTPPayload{
			Proto:   r.Proto,
			Path:    r.URL.Path,
			Method:  r.Method,
			Header:  r.Header,
			Query:   r.URL.Query(),
			Cookies: r.Cookies(),
			Body:    b,
		}, nil
	}
}

func (n *HTTPNode) write(w http.ResponseWriter, res HTTPPayload) error {
	contentType := res.Header.Get(HeaderContentType)

	b, err := MarshalMIME(res.Body, &contentType)
	if err != nil {
		return err
	}

	if res.Header == nil {
		res.Header = http.Header{}
	}
	res.Header.Set(HeaderContentType, contentType)
	for key := range w.Header() {
		w.Header().Del(key)
	}
	for key, headers := range res.Header {
		if !IsResponseHeader(key) {
			continue
		}
		for _, header := range headers {
			w.Header().Add(key, header)
		}
	}
	w.Header().Set(HeaderContentLength, strconv.Itoa(len(b)))
	w.Header().Set(HeaderContentType, contentType)

	status := res.Status
	if status == 0 {
		if len(b) == 0 {
			status = http.StatusNoContent
		} else {
			status = http.StatusOK
		}
	}
	w.WriteHeader(status)

	if _, err := w.Write(b); err != nil {
		return err
	}
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}
