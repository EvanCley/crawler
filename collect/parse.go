package collect

// RuleTree 规则引擎就像一棵树。每一个规则就是一个 ParseFunc 解析函数。
type RuleTree struct {
	Root  func() ([]*Request, error) // 根节点(执行入口)。RuleTree.Root 是一个函数，用于生成爬虫的种子网站
	Trunk map[string]*Rule           // 规则哈希表，用于存储当前任务所有的规则，哈希表的 Key 为规则名，Value 为具体的规则。
}

type Rule struct { // 每一个规则就是一个 ParseFunc 解析函数。
	ItemFields []string
	ParseFunc  func(*Context) (ParseResult, error) // 内容解析函数。ParseFunc 函数会解析从网站获取到的网站信息，并返回 Requests 数组用于进一步获取数据
}
