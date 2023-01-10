package main

import (
	"crawler/collect"
	"crawler/engine"
	"crawler/limiter"
	"crawler/log"
	"crawler/proxy"
	storage2 "crawler/storage"
	"crawler/storage/sqlstorage"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
	"time"
)

// 启动爬虫任务的方式可以分为两种，一种是加载配置文件，另一种是在调用用户接口时，传递任务名称和参数。不过在这里我们先用硬编码的形式来实现。
// 而通过配置文件和用户接口来操作任务的方式我们会有专门的课程来实现。
func main() {
	// log
	plugin := log.NewStdoutPlugin(zapcore.InfoLevel)
	logger := log.NewLogger(plugin)
	logger.Info("log init end")

	// set zap global logger
	zap.ReplaceGlobals(logger)

	// proxy
	proxyURLs := []string{"http://127.0.0.1:8888", "http://127.0.0.1:8889"}
	p, err := proxy.RoundRobinProxySwitcher(proxyURLs...)
	if err != nil {
		logger.Error("RoundRobinProxySwitcher failed")
	}

	var f collect.Fetcher = collect.BrowserFetch{
		Timeout: 3000 * time.Millisecond,
		Logger:  logger,
		Proxy:   p,
	}

	// 后端存储
	var storage storage2.Storage
	storage, err = sqlstorage.New(
		sqlstorage.WithSqlUrl("root:123456@tcp(127.0.0.1:3326)/crawler?charset=utf8"),
		sqlstorage.WithLogger(logger.Named("sqlDB")),
		sqlstorage.WithBatchCount(2),
	)
	if err != nil {
		logger.Error("create sqlstorage failed")
		return
	}

	// 设置限流器
	secondLimit := rate.NewLimiter(limiter.Per(1, 2*time.Second), 1)   // 2秒钟1个
	minuteLimit := rate.NewLimiter(limiter.Per(20, 1*time.Minute), 20) // 60秒20个
	multiLimiter := limiter.MultiLimiter(secondLimit, minuteLimit)

	seeds := make([]*collect.Task, 0, 1000)
	seeds = append(seeds, &collect.Task{
		Property: collect.Property{
			Name: "find_douban_sun_room",
		},
		Fetcher: f,
		Storage: storage,
		Limiter: multiLimiter,
	})

	s := engine.NewEngine(
		engine.WithWorkCount(5),
		engine.WithFetcher(f),
		engine.WithLogger(logger),
		engine.WithSeeds(seeds),
		engine.WithScheduler(engine.NewSchedule()),
	)
	s.Run()
}
