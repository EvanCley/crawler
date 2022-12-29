package main

import (
	"crawler/collect"
	"crawler/proxy"
	"fmt"
	"time"
)

func main() {
	proxyURLs := []string{"http://127.0.0.1:8888", "http://127.0.0.1:8889"}
	p, err := proxy.RoundRobinProxySwitcher(proxyURLs...)
	if err != nil {
		fmt.Println("RoundRobinProxySwitcher failed")
	}
	url := "<https://google.com>"
	var f collect.Fetcher = collect.BrowserFetch{
		Timeout: 3000 * time.Millisecond,
		Proxy:   p,
	}

	body, err := f.Get(url)
	if err != nil {
		fmt.Printf("read content failed: %v\n", err)
		return
	}
	fmt.Println(string(body))
}

// 借助 regexp 标准库，regexp.MustCompile 函数会在编译时提前解析好 PCRE 标准的正则表达式内容，这可以在一定程度上加速程序的运行。
// 在这个规则中，我们要找到以字符串<div class="news_li" 开头，且内部包含 <h2> 和 <a.*?target="_blank"> 的字符串。其中，[\s\S]*? 是这段表达式的精髓，[\s\S] 指代的是任意字符串。
//var headerRe = regexp.MustCompile(`<div class="small_cardcontent__BTALp"[\s\S]*?<h2>([\s\S]*?)</h2>`)
//
//func main() {
//	url := "https://www.thepaper.cn/"
//	pageBytes, err := Fetch(url)
//
//	if err != nil {
//		fmt.Printf("read content failed: %v\n", err)
//		return
//	}
//
//	// headerRe.FindAllSubmatch 是一个三维字节数组 [][][]byte。它的第一层包含的是所有满足正则条件的字符串。
//	// 第二层对每一个满足条件的字符串做了分组。其中，数组的第 0 号元素是满足当前正则表达式的这一串完整的字符串。
//	// 而第 1 号元素代表括号中特定的字符串，在我们这个例子中对应的是 标签括号中的文字，即新闻标题。第三层就是字符串实际对应的字节数组。
//	matches := headerRe.FindAllSubmatch(pageBytes, -1)
//	for _, m := range matches {
//		fmt.Printf("fetch card news: %s\n", string(m[1]))
//	}
//
//}
//
//func Fetch(url string) ([]byte, error) {
//	resp, err := http.Get(url)
//
//	if err != nil {
//		panic(err)
//	}
//
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		fmt.Printf("Error status code: %d", resp.StatusCode)
//	}
//	bodyReader := bufio.NewReader(resp.Body)
//	e := DetermineEncoding(bodyReader)                            // 单独封装了 DeterminEncoding 函数来检测并返回当前 HTML 文本的编码格式
//	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder()) // 最后，transform.NewReader 用于将 HTML 文本从特定编码转换为 UTF-8 编码，从而方便后续的处理。
//	return ioutil.ReadAll(utf8Reader)
//}
//
//func DetermineEncoding(r *bufio.Reader) encoding.Encoding {
//	bytes, err := r.Peek(1024)
//
//	// 如果返回的 HTML 文本小于 1024 字节，我们认为当前 HTML 文本有问题，直接返回默认的 UTF-8 编码就好了。
//	if err != nil {
//		fmt.Printf("fetch error: %v\n", err)
//		return unicode.UTF8
//	}
//	// charset.DetermineEncoding 函数用于检测并返回对应 HTML 文本的编码。
//	e, _, _ := charset.DetermineEncoding(bytes, "")
//	return e
//}
