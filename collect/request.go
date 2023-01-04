package collect

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

type Request struct {
	Task      *Task
	Url       string // URL，表示要访问的网站
	Method    string
	Depth     int // 表示任务的当前深度，最初始的深度为 0
	Priority  int
	RuleName  string
	ParseFunc func([]byte, *Request) ParseResult // ParseFunc 函数会解析从网站获取到的网站信息，并返回 Requests 数组用于进一步获取数据
}

type ParseResult struct {
	Requests []*Request    // 表示要当前网站下接下来要爬取的网站们
	Items    []interface{} // Items 表示获取到的数据
}

// Task 抽象。之前的 Request 结构体会在每一次请求时发生变化，
// 但是我们希望有一个字段能够表示一整个网站的爬取任务，因此我们需要抽离出一个新的结构 Task 作为一个爬虫任务，而 Request 则作为单独的请求存在。
// Task 有些参数是整个任务共有的，例如 Cookie、MaxDepth（最大深度）、WaitTime（默认等待时间）和 RootReq（任务中的第一个请求）。
type Task struct {
	Name        string // 用户界面显示的名称（作为一个任务唯一的标识，应保证唯一性）
	Url         string
	Cookie      string
	WaitTime    time.Duration
	MaxDepth    int  // 为了防止访问陷入到死循环，同时控制爬取的有效链接的数量，一般会给当前任务设置一个最大爬取深度。最大爬取深度是和任务有关的
	Reload      bool // 网站是否可以重复爬取
	Visited     map[string]bool
	VisitedLock sync.Mutex
	Fetcher     Fetcher
	Rule        RuleTree // 规则条件，其中 Root 生成了初始化的爬虫任务
}

// Context 为自定义结构体，用于传递上下文信息，也就是当前的请求参数以及要解析的内容字节数组。后续还会添加请求中的临时数据等上下文数据。
type Context struct {
	Body []byte   // 要解析的内容字节数组
	Req  *Request // 当前的请求参数
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
