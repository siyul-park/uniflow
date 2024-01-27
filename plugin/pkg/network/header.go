package network

import "regexp"

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

var blacklistResponseHeaders []*regexp.Regexp

func init() {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers
	blacklistResponseHeaderPatterns := []string{
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

	for _, pattern := range blacklistResponseHeaderPatterns {
		blacklistResponseHeaders = append(blacklistResponseHeaders, regexp.MustCompile(pattern))
	}
}

func IsResponseHeader(header string) bool {
	h := []byte(header)
	for _, header := range blacklistResponseHeaders {
		if header.Match(h) {
			return false
		}
	}
	return true
}
