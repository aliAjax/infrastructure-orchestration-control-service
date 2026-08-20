package platform

import "context"

func contextUsable(ctx context.Context) bool { return requestContextValid(ctx) }
