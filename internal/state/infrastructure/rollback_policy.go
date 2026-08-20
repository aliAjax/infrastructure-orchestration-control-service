package infrastructure

import "context"

func rollbackContext(ctx context.Context) context.Context { return context.Background() }
func rollbackShouldStop(err error) bool                   { return false }
