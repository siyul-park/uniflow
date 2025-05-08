package node

import (
	"fmt"
	"regexp"
	"strconv"
)

// Commonly used port names.
const (
	PortInit    = "init"
	PortActive  = "active"
	PortDeative = "deactive"
	PortDeinit  = "deinit"
	PortIO      = "io"
	PortIn      = "in"
	PortOut     = "out"
	PortError   = "error"
)

var subscript = regexp.MustCompile(`(\w+)\[(\d+)\]`)

// PortWithIndex formats the port name as "name[index]".
func PortWithIndex(name string, index int) string {
	return fmt.Sprintf("%s[%d]", name, index)
}

// NameOfPort extracts the base name from a port name formatted as "name[index]".
func NameOfPort(key string) string {
	if groups := subscript.FindStringSubmatch(key); groups != nil {
		return groups[1]
	}
	return key
}

// IndexOfPort extracts the index from a port name formatted as "name[index]".
func IndexOfPort(key string) (int, bool) {
	if groups := subscript.FindStringSubmatch(key); len(groups) == 3 {
		if index, err := strconv.Atoi(groups[2]); err == nil {
			return index, true
		}
	}
	return 0, false
}
