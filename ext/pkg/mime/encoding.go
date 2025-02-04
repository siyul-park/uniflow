package mime

import (
	"bytes"
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
	"github.com/siyul-park/uniflow/pkg/types"
)

type byteCounter struct {
	io.Writer
	count int
}

var (
	keyValues = types.NewString("values")
	keyFiles  = types.NewString("files")
)

var escapeQuotes = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

var _ io.Writer = (*byteCounter)(nil)
var _ io.Closer = (*byteCounter)(nil)

// Encode encodes the given types into the writer with the specified MIME headers.
func Encode(writer io.Writer, value types.Value, header textproto.MIMEHeader) error {
	if header == nil {
		header = textproto.MIMEHeader{}
	}

	typ := header.Get(HeaderContentType)
	encode := header.Get(HeaderContentEncoding)

	if typ == "" {
		if detects := DetectTypesFromValue(value); len(detects) > 0 {
			typ = detects[0]
			header.Set(HeaderContentType, typ)
		}
	}

	counter := &byteCounter{Writer: writer}
	defer header.Set(HeaderContentLength, strconv.Itoa(counter.Count()))

	w, err := Compress(counter, encode)
	if err != nil {
		return err
	}
	if c, ok := w.(io.Closer); ok {
		defer c.Close()
	}

	if v, ok := value.(types.Buffer); ok {
		if _, err := io.Copy(w, v); err != nil {
			return err
		}
		return nil
	} else if v, ok := value.(types.Binary); ok {
		if _, err := w.Write(v.Bytes()); err != nil {
			return err
		}
		return nil
	}

	typ, params, err := mime.ParseMediaType(typ)
	if err != nil {
		return err
	}

	switch typ {
	case ApplicationJSON:
		if err := json.NewEncoder(w).Encode(value); err != nil {
			return err
		}
		return nil
	case ApplicationFormURLEncoded:
		urlValues := url.Values{}
		if err := types.Unmarshal(value, &urlValues); err != nil {
			return err
		}
		if _, err := w.Write([]byte(urlValues.Encode())); err != nil {
			return err
		}
		return nil
	case TextPlain:
		if v, ok := value.(types.String); ok {
			if _, err := w.Write([]byte(v.String())); err != nil {
				return err
			}
			return nil
		}
		return nil
	case MultipartFormData:
		boundary := params["boundary"]
		if boundary == "" {
			boundary = randomMultipartBoundary()
			params["boundary"] = boundary
			header.Set(HeaderContentType, mime.FormatMediaType(typ, params))
		}

		mw := multipart.NewWriter(w)
		if err := mw.SetBoundary(boundary); err != nil {
			return err
		}

		writeFormField := func(obj types.Map, key types.Value) error {
			if key, ok := key.(types.String); ok {
				value := obj.Get(key)

				var elements types.Slice
				if v, ok := value.(types.Slice); ok {
					elements = v
				} else {
					elements = types.NewSlice(value)
				}

				for _, element := range elements.Range() {
					h := textproto.MIMEHeader{}
					h.Set(HeaderContentDisposition, fmt.Sprintf(`form-data; name="%s"`, escapeQuotes.Replace(key.String())))

					if w, err := mw.CreatePart(h); err != nil {
						return err
					} else if err := Encode(w, element, h); err != nil {
						return err
					}
				}
			}
			return nil
		}

		writeFormFields := func(value types.Value) error {
			if value, ok := value.(types.Map); ok {
				for key := range value.Range() {
					if err := writeFormField(value, key); err != nil {
						return err
					}
				}
			}
			return nil
		}

		writeFormFiles := func(value types.Value) error {
			if value, ok := value.(types.Map); ok {
				for key := range value.Range() {
					if key, ok := key.(types.String); ok {
						value := value.Get(key)

						var elements types.Slice
						if v, ok := value.(types.Slice); ok {
							elements = v
						} else {
							elements = types.NewSlice(value)
						}

						for _, element := range elements.Values() {
							data, ok := types.Get[types.Value](element, "data")
							if !ok {
								data = element
							}
							filename, ok := types.Get[string](element, "filename")
							if !ok {
								filename = key.String()
							}

							header, _ := types.Get[types.Value](element, "header")

							h := textproto.MIMEHeader{}
							_ = types.Unmarshal(header, &h)

							typ := h.Get(HeaderContentType)
							if typ == "" {
								if detects := DetectTypesFromValue(data); len(detects) > 0 {
									typ = detects[0]
									h.Set(HeaderContentType, typ)
								}
							}

							typ, params, err := mime.ParseMediaType(typ)
							if err != nil {
								return err
							}

							if typ == MultipartFormData {
								boundary := params["boundary"]
								if boundary == "" {
									boundary = randomMultipartBoundary()
									params["boundary"] = boundary
									h.Set(HeaderContentType, mime.FormatMediaType(typ, params))
								}
							}

							h.Set(HeaderContentDisposition, fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes.Replace(key.String()), escapeQuotes.Replace(filename)))

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

		if v, ok := value.(types.Map); ok {
			for key, value := range v.Range() {
				if key.Equal(keyValues) {
					if err := writeFormFields(value); err != nil {
						return err
					}
				} else if key.Equal(keyFiles) {
					if err := writeFormFiles(value); err != nil {
						return err
					}
				} else if err := writeFormField(v, key); err != nil {
					return err
				}
			}
		}

		if err := mw.Close(); err != nil {
			return err
		}
		return nil
	}

	return errors.WithStack(encoding.ErrUnsupportedType)
}

// Decode decodes the given reader with the specified MIME headers into types.
func Decode(reader io.Reader, header textproto.MIMEHeader) (types.Value, error) {
	typ := header.Get(HeaderContentType)
	encode := header.Get(HeaderContentEncoding)

	r, err := Decompress(reader, encode)
	if err != nil {
		return nil, err
	}

	defer func() {
		if c, ok := r.(io.Closer); ok {
			_ = c.Close()
		}
	}()

	if typ == "" {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		if detects := DetectTypesFromBytes(data); len(detects) > 0 {
			typ = detects[0]
			header.Set(HeaderContentType, typ)
		}

		r = bytes.NewReader(data)
	}

	typ, params, _ := mime.ParseMediaType(typ)

	switch typ {
	case ApplicationJSON:
		var data any
		d := json.NewDecoder(r)
		if err := d.Decode(&data); err != nil {
			return nil, err
		}
		return types.Marshal(data)
	case ApplicationFormURLEncoded:
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		v, err := url.ParseQuery(string(data))
		if err != nil {
			return nil, err
		}
		return types.Marshal(v)
	case TextPlain:
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return types.NewString(string(data)), nil
	case MultipartFormData:
		parts := multipart.NewReader(r, params["boundary"])

		form, err := parts.ReadForm(0)
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

				data, err := Decode(file, fh.Header)
				file.Close()
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

		return types.Marshal(map[string]any{
			keyValues.String(): form.Value,
			keyFiles.String():  files,
		})
	}

	buf := types.NewBuffer(r)
	r = io.NopCloser(r)
	return buf, nil
}

func randomMultipartBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}

func (w *byteCounter) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	w.count += n
	return n, err
}

func (w *byteCounter) Close() error {
	if c, ok := w.Writer.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (w *byteCounter) Count() int {
	return w.count
}
