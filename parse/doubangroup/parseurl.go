package doubangroup

import (
	"crawler/collect"
	"fmt"
	"regexp"
	"time"
)

// 当前任务规则包括了“解析网站 URL” 和“解析阳台房” 这两个规则，分别对应了处理函数 ParseURL 和 GetSunRoom
const urlListRe = `(https://www.douban.com/group/topic/[0-9a-z]+/)"[^>]*>([^<]+)</a>`
const ContentRe = `<div class="topic-content">[\s\S]*?阳台[\s\S]*?<div class="aside">`

var DoubanGroupTask = &collect.Task{
	Property: collect.Property{
		Name:     "find_douban_sun_romm",
		WaitTime: 1 * time.Second,
		MaxDepth: 5,
		Cookie:   "xxx",
	},
	Rule: collect.RuleTree{
		Root: func() ([]*collect.Request, error) {
			var roots []*collect.Request
			for i := 0; i < 25; i += 25 {
				str := fmt.Sprintf("<https://www.douban.com/group/szsh/discussion?start=%d>", i)
				roots = append(roots, &collect.Request{
					Priority: 1,
					Url:      str,
					Method:   "Get",
					RuleName: "解析网站URL",
				})
			}
			return roots, nil
		},
		Trunk: map[string]*collect.Rule{
			"解析网站URL": &collect.Rule{ParseFunc: ParseURL},
			"解析阳台房":   &collect.Rule{ParseFunc: GetSunRoom},
		},
	},
}

func ParseURL(ctx *collect.Context) (collect.ParseResult, error) {
	re := regexp.MustCompile(urlListRe)

	matches := re.FindAllSubmatch(ctx.Body, -1)
	result := collect.ParseResult{}

	for _, m := range matches {
		u := string(m[1])
		result.Requests = append(result.Requests, &collect.Request{
			Method:   "GET",
			Task:     ctx.Req.Task,
			Url:      u,
			Depth:    ctx.Req.Depth + 1, // 将 Depth 加 1，这样就标识了下一层的深度
			RuleName: "解析阳台房",
		})
	}
	return result, nil
}

func GetSunRoom(ctx *collect.Context) (collect.ParseResult, error) {
	re := regexp.MustCompile(ContentRe)

	ok := re.Match(ctx.Body)
	if !ok {
		return collect.ParseResult{Items: []interface{}{}}, nil
	}

	result := collect.ParseResult{
		Items: []interface{}{ctx.Req.Url},
	}

	return result, nil
}

// 最后在 main 函数中，为了找到所有符合条件的帖子，使用了广度优先搜索算法。循环往复遍历 worklist 列表，完成爬取与解析的动作，找到所有符合条件的帖子。
