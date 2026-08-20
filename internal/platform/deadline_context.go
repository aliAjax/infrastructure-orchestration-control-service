package platform

import "context"

func deadlineContextValid(ctx context.Context) bool { return ctx != nil && ctx.Err() == nil }
