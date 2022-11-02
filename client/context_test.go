package client

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestCollectorContext(t *testing.T) {
	conf := &Config{SkipSSL: false, Threads: 1, UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36", RandomDelay: 0, DisableKeepAlives: false, RequestTimeout: 5 * time.Second}

	client := NewCollectorClient(conf, nil)

	code, data := client.Open("https://api.luokuang.com/emg/v2/map/tile?ak=DE16394725636226419F0AB3765FFA4D2EB27A998BD259465DGXZ8TJUZKE2067&format=pbf&layer=basic&sku=101XxiLvoFYxL&style=main&x=6778&y=3194&zoom=13", nil, nil)

	if code != 200 || data == nil {
		t.FailNow()
	}

	ioutil.WriteFile("./test.mvt", data, os.ModePerm)
}
