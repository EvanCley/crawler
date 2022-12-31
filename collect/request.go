package collect

import (
	"errors"
	"time"
)

type Request struct {
	Url       string // URL，表示要访问的网站
	WaitTime  time.Duration
	Cookie    string
	Depth     int                                // 表示任务的当前深度，最初始的深度为 0
	MaxDepth  int                                // 为了防止访问陷入到死循环，同时控制爬取的有效链接的数量，一般会给当前任务设置一个最大爬取深度。最大爬取深度是和任务有关的
	ParseFunc func([]byte, *Request) ParseResult // ParseFunc 函数会解析从网站获取到的网站信息，并返回 Requests 数组用于进一步获取数据
}

type ParseResult struct {
	Requests []*Request    // 表示要当前网站下接下来要爬取的网站们
	Items    []interface{} // Items 表示获取到的数据
}

// Check 判断爬虫的当前深度是否超过了最大深度
func (r *Request) Check() error {
	if r.Depth > r.MaxDepth {
		return errors.New("Max depth limit reached")
	}
	return nil
}
