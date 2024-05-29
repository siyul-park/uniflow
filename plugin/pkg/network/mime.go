package network

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/object"
)

const (
	ApplicationJSON                  = "application/json"
	ApplicationJSONCharsetUTF8       = ApplicationJSON + "; " + charsetUTF8
	ApplicationJavaScript            = "application/javascript"
	ApplicationJavaScriptCharsetUTF8 = ApplicationJavaScript + "; " + charsetUTF8
	ApplicationXML                   = "application/xml"
	ApplicationXMLCharsetUTF8        = ApplicationXML + "; " + charsetUTF8
	ApplicationOctetStream           = "application/octet-stream"
	TextXML                          = "text/xml"
	TextXMLCharsetUTF8               = TextXML + "; " + charsetUTF8
	ApplicationForm                  = "application/x-www-form-urlencoded"
	ApplicationProtobuf              = "application/protobuf"
	ApplicationMsgpack               = "application/msgpack"
	TextHTML                         = "text/html"
	TextHTMLCharsetUTF8              = TextHTML + "; " + charsetUTF8
	TextPlain                        = "text/plain"
	TextPlainCharsetUTF8             = TextPlain + "; " + charsetUTF8
	MultipartFormData                = "multipart/form-data"
)

const charsetUTF8 = "charset=utf-8"

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

// IsCompatibleMIMEType checks if two media types are compatible.
func IsCompatibleMIMEType(x, y string) bool {
	if x == "*" || y == "*" || x == y {
		return true
	}

	tokensX := strings.Split(x, "/")
	tokensY := strings.Split(y, "/")

	if len(tokensX) != len(tokensY) {
		return false
	}

	for i := 0; i < len(tokensX); i++ {
		tokenX := tokensX[i]
		tokenY := tokensY[i]

		if tokenX != tokenY && tokenX != "*" && tokenY != "*" {
			return false
		}
	}

	return true
}

// MarshalMIME converts a object.Value to MIME data.
func MarshalMIME(value object.Object, contentType *string) ([]byte, error) {
	if contentType == nil {
		contentType = lo.ToPtr[string]("")
	}

	if value == nil {
		return nil, nil
	}

	var data []byte
	if v, ok := value.(object.String); ok {
		data = []byte(v.String())
	} else if v, ok := value.(object.Binary); ok {
		data = v.Bytes()
	}
	if data != nil {
		if *contentType == "" {
			*contentType = http.DetectContentType(data)
		}
		return data, nil
	}

	if *contentType == "" {
		*contentType = ApplicationJSONCharsetUTF8
	}

	mediaType, params, err := mime.ParseMediaType(*contentType)
	if err != nil {
		return nil, errors.WithStack(encoding.ErrInvalidValue)
	}

	switch mediaType {
	case ApplicationJSON:
		return json.Marshal(value.Interface())
	case ApplicationForm:
		urlValues := url.Values{}
		if err := object.Unmarshal(value, &urlValues); err != nil {
			return nil, err
		}
		return []byte(urlValues.Encode()), nil
	case TextPlain:
		return []byte(fmt.Sprintf("%v", value.Interface())), nil
	case MultipartFormData:
		boundary, ok := params["boundary"]
		if !ok {
			boundary = randomMultiPartBoundary()
			params["boundary"] = boundary
			*contentType = mime.FormatMediaType(mediaType, params)
		}

		bodyBuffer := new(bytes.Buffer)
		mw := multipart.NewWriter(bodyBuffer)
		if err := mw.SetBoundary(boundary); err != nil {
			return nil, err
		}

		writeField := func(obj *object.Map, key object.Object) error {
			if key, ok := key.(object.String); ok {
				value := obj.GetOr(key, nil)

				var elements *object.Slice
				if v, ok := value.(*object.Slice); ok {
					elements = v
				} else {
					elements = object.NewSlice(value)
				}

				for _, element := range elements.Values() {
					contentType := ""
					b, err := MarshalMIME(element, &contentType)
					if err != nil {
						return err
					}

					h := textproto.MIMEHeader{}
					h.Set(HeaderContentDisposition, fmt.Sprintf(`form-data; name="%s"`, quoteEscaper.Replace(key.String())))
					if contentType != "" && contentType != TextPlainCharsetUTF8 {
						h.Set(HeaderContentType, contentType)
					}

					if writer, err := mw.CreatePart(h); err != nil {
						return err
					} else if _, err := writer.Write(b); err != nil {
						return err
					}
				}
			}
			return nil
		}
		writeFields := func(value object.Object) error {
			if value, ok := value.(*object.Map); ok {
				for _, key := range value.Keys() {
					if err := writeField(value, key); err != nil {
						return err
					}
				}
			}
			return nil
		}
		writeFiles := func(value object.Object) error {
			if value, ok := value.(*object.Map); ok {
				for _, key := range value.Keys() {
					if key, ok := key.(object.String); ok {
						value := value.GetOr(key, nil)

						var elements *object.Slice
						if v, ok := value.(*object.Slice); ok {
							elements = v
						} else {
							elements = object.NewSlice(value)
						}

						for _, element := range elements.Values() {
							data, ok := object.Pick[object.Object](element, "data")
							if !ok {
								data = element
							}
							filename, ok := object.Pick[string](element, "filename")
							if !ok {
								filename = key.String()
							}

							header, _ := object.Pick[object.Object](element, "header")

							contentType := ""
							contentTypes, _ := object.Pick[object.Object](header, HeaderContentType)
							if contentTypes != nil {
								if c, ok := contentTypes.(*object.Slice); ok {
									contentType, _ = object.Pick[string](c, "0")
								} else if c, ok := contentTypes.(object.String); ok {
									contentType = c.String()
								}
							}

							contentEncoding := ""
							contentEncodings, _ := object.Pick[object.Object](header, HeaderContentEncoding)
							if contentEncodings != nil {
								if c, ok := contentEncodings.(*object.Slice); ok {
									contentEncoding, _ = object.Pick[string](c, "0")
								} else if c, ok := contentEncodings.(object.String); ok {
									contentEncoding = c.String()
								}
							}

							b, err := MarshalMIME(data, &contentType)
							if err != nil {
								return err
							}
							b, err = Compress(b, contentEncoding)
							if err != nil {
								return err
							}

							h := textproto.MIMEHeader{}
							_ = object.Unmarshal(header, &h)

							h.Set(HeaderContentDisposition, fmt.Sprintf(`form-data; name="%s"; filename="%s"`, quoteEscaper.Replace(key.String()), quoteEscaper.Replace(filename)))
							h.Set(HeaderContentType, contentType)

							if writer, err := mw.CreatePart(h); err != nil {
								return err
							} else if _, err := writer.Write(b); err != nil {
								return err
							}
						}
					}
				}
			}
			return nil
		}

		if v, ok := value.(*object.Map); ok {
			for _, key := range v.Keys() {
				value := v.GetOr(key, nil)

				if key == object.NewString("value") {
					if err := writeFields(value); err != nil {
						return nil, err
					}
				} else if key == object.NewString("file") {
					if err := writeFiles(value); err != nil {
						return nil, err
					}
				} else if err := writeField(v, key); err != nil {
					return nil, err
				}
			}
		}

		if err := mw.Close(); err != nil {
			return nil, err
		}
		return bodyBuffer.Bytes(), nil
	}

	return nil, errors.WithStack(encoding.ErrUnsupportedValue)
}

// UnmarshalMIME converts MIME data into a object.Value.
func UnmarshalMIME(data []byte, contentType *string) (object.Object, error) {
	if len(data) == 0 {
		return nil, nil
	}

	if contentType == nil {
		contentType = lo.ToPtr[string]("")
	}
	if *contentType == "" {
		var rawJSON json.RawMessage
		if json.Unmarshal(data, &rawJSON) == nil {
			*contentType = ApplicationJSONCharsetUTF8
		} else {
			*contentType = http.DetectContentType(data)
		}
	}

	mediaType, params, err := mime.ParseMediaType(*contentType)
	if err != nil {
		return nil, err
	}

	switch mediaType {
	case ApplicationJSON:
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return object.MarshalText(v)
	case ApplicationForm:
		v, err := url.ParseQuery(string(data))
		if err != nil {
			return nil, err
		}
		return object.MarshalText(v)
	case TextPlain:
		return object.NewString(string(data)), nil
	case MultipartFormData:
		reader := multipart.NewReader(bytes.NewReader(data), params["boundary"])
		form, err := reader.ReadForm(int64(len(data)))
		if err != nil {
			return nil, err
		}
		defer form.RemoveAll()

		files := map[string][]map[string]any{}
		for name, fhs := range form.File {
			for _, fh := range fhs {
				file, err := fh.Open()
				if err != nil {
					return nil, err
				}
				b, err := io.ReadAll(file)
				if err != nil {
					return nil, err
				}

				contentType := fh.Header.Get(HeaderContentType)
				contentEncoding := fh.Header.Get(HeaderContentEncoding)

				b, err = Decompress(b, contentEncoding)
				if err != nil {
					return nil, err
				}
				data, err := UnmarshalMIME(b, &contentType)
				if err != nil {
					return nil, err
				}

				fh.Header.Set(HeaderContentType, contentType)

				files[name] = append(files[name], map[string]any{
					"filename": fh.Filename,
					"header":   fh.Header,
					"size":     fh.Size,
					"data":     data,
				})
			}
		}

		return object.MarshalText(map[string]any{
			"value": form.Value,
			"file":  files,
		})
	case ApplicationOctetStream:
		return object.NewBinary(data), nil
	default:
		return object.NewBinary(data), nil
	}
}

func randomMultiPartBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}
