package networkx

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/internal/util"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

type (
	HTTPNodeConfig struct {
		ID      ulid.ULID
		Address string
	}
	HTTPNode struct {
		id              ulid.ULID
		address         string
		server          *http.Server
		listener        net.Listener
		listenerNetwork string
		ioPort          *port.Port
		inPort          *port.Port
		outPort         *port.Port
		errPort         *port.Port
		mu              sync.RWMutex
	}

	HTTPPayload struct {
		Proto   string           `map:"proto,omitempty"`
		Path    string           `map:"path,omitempty"`
		Method  string           `map:"method,omitempty"`
		Header  http.Header      `map:"header,omitempty"`
		Query   url.Values       `map:"query,omitempty"`
		Cookies []*http.Cookie   `map:"cookies,omitempty"`
		Body    primitive.Object `map:"body,omitempty"`
		Status  int              `map:"status"`
	}

	HTTPSpec struct {
		scheme.SpecMeta `map:",inline"`
		Address         string `map:"address"`
	}

	tcpKeepAliveListener struct {
		*net.TCPListener
	}
)

const (
	KindHTTP = "http"
)

var _ node.Node = &HTTPNode{}
var _ http.Handler = &HTTPNode{}

const (
	HeaderAccept                  = "Accept"
	HeaderAcceptCharset           = "Accept-Charset"
	HeaderAcceptEncoding          = "Accept-Encoding"
	HeaderAcceptLanguage          = "Accept-Language"
	HeaderAllow                   = "Allow"
	HeaderAuthorization           = "Authorization"
	HeaderContentDisposition      = "Content-Disposition"
	HeaderContentEncoding         = "Content-Encoding"
	HeaderContentLength           = "Content-Length"
	HeaderContentType             = "Content-Type"
	HeaderCookie                  = "Cookie"
	HeaderSetCookie               = "Set-Cookie"
	HeaderIfModifiedSince         = "If-Modified-Since"
	HeaderLastModified            = "Last-Modified"
	HeaderLocation                = "Location"
	HeaderRetryAfter              = "Retry-After"
	HeaderUpgrade                 = "Upgrade"
	HeaderUpgradeInsecureRequests = "Upgrade-Insecure-Requests"
	HeaderVary                    = "Vary"
	HeaderWWWAuthenticate         = "WWW-Authenticate"
	HeaderForwarded               = "Forwarded"
	HeaderXForwardedFor           = "X-Forwarded-For"
	HeaderXForwardedHost          = "X-Forwarded-Host"
	HeaderXForwardedProto         = "X-Forwarded-Proto"
	HeaderXForwardedProtocol      = "X-Forwarded-Protocol"
	HeaderXForwardedSsl           = "X-Forwarded-Ssl"
	HeaderXUrlScheme              = "X-Url-Scheme"
	HeaderXHTTPMethodOverride     = "X-HTTP-Method-Override"
	HeaderXRealIP                 = "X-Real-Ip"
	HeaderXRequestID              = "X-Request-Id"
	HeaderXCorrelationID          = "X-Correlation-Id"
	HeaderXRequestedWith          = "X-Requested-With"
	HeaderServer                  = "Server"
	HeaderOrigin                  = "Origin"
	HeaderCacheControl            = "Cache-Control"
	HeaderConnection              = "Connection"
	HeaderDate                    = "Date"
	HeaderDeviceMemory            = "Device-Memory"
	HeaderDNT                     = "DNT"
	HeaderDownlink                = "Downlink"
	HeaderDPR                     = "DPR"
	HeaderEarlyData               = "Early-Data"
	HeaderECT                     = "ECT"
	HeaderExpect                  = "Expect"
	HeaderExpectCT                = "Expect-CT"
	HeaderFrom                    = "From"
	HeaderHost                    = "Host"
	HeaderIfMatch                 = "If-Match"
	HeaderIfNoneMatch             = "If-None-Match"
	HeaderIfRange                 = "If-Range"
	HeaderIfUnmodifiedSince       = "If-Unmodified-Since"
	HeaderKeepAlive               = "Keep-Alive"
	HeaderMaxForwards             = "Max-Forwards"
	HeaderProxyAuthorization      = "Proxy-Authorization"
	HeaderRange                   = "Range"
	HeaderReferer                 = "Referer"
	HeaderRTT                     = "RTT"
	HeaderSaveData                = "Save-Data"
	HeaderTE                      = "TE"
	HeaderTk                      = "Tk"
	HeaderTrailer                 = "Trailer"
	HeaderTransferEncoding        = "Transfer-Encoding"
	HeaderUserAgent               = "User-Agent"
	HeaderVia                     = "Via"
	HeaderViewportWidth           = "Viewport-Width"
	HeaderWantDigest              = "Want-Digest"
	HeaderWarning                 = "Warning"
	HeaderWidth                   = "Width"

	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	HeaderStrictTransportSecurity         = "Strict-Transport-Security"
	HeaderXContentTypeOptions             = "X-Content-Type-Options"
	HeaderXXSSProtection                  = "X-XSS-Protection"
	HeaderXFrameOptions                   = "X-Frame-Options"
	HeaderContentSecurityPolicy           = "Content-Security-Policy"
	HeaderContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only"
	HeaderXCSRFToken                      = "X-CSRF-Token"
	HeaderReferrerPolicy                  = "Referrer-Policy"
)

var (
	BadRequest                    = NewHTTPPayload(http.StatusBadRequest)                    // HTTP 400 Bad Request
	Unauthorized                  = NewHTTPPayload(http.StatusUnauthorized)                  // HTTP 401 Unauthorized
	PaymentRequired               = NewHTTPPayload(http.StatusPaymentRequired)               // HTTP 402 Payment Required
	Forbidden                     = NewHTTPPayload(http.StatusForbidden)                     // HTTP 403 Forbidden
	NotFound                      = NewHTTPPayload(http.StatusNotFound)                      // HTTP 404 Not Found
	MethodNotAllowed              = NewHTTPPayload(http.StatusMethodNotAllowed)              // HTTP 405 Method Not Allowed
	NotAcceptable                 = NewHTTPPayload(http.StatusNotAcceptable)                 // HTTP 406 Not Acceptable
	ProxyAuthRequired             = NewHTTPPayload(http.StatusProxyAuthRequired)             // HTTP 407 Proxy AuthRequired
	RequestTimeout                = NewHTTPPayload(http.StatusRequestTimeout)                // HTTP 408 Request Timeout
	Conflict                      = NewHTTPPayload(http.StatusConflict)                      // HTTP 409 Conflict
	Gone                          = NewHTTPPayload(http.StatusGone)                          // HTTP 410 Gone
	LengthRequired                = NewHTTPPayload(http.StatusLengthRequired)                // HTTP 411 Length Required
	PreconditionFailed            = NewHTTPPayload(http.StatusPreconditionFailed)            // HTTP 412 Precondition Failed
	StatusRequestEntityTooLarge   = NewHTTPPayload(http.StatusRequestEntityTooLarge)         // HTTP 413 Payload Too Large
	RequestURITooLong             = NewHTTPPayload(http.StatusRequestURITooLong)             // HTTP 414 URI Too Long
	UnsupportedMediaType          = NewHTTPPayload(http.StatusUnsupportedMediaType)          // HTTP 415 Unsupported Media Type
	RequestedRangeNotSatisfiable  = NewHTTPPayload(http.StatusRequestedRangeNotSatisfiable)  // HTTP 416 Range Not Satisfiable
	ExpectationFailed             = NewHTTPPayload(http.StatusExpectationFailed)             // HTTP 417 Expectation Failed
	Teapot                        = NewHTTPPayload(http.StatusTeapot)                        // HTTP 418 I'm a teapot
	MisdirectedRequest            = NewHTTPPayload(http.StatusMisdirectedRequest)            // HTTP 421 Misdirected Request
	UnprocessableEntity           = NewHTTPPayload(http.StatusUnprocessableEntity)           // HTTP 422 Unprocessable Entity
	Locked                        = NewHTTPPayload(http.StatusLocked)                        // HTTP 423 Locked
	FailedDependency              = NewHTTPPayload(http.StatusFailedDependency)              // HTTP 424 Failed Dependency
	TooEarly                      = NewHTTPPayload(http.StatusTooEarly)                      // HTTP 425 Too Early
	UpgradeRequired               = NewHTTPPayload(http.StatusUpgradeRequired)               // HTTP 426 Upgrade Required
	PreconditionRequired          = NewHTTPPayload(http.StatusPreconditionRequired)          // HTTP 428 Precondition Required
	TooManyRequests               = NewHTTPPayload(http.StatusTooManyRequests)               // HTTP 429 Too Many Requests
	RequestHeaderFieldsTooLarge   = NewHTTPPayload(http.StatusRequestHeaderFieldsTooLarge)   // HTTP 431 Request Header Fields Too Large
	UnavailableForLegalReasons    = NewHTTPPayload(http.StatusUnavailableForLegalReasons)    // HTTP 451 Unavailable For Legal Reasons
	InternalServerError           = NewHTTPPayload(http.StatusInternalServerError)           // HTTP 500 Internal Server Error
	NotImplemented                = NewHTTPPayload(http.StatusNotImplemented)                // HTTP 501 Not Implemented
	BadGateway                    = NewHTTPPayload(http.StatusBadGateway)                    // HTTP 502 Bad Gateway
	ServiceUnavailable            = NewHTTPPayload(http.StatusServiceUnavailable)            // HTTP 503 Service Unavailable
	GatewayTimeout                = NewHTTPPayload(http.StatusGatewayTimeout)                // HTTP 504 Gateway Timeout
	HTTPVersionNotSupported       = NewHTTPPayload(http.StatusHTTPVersionNotSupported)       // HTTP 505 HTTP Version Not Supported
	VariantAlsoNegotiates         = NewHTTPPayload(http.StatusVariantAlsoNegotiates)         // HTTP 506 Variant Also Negotiates
	InsufficientStorage           = NewHTTPPayload(http.StatusInsufficientStorage)           // HTTP 507 Insufficient Storage
	LoopDetected                  = NewHTTPPayload(http.StatusLoopDetected)                  // HTTP 508 Loop Detected
	NotExtended                   = NewHTTPPayload(http.StatusNotExtended)                   // HTTP 510 Not Extended
	NetworkAuthenticationRequired = NewHTTPPayload(http.StatusNetworkAuthenticationRequired) // HTTP 511 Network Authentication Required
)

var (
	ErrInvalidListenerNetwork = errors.New("invalid listener network")
)

var (
	forbiddenResponseHeaderRegexps []*regexp.Regexp
)

func init() {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers
	forbiddenResponseHeaderPatterns := []string{
		HeaderAccept, HeaderAcceptCharset, HeaderAcceptEncoding, HeaderAcceptLanguage,
		HeaderAuthorization,
		HeaderConnection,
		HeaderCookie,
		HeaderDate,
		HeaderDeviceMemory,
		HeaderDNT,
		HeaderDownlink,
		HeaderDPR,
		HeaderEarlyData,
		HeaderECT,
		HeaderExpect, HeaderExpectCT,
		HeaderForwarded,
		HeaderXForwardedFor, HeaderXForwardedHost, HeaderXForwardedProto, HeaderXForwardedProtocol,
		HeaderFrom,
		HeaderHost,
		HeaderIfMatch, HeaderIfModifiedSince, HeaderIfNoneMatch, HeaderIfRange, HeaderIfUnmodifiedSince,
		HeaderKeepAlive,
		HeaderMaxForwards,
		HeaderOrigin,
		HeaderProxyAuthorization,
		HeaderRange,
		HeaderReferer,
		HeaderRTT,
		HeaderSaveData,
		"Sec-.*",
		HeaderTE,
		HeaderTk,
		HeaderTrailer, HeaderTransferEncoding,
		HeaderUpgrade, HeaderUpgradeInsecureRequests,
		HeaderUserAgent,
		HeaderVia,
		HeaderViewportWidth,
		HeaderWantDigest,
		HeaderWarning,
		HeaderWidth,
	}

	for _, pattern := range forbiddenResponseHeaderPatterns {
		forbiddenResponseHeaderRegexps = append(forbiddenResponseHeaderRegexps, regexp.MustCompile(pattern))
	}
}

func NewHTTPNode(config HTTPNodeConfig) *HTTPNode {
	id := config.ID
	address := config.Address

	if util.IsZero(id) {
		id = ulid.Make()
	}

	n := &HTTPNode{
		id:              id,
		address:         address,
		server:          new(http.Server),
		listenerNetwork: "tcp",
		ioPort:          port.New(),
		inPort:          port.New(),
		outPort:         port.New(),
		errPort:         port.New(),
	}
	n.server.Handler = n

	return n
}

func (n *HTTPNode) ID() ulid.ULID {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.id
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

func (n *HTTPNode) ListenerAddr() net.Addr {
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
			addr := n.ListenerAddr()
			if addr != nil && strings.Contains(addr.String(), ":") {
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

func (n *HTTPNode) Start() error {
	n.mu.Lock()
	n.server.Addr = n.address
	if err := n.configureServer(); err != nil {
		n.mu.Unlock()
		return err
	}
	n.mu.Unlock()
	return n.server.Serve(n.listener)
}

func (n *HTTPNode) Close() error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if err := n.server.Close(); err != nil {
		return err
	}
	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()

	return nil
}

func (n *HTTPNode) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	proc := process.New()
	defer func() {
		proc.Stack().Wait()
		proc.Close()
	}()

	go func() {
		select {
		case <-r.Context().Done():
			proc.Close()
		case <-proc.Done():
		}
	}()

	ioStream := n.ioPort.Open(proc)
	inStream := n.inPort.Open(proc)
	outStream := n.outPort.Open(proc)

	req, err := n.request(r)
	if err != nil {
		_ = n.response(r, w, n.errorPayload(proc, UnsupportedMediaType))
		return
	}
	outPayload, err := primitive.MarshalText(req)
	if err != nil {
		_ = n.response(r, w, n.errorPayload(proc, BadRequest))
		return
	}
	outPck := packet.New(outPayload)

	if ioStream.Links() > 0 {
		ioStream.Send(outPck)
	}
	if outStream.Links() > 0 {
		outStream.Send(outPck)
	}
	if ioStream.Links()+outStream.Links() == 0 {
		return
	}

	var inPck *packet.Packet
	var ok bool

	select {
	case inPck, ok = <-inStream.Receive():
	case inPck, ok = <-ioStream.Receive():
	}
	if !ok {
		_ = n.response(r, w, n.errorPayload(proc, ServiceUnavailable))
		return
	}
	proc.Stack().Clear(inPck.ID())

	inPayload := inPck.Payload()

	var res HTTPPayload
	if err := primitive.Unmarshal(inPayload, &res); err != nil {
		res.Body = inPayload
	}

	if err := n.response(r, w, res); err != nil {
		_ = n.response(r, w, n.errorPayload(proc, InternalServerError))
	}
}

func (n *HTTPNode) request(r *http.Request) (HTTPPayload, error) {
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

func (n *HTTPNode) response(r *http.Request, w http.ResponseWriter, res HTTPPayload) error {
	if r.Method == http.MethodHead {
		res.Header.Del(HeaderContentType)
		res.Body = nil
		if res.Status == 200 {
			res.Status = 204
		}
	}

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
		if isForbiddenResponseHeader(key) {
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

func (n *HTTPNode) errorPayload(proc *process.Process, err HTTPPayload) HTTPPayload {
	if n.errPort.Links() == 0 {
		return err
	}

	errPayload, _ := primitive.MarshalText(err)
	errPck := packet.New(errPayload)
	errStream := n.errPort.Open(proc)
	errStream.Send(errPck)

	outPck, ok := <-errStream.Receive()
	if !ok {
		return err
	}

	var res HTTPPayload
	if err := primitive.Unmarshal(outPck.Payload(), &res); err != nil {
		_ = primitive.Unmarshal(outPck.Payload(), &res.Body)
	}
	return res
}

func (n *HTTPNode) configureServer() error {
	if n.listener == nil {
		l, err := newListener(n.server.Addr, n.listenerNetwork)
		if err != nil {
			return err
		}

		if n.server.TLSConfig != nil {
			n.listener = tls.NewListener(l, n.server.TLSConfig)
		} else {
			n.listener = l
		}
	}
	return nil
}

func NewHTTPPayload(status int, body ...primitive.Object) HTTPPayload {
	he := HTTPPayload{Status: status, Body: primitive.NewString(http.StatusText(status))}
	if len(body) > 0 {
		he.Body = body[0]
	}
	return he
}

func isForbiddenResponseHeader(header string) bool {
	h := []byte(header)
	forbidden := false
	for _, forbiddenHeader := range forbiddenResponseHeaderRegexps {
		if forbiddenHeader.Match(h) {
			forbidden = true
			break
		}
	}
	return forbidden
}

func newListener(address, network string) (*tcpKeepAliveListener, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, ErrInvalidListenerNetwork
	}
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return &tcpKeepAliveListener{l.(*net.TCPListener)}, nil
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	if c, err = ln.AcceptTCP(); err != nil {
		return
	} else if err = c.(*net.TCPConn).SetKeepAlive(true); err != nil {
		return
	}
	// Ignore error from setting the KeepAlivePeriod as some systems, such as
	// OpenBSD, do not support setting TCP_USER_TIMEOUT on IPPROTO_TCP
	_ = c.(*net.TCPConn).SetKeepAlivePeriod(3 * time.Minute)
	return
}
