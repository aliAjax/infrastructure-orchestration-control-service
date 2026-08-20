package application

import "context"

// RunWithRequestContext keeps execution work within the caller's lifecycle.
// A request that is already cancelled short-circuits before work begins, and a
// live request's cancellation, deadline, and values are handed straight to the
// work instead of being detached onto a fresh background context.
func RunWithRequestContext(ctx context.Context, work func(context.Context) error) error {
	checked, err := executionContext(ctx)
	if err != nil {
		return err
	}
	if checked != nil {
		ctx = checked
	}
	if work == nil {
		return nil
	}
	return work(ctx)
}
