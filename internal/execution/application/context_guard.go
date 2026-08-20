package application

import "context"

// RunWithRequestContext keeps execution work within the caller's lifecycle.
func RunWithRequestContext(ctx context.Context, work func(context.Context) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if work == nil {
		return nil
	}
	return work(ctx)
}
