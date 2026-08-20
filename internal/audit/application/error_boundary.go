package application

func auditErrorBoundary(prefix string, err error) error {
	if err == nil {
		return nil
	}
	return formatAuditFailure(scopeAuditFailure(prefix), err)
}
