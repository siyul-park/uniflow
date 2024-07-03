package mime

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/object"
)

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

// Encode encodes the given object into the writer with the specified MIME headers.
func Encode(writer io.Writer, value object.Object, header textproto.MIMEHeader) error {
	typ := header.Get(HeaderContentType)
	encode := header.Get(HeaderContentEncoding)

	count := 0
	var w io.Writer = WriterFunc(func(b []byte) (n int, err error) {
		n, err = writer.Write(b)
		count += n
		return
	})

	if typ == "" {
		switch value.(type) {
		case object.Binary:
			typ = ApplicationOctetStream
		case object.String:
			typ = TextPlainCharsetUTF8
		default:
			typ = ApplicationJSONCharsetUTF8
		}
		header.Set(HeaderContentType, typ)
	}

	typ, params, err := mime.ParseMediaType(typ)
	if err != nil {
		return err
	}

	w, err = Compress(w, encode)
	if err != nil {
		return err
	}

	switch typ {
	case ApplicationJSON:
		e := json.NewEncoder(w)
		if err := e.Encode(object.InterfaceOf(value)); err != nil {
			return err
		} else {
			header.Set(HeaderContentLength, strconv.Itoa(count))
		}
		return nil
	case ApplicationFormURLEncoded:
		urlValues := url.Values{}
		if err := object.Unmarshal(value, &urlValues); err != nil {
			return err
		} else if _, err := w.Write([]byte(urlValues.Encode())); err != nil {
			return err
		} else {
			header.Set(HeaderContentLength, strconv.Itoa(count))
		}
		return nil
	case MultipartFormData:
		boundary := params["boundary"]
		if boundary == "" {
			boundary = randomMultiPartBoundary()
			params["boundary"] = boundary
			header.Set(HeaderContentType, mime.FormatMediaType(typ, params))
		}

		mw := multipart.NewWriter(w)
		if err := mw.SetBoundary(boundary); err != nil {
			return err
		}

		writeField := func(obj object.Map, key object.Object) error {
			if key, ok := key.(object.String); ok {
				value := obj.GetOr(key, nil)

				var elements object.Slice
				if v, ok := value.(object.Slice); ok {
					elements = v
				} else {
					elements = object.NewSlice(value)
				}

				for _, element := range elements.Values() {
					h := textproto.MIMEHeader{}
					h.Set(HeaderContentDisposition, fmt.Sprintf(`form-data; name="%s"`, quoteEscaper.Replace(key.String())))

					if w, err := mw.CreatePart(h); err != nil {
						return err
					} else if err := Encode(w, element, h); err != nil {
						return err
					}
				}
			}
			return nil
		}

		writeFields := func(value object.Object) error {
			if value, ok := value.(object.Map); ok {
				for _, key := range value.Keys() {
					if err := writeField(value, key); err != nil {
						return err
					}
				}
			}
			return nil
		}

		writeFiles := func(value object.Object) error {
			if value, ok := value.(object.Map); ok {
				for _, key := range value.Keys() {
					if key, ok := key.(object.String); ok {
						value := value.GetOr(key, nil)

						var elements object.Slice
						if v, ok := value.(object.Slice); ok {
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

							h := textproto.MIMEHeader{}
							_ = object.Unmarshal(header, &h)

							typ := h.Get(HeaderContentType)
							if typ == "" {
								switch data.(type) {
								case object.Binary:
									typ = ApplicationOctetStream
								case object.String:
									typ = TextPlainCharsetUTF8
								default:
									typ = ApplicationJSONCharsetUTF8
								}
								h.Set(HeaderContentType, typ)
							}

							typ, params, err := mime.ParseMediaType(typ)
							if err != nil {
								return err
							}

							if typ == MultipartFormData {
								boundary := params["boundary"]
								if boundary == "" {
									boundary = randomMultiPartBoundary()
									params["boundary"] = boundary
									h.Set(HeaderContentType, mime.FormatMediaType(typ, params))
								}
							}

							h.Set(HeaderContentDisposition, fmt.Sprintf(`form-data; name="%s"; filename="%s"`, quoteEscaper.Replace(key.String()), quoteEscaper.Replace(filename)))

							if writer, err := mw.CreatePart(h); err != nil {
								return err
							} else if err := Encode(writer, data, h); err != nil {
								return err
							}
						}
					}
				}
			}
			return nil
		}

		if v, ok := value.(object.Map); ok {
			for _, key := range v.Keys() {
				value := v.GetOr(key, nil)

				if key.Equal(object.NewString("value")) {
					if err := writeFields(value); err != nil {
						return err
					}
				} else if key.Equal(object.NewString("file")) {
					if err := writeFiles(value); err != nil {
						return err
					}
				} else if err := writeField(v, key); err != nil {
					return err
				}
			}
		}

		if err := mw.Close(); err != nil {
			return err
		} else {
			header.Set(HeaderContentLength, strconv.Itoa(count))
		}
		return nil
	}

	if v, ok := value.(object.Binary); ok {
		if _, err := w.Write(v.Bytes()); err != nil {
			return err
		} else {
			header.Set(HeaderContentLength, strconv.Itoa(count))
		}
		return nil
	}
	if v, ok := value.(object.String); ok {
		if _, err := w.Write([]byte(v.String())); err != nil {
			return err
		} else {
			header.Set(HeaderContentLength, strconv.Itoa(count))
		}
		return nil
	}

	return errors.WithStack(encoding.ErrUnsupportedValue)
}

// Decode decodes the given reader with the specified MIME headers into an object.
func Decode(reader io.Reader, header textproto.MIMEHeader) (object.Object, error) {
	typ := header.Get(HeaderContentType)
	encode := header.Get(HeaderContentEncoding)

	typ, params, _ := mime.ParseMediaType(typ)

	reader, err := Decompress(reader, encode)
	if err != nil {
		return nil, err
	}

	switch typ {
	case ApplicationJSON:
		var data any
		d := json.NewDecoder(reader)
		if err := d.Decode(&data); err != nil {
			return nil, err
		}
		return object.MarshalText(data)
	case ApplicationFormURLEncoded:
		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		v, err := url.ParseQuery(string(data))
		if err != nil {
			return nil, err
		}
		return object.MarshalText(v)
	case TextPlain:
		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		return object.NewString(string(data)), nil
	case MultipartFormData:
		reader := multipart.NewReader(reader, params["boundary"])

		form, err := reader.ReadForm(0)
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
				defer file.Close()

				data, err := Decode(file, fh.Header)
				if err != nil {
					return nil, err
				}

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
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return object.NewBinary(data), nil
}

func randomMultiPartBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}
