package client

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestCollectorContext(t *testing.T) {
	conf := &Config{SkipSSL: false, Threads: 1, UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36", RandomDelay: 0, DisableKeepAlives: false, RequestTimeout: 5 * time.Second}
	client := NewCollectorClient(conf)

	code, data := client.Open("https://api.mapbox.com/v4/mapbox.mapbox-streets-v8/1/0/0.mvt?access_token=pk.eyJ1IjoiYW5pbmdnbyIsImEiOiJja291c2piaGwwMDYyMm5wbWI1aGl4Y2VjIn0.slAHkiCz89a6ukssQ7lebQ", nil)

	if code != 200 || data == nil {
		t.FailNow()
	}

	ioutil.WriteFile("./test.mvt", data, os.ModePerm)

	os.Remove("./test.mvt")
}
