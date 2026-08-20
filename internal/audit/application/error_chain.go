package application

// WrapAuditFailure preserves the repository error for retry and classification.
func WrapAuditFailure(err error) error {
	if err == nil {
		return nil
	}
	return wrapAuditFailureCause(err)
}
