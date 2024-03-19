package cli

const (
	flagNamespace = "namespace"
	flagFilename  = "filename"

	flagCPUProfile = "cpuprofile"
	flagMemProfile = "memprofile"
)

func toShorthand(flag string) string {
	if flag == "" {
		return ""
	}
	return flag[0:1]
}
