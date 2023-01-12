package limiter

import (
	"context"
	"golang.org/x/time/rate"
	"sort"
	"time"
)

type RateLimiter interface {
	Wait(context.Context) error
	Limit() rate.Limit
}

// Per 封装了生成速率的函数，例如 limiter.Per(20, 1*time.Minute) 代表速率是每 1 分钟补充 20 个令牌。
func Per(eventCount int, duration time.Duration) rate.Limit {
	return rate.Every(duration / time.Duration(eventCount)) // rate.Every(500*time.Millisecond) 表示每 500 毫秒放入一个令牌，换算过来就是每秒钟放入 2 个令牌。
}

type multiLimiter struct {
	limiters []RateLimiter
}

// MultiLimiter 函数用于聚合多个 RateLimiter，并将速率由小到大排序。
func MultiLimiter(limiters ...RateLimiter) *multiLimiter {
	sort.Slice(limiters, func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	})
	return &multiLimiter{limiters: limiters}
}

// Wait 方法会循环遍历多层限速器 multiLimiter 中所有的限速器并索要令牌，只有当所有的限速器规则都满足后，才会正常执行后续的操作。
func (l *multiLimiter) Wait(ctx context.Context) error {
	for _, l := range l.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (l *multiLimiter) Limit() rate.Limit {
	return l.limiters[0].Limit()
}
