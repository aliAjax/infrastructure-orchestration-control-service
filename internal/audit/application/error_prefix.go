package application

func auditFailurePrefix(prefix string) string {
	if prefix == "" {
		return "audit persistence failed"
	}
	return prefix
}
