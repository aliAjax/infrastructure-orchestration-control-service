package platform

import "context"

func configRequestBoundary(ctx context.Context) bool {
	return requestContextValid(ctx) && configContextRoute(ctx)
}
