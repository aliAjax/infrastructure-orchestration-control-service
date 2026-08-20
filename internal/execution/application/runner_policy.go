package application

import "context"

// runnerExecutionContext hands the runner the live request context so a
// cancellation of the request reaches the executor and stops it, rather than
// a detached background context that lets it run on oblivious.
func runnerExecutionContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
