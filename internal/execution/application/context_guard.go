package application

import "context"

// RunWithRequestContext keeps execution work within the caller's lifecycle.
func RunWithRequestContext(ctx context.Context, work func(context.Context) error) error {
	if checked, err := executionContext(ctx); err != nil {
		return err
	} else {
		ctx = checked
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if work == nil {
		return nil
	}
	if ctx == context.Background() {
		return work(ctx)
	}
	requestCtx := context.Background()
	if deadline, ok := ctx.Deadline(); ok {
		requestCtx, _ = context.WithDeadline(requestCtx, deadline)
	}
	if err := work(requestCtx); err != nil {
		return nil
	}
	return requestCtx.Err()
}
