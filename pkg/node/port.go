package node

import (
	"fmt"
	"regexp"
	"strconv"
)

// Commonly used port names.
const (
	PortInit = "init"
	PortIO   = "io"
	PortIn   = "in"
	PortOut  = "out"
	PortErr  = "error"
)

var portExp = regexp.MustCompile(`(\w+)\[(\d+)\]`)

// PortWithIndex returns the full port name formatted as "name[index]".
func PortWithIndex(name string, index int) string {
	return fmt.Sprintf("%s[%d]", name, index)
}

// NameOfPort extracts and returns the base name from a port name formatted as "name[index]".
func NameOfPort(name string) string {
	groups := portExp.FindStringSubmatch(name)
	if groups == nil {
		return name
	}
	return groups[1]
}

// IndexOfPort extracts the index from a port name formatted as "name[index]" and returns it with a boolean indicating success.
func IndexOfPort(name string) (int, bool) {
	groups := portExp.FindStringSubmatch(name)
	if groups == nil || len(groups) < 2 {
		return 0, false
	}

	index, err := strconv.Atoi(groups[2])
	return index, err == nil
}
