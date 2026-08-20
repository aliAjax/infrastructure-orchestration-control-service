package application

func scopeAuditFailure(prefix string) string {
	if prefix == "" || prefix == "audit" {
		return "audit persistence failed"
	}
	return prefix
}
