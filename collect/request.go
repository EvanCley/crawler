package collect

import (
	"context"
	"crawler/limiter"
	"crawler/storage"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"go.uber.org/zap"
	"math/rand"
	"sync"
	"time"
)

type ParseResult struct {
	Requests []*Request    // 表示要当前网站下接下来要爬取的网站们
	Items    []interface{} // Items 表示获取到的数据。将最终的数据存储到 Items 中，供我们专门的协程去处理了。
}

// Task 抽象，表示一个任务实例。之前的 Request 结构体会在每一次请求时发生变化，
// 但是我们希望有一个字段能够表示一整个网站的爬取任务，因此我们需要抽离出一个新的结构 Task 作为一个爬虫任务，而 Request 则作为单独的请求存在。
// Task 有些参数是整个任务共有的，例如 Cookie、MaxDepth（最大深度）、WaitTime（默认等待时间）和 RootReq（任务中的第一个请求）。
type Task struct {
	Property
	Visited     map[string]bool
	VisitedLock sync.Mutex
	Fetcher     Fetcher
	Storage     storage.Storage
	Rule        RuleTree // 规则条件，其中 Root 生成了初始化的爬虫任务
	Logger      *zap.Logger
	Limiter     limiter.RateLimiter
}

type Property struct {
	Name     string // 用户界面显示的名称（作为一个任务唯一的标识，应保证唯一性）
	Url      string
	Cookie   string
	WaitTime time.Duration
	MaxDepth int  // 为了防止访问陷入到死循环，同时控制爬取的有效链接的数量，一般会给当前任务设置一个最大爬取深度。最大爬取深度是和任务有关的
	Reload   bool // 网站是否可以重复爬取
}

// Context 为自定义结构体，用于传递上下文信息，也就是当前的请求参数以及要解析的内容字节数组。后续还会添加请求中的临时数据等上下文数据。
type Context struct {
	Body []byte   // 要解析的内容字节数组
	Req  *Request // 当前的请求参数
}

func (c *Context) GetRule(ruleName string) *Rule {
	return c.Req.Task.Rule.Trunk[ruleName]
}

func (c *Context) Output(data interface{}) *storage.DataCell {
	// 定义“Data”对应的数据结构又是一个哈希表 map[string]interface{}。在这个哈希表中，Key 为“书名”“评分”等字段名，Value 为字段对应的值。
	res := &storage.DataCell{}
	res.Data = map[string]interface{}{
		"Task": c.Req.Task.Name,
		"Rule": c.Req.RuleName,
		"Data": data,
		"Url":  c.Req.Url,
		"Time": time.Now().Format("2006-01-02 15:04:05"),
	}
	return res
}

// Request 单个请求
type Request struct {
	Task     *Task
	Url      string // URL，表示要访问的网站
	Method   string
	Depth    int // 表示任务的当前深度，最初始的深度为 0
	Priority int
	RuleName string
	TmpData  *Temp
	//ParseFunc func([]byte, *Request) ParseResult // ParseFunc 函数会解析从网站获取到的网站信息，并返回 Requests 数组用于进一步获取数据
}

func (r *Request) Fetch() ([]byte, error) {
	if err := r.Task.Limiter.Wait(context.Background()); err != nil {
		return nil, err
	}
	// 随机休眠，模拟人类行为
	sleepTime := rand.Int63n(int64(r.Task.WaitTime * 1000))
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
	return r.Task.Fetcher.Get(r)
}

// Check 判断爬虫的当前深度是否超过了最大深度
func (r *Request) Check() error {
	if r.Depth > r.Task.MaxDepth {
		return errors.New("Max depth limit reached")
	}
	return nil
}

// Unique 请求的唯一识别码
func (r *Request) Unique() string {
	block := md5.Sum([]byte(r.Url + r.Method))
	return hex.EncodeToString(block[:])
}
