package flag

import "github.com/iancoleman/strcase"

func ToKey(flag string) string {
	return strcase.ToSnake(flag)
}

func ToShorthand(flag string) string {
	if flag == "" {
		return ""
	}
	return flag[0:1]
}
