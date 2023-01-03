package main

import (
	"crawler/collect"
	"crawler/engine"
	"crawler/log"
	"crawler/parse/doubangroup"
	"fmt"
	"go.uber.org/zap/zapcore"
	"time"
)

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

	// douban cookie
	var seeds []*collect.Task
	for i := 0; i <= 0; i += 25 {
		str := fmt.Sprintf("https://www.douban.com/group/szsh/discussion?start=%d", i)
		seeds = append(seeds, &collect.Task{ // 生成初始网址列表作为种子任务
			Url:      str,
			WaitTime: 1 * time.Second,
			Cookie:   "gr_user_id=63380bb1-6e3f-4d56-aa90-9c6a7b0f102d; douban-fav-remind=1; _pk_id.100001.8cb4=d9738ae7115c7fbc.1574325496.11.1616726158.1614739288.; __utma=30149280.614434265.1574252090.1614739288.1616726158.14; viewed=\"35100082_26854226_11589828_35130972_35424872_26997846_26894736_26632674_1929984_35217981\"; bid=SOlBF_IZqnY",
			RootReq:  &collect.Request{ParseFunc: doubangroup.ParseURL},
		})
	}

	var f collect.Fetcher = collect.BrowserFetch{
		Timeout: 3000 * time.Millisecond,
		Logger:  logger,
		//Proxy:   p,
	}

	s := engine.NewEngine(
		engine.WithWorkCount(5),
		engine.WithFetcher(f),
		engine.WithLogger(logger),
		engine.WithSeeds(seeds),
	)
	s.Run()
}
