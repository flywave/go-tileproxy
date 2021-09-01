package client

import (
	"sync"
	"testing"
	"time"
)

var (
	httpConf = Config{
		SkipSSL:           false,
		Threads:           2,
		UserAgent:         "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0",
		RandomDelay:       2,
		DisableKeepAlives: false,
		Proxys:            nil,
		RequestTimeout:    time.Duration(20 * time.Second),
		MaxQueueSize:      1000,
	}

	testUrls = []string{
		"http://news.baidu.com/guonei",
		"http://news.baidu.com/guoji",
		"http://news.baidu.com/mil",
		"http://news.baidu.com/finance",
		"http://news.baidu.com/ent",
		"http://news.baidu.com/sports",
		"http://news.baidu.com/internet",
		"http://news.baidu.com/tech",
		"http://news.baidu.com/game",
		"http://news.baidu.com/lady",
		"http://news.baidu.com/auto",
		"http://news.baidu.com/house",
	}
	testResult = map[string]string{}
)

func open_url(client *CollectorClient, url string, i int, wg *sync.WaitGroup) {
	defer wg.Done()
	_, data := client.Open(url, nil)
	testResult[url] = string(data)
}

func TestHttpFetch(t *testing.T) {
	client := NewCollectorClient(&httpConf, nil)
	var wg sync.WaitGroup

	for i, url := range testUrls {
		wg.Add(1)
		go open_url(client, url, i, &wg)
	}

	wg.Wait()

	if len(testResult) == 0 {
		t.FailNow()
	}
}
