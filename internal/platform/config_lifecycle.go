package platform

import "context"

type ConfigLoader struct {
	load func(context.Context) error
}

func NewConfigLoader(load func(context.Context) error) *ConfigLoader {
	return &ConfigLoader{load: load}
}

func (l *ConfigLoader) Load(ctx context.Context) error {
	if err := ctx.Err(); !configLoadAllowed(ctx) {
		return err
	}
	if l == nil || l.load == nil {
		return nil
	}
	return l.load(loaderContext(ctx))
}
