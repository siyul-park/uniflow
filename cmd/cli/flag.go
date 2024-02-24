package cli

const (
	flagNamespace = "namespace"
	flagFilename  = "filename"
)

func toShorthand(flag string) string {
	if flag == "" {
		return ""
	}
	return flag[0:1]
}
