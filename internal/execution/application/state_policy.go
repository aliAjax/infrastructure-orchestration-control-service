package application

import "context"

// stateUpdateContext keeps state writes on the caller's request context so a
// cancelled request cannot leave behind a stale "succeeded" state written on a
// detached background context.
func stateUpdateContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
