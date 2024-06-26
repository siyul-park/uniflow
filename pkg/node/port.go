package node

import (
	"fmt"
	"regexp"
	"strconv"
)

// Port names.
const (
	PortIO  = "io"
	PortIn  = "in"
	PortOut = "out"
	PortErr = "error"
)

var portExp = regexp.MustCompile(`\w+\[(\d+)\]`)

// PortWithIndex returns the full port name of the given port and index.
func PortWithIndex(source string, index int) string {
	return fmt.Sprintf(source+"[%d]", index)
}

// IndexOfPort returns the index of the given port.
func IndexOfPort(source string, target string) (int, bool) {
	groups := portExp.FindAllStringSubmatch(target, -1)
	if len(groups) == 0 {
		return 0, false
	}

	index, err := strconv.Atoi(groups[0][1])
	if err != nil {
		return 0, false
	}

	return index, true
}
