package collect

type Request struct {
	Url       string // URL，表示要访问的网站
	Cookie    string
	ParseFunc func([]byte) ParseResult // ParseFunc 函数会解析从网站获取到的网站信息，并返回 Requests 数组用于进一步获取数据
}

type ParseResult struct {
	Requests []*Request    // 表示要当前网站下接下来要爬取的网站们
	Items    []interface{} // Items 表示获取到的数据
}
