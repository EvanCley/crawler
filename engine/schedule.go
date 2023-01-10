package engine

import (
	"crawler/collect"
	"crawler/parse/doubangroup"
	"crawler/storage"
	"go.uber.org/zap"
	"sync"
)

// Crawler 构建一个新的结构体作为全局的爬取实例，将之前 Schedule 中的 options 迁移到 Crawler 中
type Crawler struct {
	out         chan collect.ParseResult    // out 负责处理爬取后的数据，完成下一步的存储操作
	Visited     map[string]bool             // 用一个哈希表结构来存储历史请求。将唯一标识 key 设置为 URL + method 方法，并使用 MD5 生成唯一键
	VisitedLock sync.Mutex                  // 增加 VisitedLock 来确保并发安全
	failures    map[string]*collect.Request // 失败请求 id -> 失败请求
	failureLock sync.Mutex                  // failureLock 互斥锁用于并发安全
	options
}

// Scheduler 只处理与调度有关的工作，并抽象为了 Scheduler 接口。调度器抽象为接口后，如果我们有其他的调度器算法实现，也能够非常方便完成替换了。
type Scheduler interface {
	Schedule()                // Schedule 方法负责启动调度器
	Push(...*collect.Request) // Push 方法会将请求放入到调度器中
	Pull() *collect.Request   // Pull 方法则会从调度器中获取请求
}

type Schedule struct {
	requestCh   chan *collect.Request // requestCh 负责接收请求
	workerCh    chan *collect.Request // workerCh 负责分配任务给 worker
	priReqQueue []*collect.Request    // 在调度函数 Schedule 中，我们会优先从优先队列中获取请求。而在放入请求时，如果请求的优先级更高，也会单独放入优先级队列。
	reqQueue    []*collect.Request
	Logger      *zap.Logger
}

var Store = &CrawlerStore{
	list: []*collect.Task{},
	hash: map[string]*collect.Task{},
}

func GetFields(taskName string, ruleName string) []string {
	return Store.hash[taskName].Rule.Trunk[ruleName].ItemFields
}

type CrawlerStore struct {
	list []*collect.Task
	hash map[string]*collect.Task
}

func init() {
	Store.Add(doubangroup.DoubanGroupTask)
}

func (cs *CrawlerStore) Add(task *collect.Task) {
	cs.hash[task.Name] = task
	cs.list = append(cs.list, task)
}

func NewEngine(opts ...Option) *Crawler {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	e := &Crawler{
		Visited: make(map[string]bool, 100),     // 哈希表需要用 make 进行初始化，要不然在运行时访问哈希表会直接报错。
		out:     make(chan collect.ParseResult), // channel 也需要初始化
	}
	e.options = options
	return e
}

func NewSchedule() *Schedule {
	s := &Schedule{}
	requestCh := make(chan *collect.Request)
	workerCh := make(chan *collect.Request)
	s.requestCh = requestCh
	s.workerCh = workerCh
	return s
}

func (s *Schedule) Push(reqs ...*collect.Request) {
	for _, req := range reqs {
		s.requestCh <- req
	}
}

func (s *Schedule) Pull() *collect.Request {
	r := <-s.workerCh
	return r
}

func (s *Schedule) Output() *collect.Request {
	r := <-s.workerCh
	return r
}

// Schedule 方法就是将 requestCh 中的请求扔进 workerCh 中
func (s *Schedule) Schedule() {
	var req *collect.Request
	var ch chan *collect.Request
	for { // 使用了 for 语句，让调度器循环往复地获取外界的爬虫任务，并将任务分发到 worker 中。
		if req == nil && len(s.priReqQueue) > 0 { // 优先考虑优先级队列中的任务
			req = s.priReqQueue[0]
			s.priReqQueue = s.priReqQueue[1:]
			ch = s.workerCh
		}
		if req == nil && len(s.reqQueue) > 0 { // 如果任务队列 reqQueue 大于 0，意味着有爬虫任务，这时我们获取队列中第一个任务，并将其剔除出队列
			req = s.reqQueue[0]
			s.reqQueue = s.reqQueue[1:]
			ch = s.workerCh
		}

		select { // REVIEW 这里如果 select 选择第一个分支，怎么保证当前这一轮循环的 req 进入到 Ch
		case r := <-s.requestCh: // 若没有取到任务，则将 requestCh 通道接收到的外界请求，存储到 reqQueue 队列中
			if r.Priority > 0 { // 在放入请求时，如果请求的优先级更高，也会单独放入优先级队列。
				s.priReqQueue = append(s.priReqQueue, r)
			} else {
				s.reqQueue = append(s.reqQueue, r)
			}
		case ch <- req: // 最后 ch <- req 会将任务发送到 workerCh 通道中，等待 worker 接收。
			req = nil
			ch = nil
		}
	}
}

func (e *Crawler) Schedule() {
	var reqs []*collect.Request
	for _, seed := range e.Seeds {
		task := Store.hash[seed.Name]
		task.Fetcher = seed.Fetcher
		task.Storage = seed.Storage
		task.Limiter = seed.Limiter
		task.Logger = seed.Logger
		// 在调度器启动时，通过 task.Rule.Root() 获取初始化任务，并加入到任务队列中。
		rootReqs, err := task.Rule.Root()
		if err != nil {
			e.Logger.Error("get root failed", zap.Error(err))
			continue
		}
		reqs = append(reqs, rootReqs...)
	}
	go e.scheduler.Schedule()
	go e.scheduler.Push(reqs...)
}

func (e *Crawler) Run() {
	// schedule 函数会创建调度程序，负责的是调度的核心逻辑。
	go e.Schedule()
	// 下一步，创建指定数量的 worker，完成实际任务的处理。其中 WorkCount 为执行任务的数量，可以灵活地去配置。
	for i := 0; i < e.WorkCount; i++ {
		go e.CreateWork()
	}
	e.HandleResult()
}

// CreateWork 创建出实际处理任务的函数
func (e *Crawler) CreateWork() {
	for {
		req := e.scheduler.Pull()           // 接收到调度器分配的任务；
		if err := req.Check(); err != nil { // 判断当前请求是否达到最大深度
			e.Logger.Error("depth check failed", zap.Error(err))
			continue
		}
		if !req.Task.Reload && e.HasVisited(req) { // 判断当前请求是否已被访问
			e.Logger.Debug("request has visited", zap.String("url:", req.Url))
			continue
		}
		e.StoreVisited(req)      // 设置当前请求已被访问
		body, err := req.Fetch() // 访问服务器
		if err != nil {
			e.Logger.Error("can't fetch ", zap.Error(err), zap.String("url", req.Url))
			e.SetFailure(req) // 设置请求失败
			continue
		}
		if len(body) < 6000 {
			e.Logger.Error("can't fetch ", zap.Int("length", len(body)), zap.String("url", req.Url))
			e.SetFailure(req)
			continue
		}
		rule := req.Task.Rule.Trunk[req.RuleName]
		result, err := rule.ParseFunc(&collect.Context{Body: body, Req: req}) // 解析服务器返回的数据
		if err != nil {
			e.Logger.Error("ParseFunc failed", zap.Error(err), zap.String("url", req.Url))
		}
		if len(result.Requests) > 0 {
			go e.scheduler.Push(result.Requests...)
		}

		e.out <- result // 将返回的数据发送到 out 通道中，方便后续的处理
	}
}

func (e *Crawler) HandleResult() {
	for {
		select {
		case result := <-e.out: // 接收所有 worker 解析后的数据
			for _, req := range result.Requests { // 其中要进一步爬取的 Requests 列表将全部发送回 s.requestCh 通道
				e.scheduler.Push(req)
			}
			// 循环遍历 Items，判断其中的数据类型，如果数据类型为 DataCell，我们就要用专门的存储引擎将这些数据存储起来。
			// 存储引擎是和每一个爬虫任务绑定在一起的，不同的爬虫任务可能会有不同的存储引擎。
			for _, item := range result.Items { // result.Items 里包含了我们实际希望得到的结果
				switch d := item.(type) {
				case *storage.DataCell:
					name := d.GetTaskName()
					task := Store.hash[name]
					task.Storage.Save(d)
				}
				// 用日志把结果打印出来
				e.Logger.Sugar().Info("get result", item)
			}
		}
	}
}

func (e *Crawler) HasVisited(r *collect.Request) bool {
	e.VisitedLock.Lock()
	defer e.VisitedLock.Unlock()
	unique := r.Unique()
	return e.Visited[unique]
}

func (e *Crawler) StoreVisited(reqs ...*collect.Request) {
	e.VisitedLock.Lock()
	defer e.VisitedLock.Unlock()

	for _, r := range reqs {
		unique := r.Unique()
		e.Visited[unique] = true
	}
}

// SetFailure 当请求失败之后，调用 SetFailure 方法将请求加入到 failures 哈希表中，并且把它重新交由调度引擎进行调度。
func (e *Crawler) SetFailure(req *collect.Request) {
	if !req.Task.Reload { // 如果网站不允许重复爬取，则从 Visited map 中删除当前 req
		e.VisitedLock.Lock()
		unique := req.Unique()
		delete(e.Visited, unique)
		e.VisitedLock.Unlock()
	}
	e.VisitedLock.Lock()
	defer e.VisitedLock.Unlock()
	// 首次失败时，再重新执行一次
	if _, ok := e.failures[req.Unique()]; ok {
		e.failures[req.Unique()] = req
		e.scheduler.Push(req)
	}
	// TODO: 失败两次，加载到失败队列中
}
