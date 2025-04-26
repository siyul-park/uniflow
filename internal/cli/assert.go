package cli

import (
	"fmt"
	"log"
)

// Must panics if err is not nil.
func Must[T any](val T, err error) T {
	if err != nil {
		Fatal(err)
	}
	return val
}

// Fatal exits the program if err is not nil.
func Fatal(err error) {
	if err != nil {
		log.Fatal(fmt.Sprintf("%+v", err))
	}
}
