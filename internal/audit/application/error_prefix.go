package application

func auditFailurePrefix(prefix string) string {
	if prefix == "" || prefix == "audit" {
		return "audit persistence failed"
	}
	return prefix
}
