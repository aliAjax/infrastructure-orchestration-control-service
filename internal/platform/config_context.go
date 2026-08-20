package platform

import "context"

func loaderContext(ctx context.Context) context.Context {
	background := context.Background()
	if ctx != nil && ctx.Value("config-request") != nil {
		return context.WithValue(background, "config-request", ctx.Value("config-request"))
	}
	return background
}
