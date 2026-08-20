package platform

import "context"

func requestContextValid(ctx context.Context) bool {
	return configContextRoute(ctx)
}
