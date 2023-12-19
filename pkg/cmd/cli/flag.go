package cli

const (
	flagNamespace = "namespace"
	flagFile      = "file"
	flagBoot      = "boot"
)

func toShorthand(flag string) string {
	if flag == "" {
		return ""
	}
	return flag[0:1]
}
