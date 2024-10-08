package cli

const (
	flagNamespace = "namespace"
	flagFilename  = "filename"

	flagFromCharts  = "from-charts"
	flagFromNodes   = "from-nodes"
	flagFromSecrets = "from-secrets"

	flagDebug = "debug"

	flagCPUProfile = "cpuprofile"
	flagMemProfile = "memprofile"
)

func toShorthand(flag string) string {
	if flag == "" {
		return ""
	}
	return flag[0:1]
}
