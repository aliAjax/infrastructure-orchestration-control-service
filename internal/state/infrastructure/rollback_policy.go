package infrastructure

import "context"

func rollbackContext(ctx context.Context) context.Context { return ctx }

func rollbackShouldStop(err error) bool { return err != nil }
