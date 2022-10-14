package plumber

import (
	"context"
)

type plumberKey int

const configKey plumberKey = 1

type Config struct {
	PlizerEnabled  bool                     // false globally disable parallelization
	RglatorsByName map[string]uint          // number of parallelizable functions by name
	rglatorChans   map[string]chan struct{} // buffered channels used for regulation
}

func ContextWithConfig(ctx context.Context, config Config) context.Context {
	config.RglatorsByName[""] = 0
	config.rglatorChans = map[string]chan struct{}{}
	for name, limit := range config.RglatorsByName {
		config.rglatorChans[name] = make(chan struct{}, limit)
		for i := uint(0); i < limit; i++ {
			config.rglatorChans[name] <- struct{}{}
		}
	}
	return context.WithValue(ctx, configKey, config)
}

func ConfigFromContext(ctx context.Context) (config Config, ok bool) {
	config, ok = ctx.Value(configKey).(Config)
	return
}
