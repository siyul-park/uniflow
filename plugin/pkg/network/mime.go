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
	"github.com/siyul-park/uniflow/pkg/primitive"
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

// MarshalMIME converts a primitive.Value to MIME data.
func MarshalMIME(value primitive.Value, contentType *string) ([]byte, error) {
	if contentType == nil {
		contentType = lo.ToPtr[string]("")
	}

	if value == nil {
		return nil, nil
	}

	var data []byte
	if v, ok := value.(primitive.String); ok {
		data = []byte(v.String())
	} else if v, ok := value.(primitive.Binary); ok {
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
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	}

	switch mediaType {
	case ApplicationJSON:
		return json.Marshal(value.Interface())
	case ApplicationForm:
		if v, ok := value.(*primitive.Map); !ok {
			return nil, errors.WithStack(encoding.ErrUnsupportedValue)
		} else {
			urlValues := url.Values{}
			for _, key := range v.Keys() {
				if k, ok := key.(primitive.String); ok {
					value := v.GetOr(k, nil)
					if v, ok := value.(primitive.String); ok {
						urlValues.Add(k.String(), v.String())
					} else if v, ok := value.(*primitive.Slice); ok {
						for i := 0; i < v.Len(); i++ {
							if e, ok := v.Get(i).(primitive.String); ok {
								urlValues.Add(k.String(), e.String())
							}
						}
					}
				}
			}
			return []byte(urlValues.Encode()), nil
		}
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

		writeField := func(obj *primitive.Map, key primitive.Value) error {
			if key, ok := key.(primitive.String); ok {
				value := obj.GetOr(key, nil)

				var elements *primitive.Slice
				if v, ok := value.(*primitive.Slice); ok {
					elements = v
				} else {
					elements = primitive.NewSlice(value)
				}

				for _, element := range elements.Values() {
					contentType := ""
					bytes, err := MarshalMIME(element, &contentType)
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
					} else if _, err := writer.Write(bytes); err != nil {
						return err
					}
				}
			}
			return nil
		}
		writeFields := func(value primitive.Value) error {
			if value, ok := value.(*primitive.Map); ok {
				for _, key := range value.Keys() {
					if err := writeField(value, key); err != nil {
						return err
					}
				}
			}
			return nil
		}
		writeFiles := func(value primitive.Value) error {
			if value, ok := value.(*primitive.Map); ok {
				for _, key := range value.Keys() {
					if key, ok := key.(primitive.String); ok {
						value := value.GetOr(key, nil)

						var elements *primitive.Slice
						if v, ok := value.(*primitive.Slice); ok {
							elements = v
						} else {
							elements = primitive.NewSlice(value)
						}

						for _, element := range elements.Values() {
							data, ok := primitive.Pick[primitive.Value](element, "data")
							if !ok {
								data = element
							}
							filename, ok := primitive.Pick[string](element, "filename")
							if !ok {
								filename = key.String()
							}

							header, _ := primitive.Pick[primitive.Value](element, "header")

							contentType := ""
							contentTypes, _ := primitive.Pick[primitive.Value](header, HeaderContentType)
							if contentTypes != nil {
								if c, ok := contentTypes.(*primitive.Slice); ok {
									contentType, _ = primitive.Pick[string](c, "0")
								} else if c, ok := contentTypes.(primitive.String); ok {
									contentType = c.String()
								}
							}

							contentEncoding := ""
							contentEncodings, _ := primitive.Pick[primitive.Value](header, HeaderContentEncoding)
							if contentEncodings != nil {
								if c, ok := contentEncodings.(*primitive.Slice); ok {
									contentEncoding, _ = primitive.Pick[string](c, "0")
								} else if c, ok := contentEncodings.(primitive.String); ok {
									contentEncoding = c.String()
								}
							}

							bytes, err := MarshalMIME(data, &contentType)
							if err != nil {
								return err
							}
							bytes, err = Compress(bytes, contentEncoding)
							if err != nil {
								return err
							}

							h := textproto.MIMEHeader{}
							if err := primitive.Unmarshal(header, &h); err != nil {
								return err
							}
							h.Set(HeaderContentDisposition, fmt.Sprintf(`form-data; name="%s"; filename="%s"`, quoteEscaper.Replace(key.String()), quoteEscaper.Replace(filename)))
							h.Set(HeaderContentType, contentType)

							if writer, err := mw.CreatePart(h); err != nil {
								return err
							} else if _, err := writer.Write(bytes); err != nil {
								return err
							}
						}
					}
				}
			}
			return nil
		}

		if v, ok := value.(*primitive.Map); ok {
			for _, key := range v.Keys() {
				value := v.GetOr(key, nil)

				if key == primitive.NewString("value") {
					writeFields(value)
				} else if key == primitive.NewString("file") {
					writeFiles(value)
				} else {
					writeField(v, key)
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

// UnmarshalMIME converts MIME data into a primitive.Value.
func UnmarshalMIME(data []byte, contentType *string) (primitive.Value, error) {
	if len(data) == 0 {
		return nil, nil
	}

	if contentType == nil {
		contentType = lo.ToPtr[string]("")
	}
	if *contentType == "" {
		*contentType = http.DetectContentType(data)
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
		return primitive.MarshalText(v)
	case ApplicationForm:
		v, err := url.ParseQuery(string(data))
		if err != nil {
			return nil, err
		}
		return primitive.MarshalText(v)
	case TextPlain:
		return primitive.NewString(string(data)), nil
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
				bytes, err := io.ReadAll(file)
				if err != nil {
					return nil, err
				}

				contentType := fh.Header.Get(HeaderContentType)
				contentEncoding := fh.Header.Get(HeaderContentEncoding)

				bytes, err = Decompress(bytes, contentEncoding)
				if err != nil {
					return nil, err
				}
				data, err := UnmarshalMIME(bytes, &contentType)
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

		return primitive.MarshalText(map[string]any{
			"value": form.Value,
			"file":  files,
		})
	case ApplicationOctetStream:
		return primitive.NewBinary(data), nil
	default:
		return primitive.NewBinary(data), nil
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
