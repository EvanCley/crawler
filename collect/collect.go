package collect

import (
	"bufio"
	"crawler/proxy"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/http"
	"time"
)

// Fetcher 接口，内部有一个方法签名 Get，参数为网站的 URL
type Fetcher interface {
	Get(request *Request) ([]byte, error)
}

// BaseFetch 用最基本的爬取逻辑实现 Fetcher 接口
type BaseFetch struct{}

func (BaseFetch) Get(request *Request) ([]byte, error) {
	resp, err := http.Get(request.Url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error status code:%d", resp.StatusCode)
	}
	bodyReader := bufio.NewReader(resp.Body)
	e := DetermineEncoding(bodyReader)
	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())

	return ioutil.ReadAll(utf8Reader)
}

// 模拟浏览器访问
type BrowserFetch struct {
	Timeout time.Duration // 增加 Timeout 超时参数，进行超时控制
	Logger  *zap.Logger
	Proxy   proxy.ProxyFunc
}

func (b BrowserFetch) Get(request *Request) ([]byte, error) {
	client := &http.Client{
		Timeout: b.Timeout,
	}
	if b.Proxy != nil { // 更新 http.Client 变量中的 Transport 结构中的 Proxy 函数，将其替换为我们自定义的代理函数。
		// 在 Go http 标准库中，默认 Transport 为 http.DefaultTransport ，它定义了包括超时时间在内的诸多默认参数，并且实现了一个默认的 Proxy 函数 ProxyFromEnvironment。
		transport := http.DefaultTransport.(*http.Transport)
		transport.Proxy = b.Proxy
		client.Transport = transport
	}

	req, err := http.NewRequest("GET", request.Url, nil)
	if err != nil {
		return nil, fmt.Errorf("get url failed: %v", err)
	}

	if len(request.Cookie) > 0 {
		req.Header.Set("Cookie", request.Cookie)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/108.0.1462.46")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	fmt.Println("resp content", resp)
	bodyReader := bufio.NewReader(resp.Body)
	fmt.Println("bodyReader", bodyReader)
	e := DetermineEncoding(bodyReader)
	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())

	return ioutil.ReadAll(utf8Reader)
}

func DetermineEncoding(r *bufio.Reader) encoding.Encoding {
	bytes, err := r.Peek(1024)
	// 如果返回的 HTML 文本小于 1024 字节，我们认为当前 HTML 文本有问题，直接返回默认的 UTF-8 编码就好了。
	if err != nil {
		fmt.Printf("fetch error: %v\n", err)
		return unicode.UTF8
	}
	// charset.DetermineEncoding 函数用于检测并返回对应 HTML 文本的编码。
	e, _, _ := charset.DetermineEncoding(bytes, "")
	return e
}
