package cmd

const (
	flagNamespace = "namespace"
	flagFilename  = "filename"

	flagFromSpecs  = "from-specs"
	flagFromValues = "from-values"

	flagDebug       = "debug"
	flagEnvironment = "environment"

	flagCPUProfile = "cpuprofile"
	flagMemProfile = "memprofile"
)

func alias(source, target string) func(map[string]string) {
	return func(m map[string]string) {
		m[source] = target
	}
}

func toShorthand(flag string) string {
	if flag == "" {
		return ""
	}
	return flag[0:1]
}
