package application

func classifyAuditFailure(err error) error {
	return auditErrorBoundary("audit persistence failed", err)
}

func wrapAuditFailureCause(err error) error {
	return classifyAuditFailure(err)
}
