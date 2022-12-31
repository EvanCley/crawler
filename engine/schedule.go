package engine

import (
	"crawler/collect"
	"go.uber.org/zap"
)

type Schedule struct {
	requestCh chan *collect.Request    // requestCh 负责接收请求
	workerCh  chan *collect.Request    // workerCh 负责分配任务给 worker
	out       chan collect.ParseResult // out 负责处理爬取后的数据，完成下一步的存储操作
	options
}

func NewSchedule(opts ...Option) *Schedule {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	s := &Schedule{}
	s.options = options
	return s
}

func (s *Schedule) Run() {
	requestCh := make(chan *collect.Request)
	workerCh := make(chan *collect.Request)
	out := make(chan collect.ParseResult)
	s.requestCh = requestCh
	s.workerCh = workerCh
	s.out = out
	// schedule 函数会创建调度程序，负责的是调度的核心逻辑。
	go s.Schedule()
	// 下一步，创建指定数量的 worker，完成实际任务的处理。其中 WorkCount 为执行任务的数量，可以灵活地去配置。
	for i := 0; i < s.WorkCount; i++ {
		go s.CreateWork()
	}
	s.HandleResult()
}

// Schedule 方法就是将 requestCh 中的请求扔进 workerCh 中
func (s *Schedule) Schedule() {
	var reqQueue = s.Seeds
	go func() {
		for { // 使用了 for 语句，让调度器循环往复地获取外界的爬虫任务，并将任务分发到 worker 中。
			var req *collect.Request
			var ch chan *collect.Request

			if len(reqQueue) > 0 { // 如果任务队列 reqQueue 大于 0，意味着有爬虫任务，这时我们获取队列中第一个任务，并将其剔除出队列
				req = reqQueue[0]
				reqQueue = reqQueue[1:]
				ch = s.workerCh
			}
			select {
			case r := <-s.requestCh: // 若没有取到任务，则将 requestCh 通道接收到的外界请求，存储到 reqQueue 队列中
				reqQueue = append(reqQueue, r)
			case ch <- req: // 最后 ch <- req 会将任务发送到 workerCh 通道中，等待 worker 接收。
			}
		}
	}()
}

// CreateWork 创建出实际处理任务的函数
func (s *Schedule) CreateWork() {
	for {
		r := <-s.workerCh // 接收到调度器分配的任务；
		if err := r.Check(); err != nil {
			s.Logger.Error("depth check failed", zap.Error(err))
			continue
		}
		body, err := s.Fetcher.Get(r) // 访问服务器
		if err != nil {
			s.Logger.Error("can't fetch ", zap.Error(err))
			continue
		}
		result := r.ParseFunc(body, r) // 解析服务器返回的数据
		s.out <- result                // 将返回的数据发送到 out 通道中，方便后续的处理
	}
}

func (s *Schedule) HandleResult() {
	for {
		select {
		case result := <-s.out: // 接收所有 worker 解析后的数据
			for _, req := range result.Requests { // 其中要进一步爬取的 Requests 列表将全部发送回 s.requestCh 通道
				s.requestCh <- req
			}
			for _, item := range result.Items { // result.Items 里包含了我们实际希望得到的结果，所以我们先用日志把结果打印出来
				// TODO: store
				s.Logger.Sugar().Info("get result", item)
			}
		}
	}
}
