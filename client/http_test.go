package client

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

var (
	httpConf = Config{
		SkipSSL:           false,
		Threads:           1,
		UserAgent:         "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0",
		RandomDelay:       2,
		DisableKeepAlives: false,
		Proxys:            nil,
		RequestTimeout:    time.Duration(20 * time.Second),
		MaxQueueSize:      10,
	}
)

func open_url(client *CollectorClient, url string, x, y int, t map[[2]int]int) {
	u := fmt.Sprintf(url, x, y)
	_, data := client.Open(u, nil)

	z, err := strconv.Atoi(string(data))

	if err != nil {
		return
	}

	if z != x+y {
		return
	}

	t[[2]int{x, y}] = z
}

func TestHttpFetch(t *testing.T) {
	client := NewCollectorClient(&httpConf, nil)
	result := make(map[[2]int]int)

	for i := 0; i < 100; i++ {
		open_url(client, "http://127.0.0.1:8001/mock/%d/%d.add", i, 29, result)
	}

	if result == nil {
		t.FailNow()
	}
}
