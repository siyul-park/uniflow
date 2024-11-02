package node

import (
	"fmt"
	"regexp"
	"strconv"
)

// Commonly used port names.
const (
	PortInit = "init"
	PortTerm = "term"
	PortIO   = "io"
	PortIn   = "in"
	PortOut  = "out"
	PortErr  = "error"
)

var portExp = regexp.MustCompile(`(\w+)\[(\d+)\]`)

// PortWithIndex formats the port name as "name[index]".
func PortWithIndex(name string, index int) string {
	return fmt.Sprintf("%s[%d]", name, index)
}

// NameOfPort extracts the base name from a port name formatted as "name[index]".
func NameOfPort(name string) string {
	if groups := portExp.FindStringSubmatch(name); groups != nil {
		return groups[1]
	}
	return name
}

// IndexOfPort extracts the index from a port name formatted as "name[index]".
func IndexOfPort(name string) (int, bool) {
	if groups := portExp.FindStringSubmatch(name); len(groups) == 3 {
		if index, err := strconv.Atoi(groups[2]); err == nil {
			return index, true
		}
	}
	return 0, false
}
