package application

func scopeAuditFailure(prefix string) string {
	if prefix == "" {
		return "audit persistence failed"
	}
	return prefix
}
