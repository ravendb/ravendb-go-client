package ravendb

import "time"

type AggressiveCacheOptions struct {
	duration time.Duration
}

func (o *AggressiveCacheOptions) getDuration() time.Duration {
	return o.duration
}

func NewAggressiveCacheOptions(duration time.Duration) *AggressiveCacheOptions {
	return &AggressiveCacheOptions{
		duration: duration,
	}
}
