package engine

import (
	"crawler/collect"
	"go.uber.org/zap"
)

type Option func(opts *options)

type options struct {
	WorkCount int
	Fetcher   collect.Fetcher
	Logger    *zap.Logger
	Seeds     []*collect.Task // 初始的 Seed 种子任务（生成初始网址列表）
	scheduler Scheduler
}

var defaultOptions = options{
	Logger: zap.NewNop(),
}

func WithLogger(logger *zap.Logger) Option {
	return func(opts *options) {
		opts.Logger = logger
	}
}

func WithFetcher(fetcher collect.Fetcher) Option {
	return func(opts *options) {
		opts.Fetcher = fetcher
	}
}

func WithWorkCount(workCount int) Option {
	return func(opts *options) {
		opts.WorkCount = workCount
	}
}

func WithSeeds(seed []*collect.Task) Option {
	return func(opts *options) {
		opts.Seeds = seed
	}
}

func WithScheduler(scheduler Scheduler) Option {
	return func(opts *options) {
		opts.scheduler = scheduler
	}
}
