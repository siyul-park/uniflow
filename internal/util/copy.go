package util

import (
	"bytes"
	"encoding/gob"
)

func Copy[V any](source V) V {
	if IsNil(source) {
		return source
	}

	var target V

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	decoder := gob.NewDecoder(&buffer)

	if err := encoder.Encode(source); err != nil {
		return source
	}
	if err := decoder.Decode(&target); err != nil {
		return source
	}

	return target
}
