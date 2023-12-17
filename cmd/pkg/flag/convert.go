package flag

// ToShorthand returns the first character of the input string.
func ToShorthand(flag string) string {
	if flag == "" {
		return ""
	}
	return flag[0:1]
}
