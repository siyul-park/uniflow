package cli

const (
	flagNamespace = "namespace"
	flagFile      = "file"
)

func toShorthand(flag string) string {
	if flag == "" {
		return ""
	}
	return flag[0:1]
}
