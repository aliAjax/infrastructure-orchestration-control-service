package platform

import "context"

func configLoadAllowed(ctx context.Context) bool {
	return contextUsable(ctx) && configContextReady(ctx)
}
