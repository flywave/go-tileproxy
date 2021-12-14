package client

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/flywave/go-tileproxy/utils"
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

func Get(url string) []byte {

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}

	return result.Bytes()
}

func get_url(client *CollectorClient, url string, font, name string, i int) {
	if !utils.FileExists(fmt.Sprintf("%s/%s", font, name)) {
		data := Get(url)

		ioutil.WriteFile(fmt.Sprintf("%s/%s", font, name), data, os.ModePerm)
	}
}

func TestFetchFonts(t *testing.T) {
	client := NewCollectorClient(&httpConf, nil)

	fonts := []string{
		"DINOffcProRegularNotoSansCJKSCDemiLightArialUnicodeMSRegular",
		"DINOffcProMediumNotoSansCJKSCMediumArialUnicodeMSRegular",
		"Noto Sans CJK SC DemiLight,Arial Unicode MS Regular",
		"NotoSansCJKSCDemiLightArialUnicodeMSRegular",
		"NotoSansCJKSCMediumArialUnicodeMSRegular",
		"NotoSansCJKSCMediumArialUnicodeMSBold",
	}

	for _, font := range fonts {
		for i := 19968; i < 65535; i += 256 {
			url := fmt.Sprintf("https://api.luokuang.com/emg/fonts/%s/%d-%d.pbf?ak=DB1589338764061116C2A51A2CF1B144F2BFA1D29334859EE996VY8OQZR7MGD8", font, i, i+255)
			get_url(client, url, font, fmt.Sprintf("%d-%d.pbf", i, i+255), i)
		}
	}

	if len(testResult) == 0 {
		t.FailNow()
	}
}
