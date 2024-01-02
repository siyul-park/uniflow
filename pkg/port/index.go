package port

import (
	"fmt"
	"regexp"
	"strconv"
)

// GetIndex returns the index of the given port.
func GetIndex(source string, target string) (int, bool) {
	regex, err := regexp.Compile(source + `\[(\d+)\]`)
	if err != nil {
		return 0, false
	}
	groups := regex.FindAllStringSubmatch(target, -1)
	if len(groups) == 0 {
		return 0, false
	}
	index, err := strconv.Atoi(groups[0][1])
	if err != nil {
		return 0, false
	}
	return index, true
}

// SetIndex returns the full port name of the given port and index.
func SetIndex(source string, index int) string {
	return fmt.Sprintf(source+"[%d]", index)
}
