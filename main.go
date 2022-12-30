package main

import (
	"crawler/collect"
	"crawler/log"
	"crawler/parse/doubangroup"
	"crawler/proxy"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

func main() {
	// log
	plugin := log.NewStdoutPlugin(zapcore.InfoLevel)
	logger := log.NewLogger(plugin)
	logger.Info("log init end")

	// proxy
	proxyURLs := []string{"http://127.0.0.1:8888", "http://127.0.0.1:8889"}
	p, err := proxy.RoundRobinProxySwitcher(proxyURLs...)
	if err != nil {
		logger.Error("RoundRobinProxySwitcher failed")
	}

	// douban cookie
	cookie := "gr_user_id=63380bb1-6e3f-4d56-aa90-9c6a7b0f102d; douban-fav-remind=1; _pk_id.100001.8cb4=d9738ae7115c7fbc.1574325496.11.1616726158.1614739288.; __utma=30149280.614434265.1574252090.1614739288.1616726158.14; viewed=\"35100082_26854226_11589828_35130972_35424872_26997846_26894736_26632674_1929984_35217981\"; bid=SOlBF_IZqnY; ap_v=0,6.0"
	var worklist []*collect.Request
	for i := 0; i <= 100; i += 25 {
		str := fmt.Sprintf("<http://www.douban.com/group/szsh/discussion?start=%d>", i)
		worklist = append(worklist, &collect.Request{
			Url:       str,
			Cookie:    cookie,
			ParseFunc: doubangroup.ParseURL,
		})
	}

	var f collect.Fetcher = collect.BrowserFetch{
		Timeout: 3000 * time.Millisecond,
		Proxy:   p,
	}

	for len(worklist) > 0 {
		items := worklist
		worklist = nil
		for _, item := range items {
			body, err := f.Get(item)
			time.Sleep(1 * time.Second)
			if err != nil {
				logger.Error("read content failed", zap.Error(err))
				continue
			}
			res := item.ParseFunc(body)
			for _, item := range res.Items {
				logger.Info("result", zap.String("get url:", item.(string)))
			}
			worklist = append(worklist, res.Requests...)
		}
	}
}
