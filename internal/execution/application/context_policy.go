package application

import "context"

// executionContext is the gatekeeper for request-scoped execution: it rejects
// callers whose request has already been cancelled so downstream work cannot
// run detached from the caller's lifecycle, and otherwise hands back the very
// same context so its cancellation, deadline, and values stay reachable.
func executionContext(ctx context.Context) (context.Context, error) {
	if ctx == nil {
		return nil, context.Canceled
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return ctx, nil
}
