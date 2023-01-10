package doubanbook

import (
	"crawler/collect"
	"regexp"
	"strconv"
	"time"
)

var DoubanBookTask = &collect.Task{
	Property: collect.Property{
		Name:     "douban_book_list",
		WaitTime: 1 * time.Second,
		MaxDepth: 5,
		Cookie:   "gr_user_id=63380bb1-6e3f-4d56-aa90-9c6a7b0f102d; douban-fav-remind=1; _pk_id.100001.8cb4=d9738ae7115c7fbc.1574325496.11.1616726158.1614739288.; __utma=30149280.614434265.1574252090.1614739288.1616726158.14; viewed=\"35100082_26854226_11589828_35130972_35424872_26997846_26894736_26632674_1929984_35217981\"; bid=SOlBF_IZqnY; ap_v=0,6.0",
	},
	Rule: collect.RuleTree{
		Root: func() ([]*collect.Request, error) {
			roots := []*collect.Request{
				&collect.Request{
					Priority: 1,
					Url:      "https://book.douban.com",
					Method:   "GET",
					RuleName: "数据tag",
				},
			}
			return roots, nil
		},
		Trunk: map[string]*collect.Rule{
			"数据tag": &collect.Rule{ParseFunc: ParseTag},
			"数据列表":  &collect.Rule{ParseFunc: ParseBookList},
			"书籍简介": &collect.Rule{
				ParseFunc:  ParseBookDetail,
				ItemFields: []string{"书名", "作者", "页数", "出版社", "得分", "价格", "简介"},
			},
		},
	},
}

const regexpStr = `<a href="([^"]+)" class="tag">([^<]+)</a>`

func ParseTag(ctx *collect.Context) (collect.ParseResult, error) {
	re := regexp.MustCompile(regexpStr)

	matches := re.FindAllSubmatch(ctx.Body, -1)

	result := collect.ParseResult{}
	for _, m := range matches {
		result.Requests = append(result.Requests, &collect.Request{
			Method:   "GET",
			Task:     ctx.Req.Task,
			Url:      "<https://book.douban.com>" + string(m[1]),
			Depth:    ctx.Req.Depth + 1,
			RuleName: "数据列表",
		})
	}
	return result, nil
}

const BookListRe = `<a.*?href="([^"]+)" title="([^"]+)"`

func ParseBookList(ctx *collect.Context) (collect.ParseResult, error) {
	re := regexp.MustCompile(BookListRe)
	matches := re.FindAllSubmatch(ctx.Body, -1)
	result := collect.ParseResult{}
	for _, m := range matches {
		req := &collect.Request{
			Method:   "GET",
			Task:     ctx.Req.Task,
			Url:      string(m[1]),
			Depth:    ctx.Req.Depth + 1,
			RuleName: "书籍简介",
		}
		// 这里获取到书名之后，将书名缓存到了临时的 tmp 结构中供下一个阶段读取。这是因为我们希望得到的某些信息是在之前的阶段获得的。
		// 在这里将缓存结构定义为了一个哈希表，并封装了 Get 与 Set 两个函数来获取和设置请求中的缓存。
		req.TmpData = &collect.Temp{}
		req.TmpData.Set("book_name", string(m[2]))
		result.Requests = append(result.Requests, req)
	}

	return result, nil
}

var autoRe = regexp.MustCompile(`<span class="pl"> 作者</span>:[\d\D]*?<a.*?>([^<]+)</a>`)
var public = regexp.MustCompile(`<span class="pl">出版社:</span>([^<]+)<br/>`)
var pageRe = regexp.MustCompile(`<span class="pl">页数:</span> ([^<]+)<br/>`)
var priceRe = regexp.MustCompile(`<span class="pl">定价:</span>([^<]+)<br/>`)
var scoreRe = regexp.MustCompile(`<strong class="ll rating_num " property="v:average">([^<]+)</strong>`)
var intoRe = regexp.MustCompile(`<div class="intro">[\d\D]*?<p>([^<]+)</p></div>`)

func ParseBookDetail(ctx *collect.Context) (collect.ParseResult, error) {
	bookName := ctx.Req.TmpData.Get("book_name") // 其中，书名是从缓存中得到的。
	page, _ := strconv.Atoi(ExtraString(ctx.Body, pageRe))

	book := map[string]interface{}{
		"书名":  bookName,
		"作者":  ExtraString(ctx.Body, autoRe),
		"页数":  page,
		"出版社": ExtraString(ctx.Body, public),
		"得分":  ExtraString(ctx.Body, scoreRe),
		"价格":  ExtraString(ctx.Body, priceRe),
		"简介":  ExtraString(ctx.Body, intoRe),
	}
	data := ctx.Output(book)

	result := collect.ParseResult{
		Items: []interface{}{data},
	}

	return result, nil
}

func ExtraString(contents []byte, re *regexp.Regexp) string {

	match := re.FindSubmatch(contents)

	if len(match) >= 2 {
		return string(match[1])
	} else {
		return ""
	}
}
