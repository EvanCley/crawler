package doubangroup

import (
	"crawler/collect"
	"regexp"
)

// 获取所有帖子的 URL，这里选择使用正则表达式的方式来实现
const cityListRe = `(<https://www.douban.com/group/topic/[0-9a-z]+/>)"[^>]*>([^<]+)</a>`

func ParseURL(contents []byte, req *collect.Request) collect.ParseResult {
	re := regexp.MustCompile(cityListRe)

	matches := re.FindAllSubmatch(contents, -1)
	result := collect.ParseResult{}

	for _, m := range matches {
		u := string(m[1])
		result.Requests = append(result.Requests, &collect.Request{
			Url:      u,
			WaitTime: req.WaitTime,
			Cookie:   req.Cookie,
			Depth:    req.Depth + 1, // 将 Depth 加 1，这样就标识了下一层的深度
			MaxDepth: req.MaxDepth,
			ParseFunc: func(c []byte, request *collect.Request) collect.ParseResult {
				return GetContent(c, u)
			},
		})
	}
	return result
}

// 新的 Request 需要有不同的解析规则，这里我们想要获取的是正文中带有“阳台”字样的帖子（注意不要匹配到侧边栏的文字）。
// 查看 HTML 文本的规则会发现，正本包含在 <div class="topic-content">xxxx <div> 当中
const ContentRe = `<div class="topic-content">[\s\S]*?<div`

func GetContent(contents []byte, url string) collect.ParseResult {
	re := regexp.MustCompile(ContentRe)

	ok := re.Match(contents)
	if !ok {
		return collect.ParseResult{
			Items: []interface{}{},
		}
	}

	result := collect.ParseResult{
		Items: []interface{}{url},
	}

	return result
}

// 最后在 main 函数中，为了找到所有符合条件的帖子，使用了广度优先搜索算法。循环往复遍历 worklist 列表，完成爬取与解析的动作，找到所有符合条件的帖子。
