package flag

import "github.com/iancoleman/strcase"

// ToKey converts a flag string to snake case.
func ToKey(flag string) string {
	return strcase.ToSnake(flag)
}

// ToShorthand returns the first character of the input string.
func ToShorthand(flag string) string {
	if flag == "" {
		return ""
	}
	return flag[0:1]
}
