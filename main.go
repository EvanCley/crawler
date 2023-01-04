package main

import (
	"crawler/collect"
	"crawler/engine"
	"crawler/log"
	"go.uber.org/zap/zapcore"
	"time"
)

// 启动爬虫任务的方式可以分为两种，一种是加载配置文件，另一种是在调用用户接口时，传递任务名称和参数。不过在这里我们先用硬编码的形式来实现。
// 而通过配置文件和用户接口来操作任务的方式我们会有专门的课程来实现。
func main() {
	// log
	plugin := log.NewStdoutPlugin(zapcore.InfoLevel)
	logger := log.NewLogger(plugin)
	logger.Info("log init end")

	// proxy
	//proxyURLs := []string{"http://127.0.0.1:8888", "http://127.0.0.1:8889"}
	//p, err := proxy.RoundRobinProxySwitcher(proxyURLs...)
	//if err != nil {
	//	logger.Error("RoundRobinProxySwitcher failed")
	//}

	var f collect.Fetcher = collect.BrowserFetch{
		Timeout: 3000 * time.Millisecond,
		Logger:  logger,
		//Proxy:   p,
	}

	seeds := make([]*collect.Task, 0, 1000)
	seeds = append(seeds, &collect.Task{
		Name:    "find_douban_sun_room",
		Fetcher: f,
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
