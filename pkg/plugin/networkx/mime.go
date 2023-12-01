package networkx

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

// MIME content types
const (
	ApplicationJSON                  = "application/json"
	ApplicationJSONCharsetUTF8       = ApplicationJSON + "; " + charsetUTF8
	ApplicationJavaScript            = "application/javascript"
	ApplicationJavaScriptCharsetUTF8 = ApplicationJavaScript + "; " + charsetUTF8
	ApplicationXML                   = "application/xml"
	ApplicationXMLCharsetUTF8        = ApplicationXML + "; " + charsetUTF8
	TextXML                          = "text/xml"
	TextXMLCharsetUTF8               = TextXML + "; " + charsetUTF8
	ApplicationForm                  = "application/x-www-form-urlencoded"
	ApplicationProtobuf              = "application/protobuf"
	ApplicationMsgpack               = "application/msgpack"
	TextHTML                         = "text/html"
	TextHTMLCharsetUTF8              = TextHTML + "; " + charsetUTF8
	TextPlain                        = "text/plain"
	TextPlainCharsetUTF8             = TextPlain + "; " + charsetUTF8
	MultipartForm                    = "multipart/form-data"
	OctetStream                      = "application/octet-stream"
)

const charsetUTF8 = "charset=utf-8"

// MarshalMIME marshals a primitive.Value to MIME format.
func MarshalMIME(value primitive.Value, typ *string) ([]byte, error) {
	if typ == nil {
		content := ""
		typ = &content
	}

	if value == nil {
		return nil, nil
	} else if v, ok := value.(primitive.String); ok {
		data := []byte(v.String())
		if *typ == "" {
			*typ = http.DetectContentType(data)
		}
		return data, nil
	} else if v, ok := value.(primitive.Binary); ok {
		data := v.Bytes()
		if *typ == "" {
			*typ = http.DetectContentType(data)
		}
		return data, nil
	}

	if *typ == "" {
		*typ = ApplicationJSONCharsetUTF8
	}

	mediatype, params, err := mime.ParseMediaType(*typ)
	if err != nil {
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	}

	switch mediatype {
	case ApplicationJSON:
		return json.Marshal(value.Interface())
	case ApplicationXML, TextXML:
		return xml.Marshal(value.Interface())
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
	case MultipartForm:
		boundary, ok := params["boundary"]
		if !ok {
			boundary = randomMultiPartBoundary()
			params["boundary"] = boundary
			*typ = mime.FormatMediaType(mediatype, params)
		}

		bodyBuffer := new(bytes.Buffer)
		mw := multipart.NewWriter(bodyBuffer)
		if err := mw.SetBoundary(boundary); err != nil {
			return nil, err
		}

		writeField := func(obj *primitive.Map, key primitive.Value) error {
			if key, ok := key.(primitive.String); ok {
				elements := obj.GetOr(key, nil)
				if e, ok := elements.(primitive.String); ok {
					if err := mw.WriteField(key.String(), e.String()); err != nil {
						return err
					}
				} else if e, ok := elements.(*primitive.Slice); ok {
					for i := 0; i < e.Len(); i++ {
						if e, ok := e.Get(i).(primitive.String); ok {
							if err := mw.WriteField(key.String(), e.String()); err != nil {
								return err
							}
						}
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
						elements := value.GetOr(key, nil)
						if e, ok := elements.(*primitive.Map); ok {
							filename, ok := e.GetOr(primitive.NewString("filename"), nil).(primitive.String)
							if !ok {
								continue
							}
							writer, err := mw.CreateFormFile(key.String(), filename.String())
							if err != nil {
								return err
							}

							data, ok := e.Get(primitive.NewString("data"))
							if !ok {
								continue
							}
							if d, ok := data.(primitive.Binary); ok {
								if _, err := writer.Write(d.Bytes()); err != nil {
									return err
								}
							} else if d, ok := data.(primitive.String); ok {
								if _, err := writer.Write([]byte(d.String())); err != nil {
									return err
								}
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

// UnmarshalMIME unmarshals MIME data to a primitive.Value.
func UnmarshalMIME(data []byte, typ *string) (primitive.Value, error) {
	if len(data) == 0 {
		return nil, nil
	}

	if typ == nil {
		content := ""
		typ = &content
	}
	if *typ == "" {
		*typ = http.DetectContentType(data)
	}

	mediatype, params, err := mime.ParseMediaType(*typ)
	if err != nil {
		if len(data) == 0 {
			return nil, nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	}

	switch mediatype {
	case ApplicationJSON:
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return primitive.MarshalText(v)
	case ApplicationXML, TextXML:
		var v any
		if err := xml.Unmarshal(data, &v); err != nil {
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
	case MultipartForm:
		reader := multipart.NewReader(bytes.NewReader(data), params["boundary"])
		form, err := reader.ReadForm(int64(len(data)))
		if err != nil {
			return nil, err
		}
		defer form.RemoveAll()

		formFile := map[string][]map[string]any{}
		for name, fhs := range form.File {
			for _, fh := range fhs {
				file, err := fh.Open()
				if err != nil {
					return nil, err
				}
				data, err := io.ReadAll(file)
				if err != nil {
					return nil, err
				}

				formFile[name] = append(formFile[name], map[string]any{
					"filename": fh.Filename,
					"header":   fh.Header,
					"size":     fh.Size,
					"data":     data,
				})
			}
		}

		return primitive.MarshalText(map[string]any{
			"value": form.Value,
			"file":  formFile,
		})
	case OctetStream:
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
