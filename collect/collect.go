package collect

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Fetcher 接口，内部有一个方法签名 Get，参数为网站的 URL
type Fetcher interface {
	Get(url string) ([]byte, error)
}

// BaseFetch 用最基本的爬取逻辑实现 Fetcher 接口
type BaseFetch struct {}

func (BaseFetch) Get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error status code:%d", resp.StatusCode)
	}
	bodyReader := bufio.NewReader(resp.Body)
	e := DeterminEncoding(bodyReader)
	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())

	return ioutil.ReadAll(utf8Reader)
}