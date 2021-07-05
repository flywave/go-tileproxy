package crawler

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/flywave/go-tileproxy/debug"
)

var serverIndexResponse = []byte("hello world\n")
var robotsFile = `
User-agent: *
Allow: /allowed
Disallow: /disallowed
Disallow: /allowed*q=
`

func newTestServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(serverIndexResponse)
	})

	mux.HandleFunc("/html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
<title>Test Page</title>
</head>
<body>
<h1>Hello World</h1>
<p class="description">This is a test page</p>
<p class="description">This is a test paragraph</p>
</body>
</html>
		`))
	})

	mux.HandleFunc("/xml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<page>
	<title>Test Page</title>
	<paragraph type="description">This is a test page</paragraph>
	<paragraph type="description">This is a test paragraph</paragraph>
</page>
		`))
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(r.FormValue("name")))
		}
	})

	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(robotsFile))
	})

	mux.HandleFunc("/allowed", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("allowed"))
	})

	mux.HandleFunc("/disallowed", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("disallowed"))
	})

	mux.Handle("/redirect", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redirected/", http.StatusSeeOther)

	}))

	mux.Handle("/redirected/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<a href="test">test</a>`)
	}))

	mux.HandleFunc("/set_cookie", func(w http.ResponseWriter, r *http.Request) {
		c := &http.Cookie{Name: "test", Value: "testv", HttpOnly: false}
		http.SetCookie(w, c)
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/check_cookie", func(w http.ResponseWriter, r *http.Request) {
		cs := r.Cookies()
		if len(cs) != 1 || r.Cookies()[0].Value != "testv" {
			w.WriteHeader(500)
			w.Write([]byte("nok"))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/500", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(500)
		w.Write([]byte("<p>error</p>"))
	})

	mux.HandleFunc("/user_agent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(r.Header.Get("User-Agent")))
	})

	mux.HandleFunc("/base", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
<title>Test Page</title>
<base href="http://xy.com/" />
</head>
<body>
<a href="z">link</a>
</body>
</html>
		`))
	})

	mux.HandleFunc("/base_relative", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
<title>Test Page</title>
<base href="/foobar/" />
</head>
<body>
<a href="z">link</a>
</body>
</html>
		`))
	})

	mux.HandleFunc("/large_binary", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		ww := bufio.NewWriter(w)
		defer ww.Flush()
		for {
			// have to check error to detect client aborting download
			if _, err := ww.Write([]byte{0x41}); err != nil {
				return
			}
		}
	})

	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		i := 0

		for {
			select {
			case <-r.Context().Done():
				return
			case t := <-ticker.C:
				fmt.Fprintf(w, "%s\n", t)
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
				i++
				if i == 10 {
					return
				}
			}
		}
	})

	return httptest.NewServer(mux)
}

var newCollectorTests = map[string]func(*testing.T){
	"UserAgent": func(t *testing.T) {
		for _, ua := range []string{
			"foo",
			"bar",
		} {
			c := NewCollector(UserAgent(ua))

			if got, want := c.UserAgent, ua; got != want {
				t.Fatalf("c.UserAgent = %q, want %q", got, want)
			}
		}
	},
	"AllowedDomains": func(t *testing.T) {
		for _, domains := range [][]string{
			{"example.com", "example.net"},
			{"example.net"},
			{},
			nil,
		} {
			c := NewCollector(AllowedDomains(domains...))

			if got, want := c.AllowedDomains, domains; !reflect.DeepEqual(got, want) {
				t.Fatalf("c.AllowedDomains = %q, want %q", got, want)
			}
		}
	},
	"DisallowedDomains": func(t *testing.T) {
		for _, domains := range [][]string{
			{"example.com", "example.net"},
			{"example.net"},
			{},
			nil,
		} {
			c := NewCollector(DisallowedDomains(domains...))

			if got, want := c.DisallowedDomains, domains; !reflect.DeepEqual(got, want) {
				t.Fatalf("c.DisallowedDomains = %q, want %q", got, want)
			}
		}
	},
	"DisallowedURLFilters": func(t *testing.T) {
		for _, filters := range [][]*regexp.Regexp{
			{regexp.MustCompile(`.*not_allowed.*`)},
		} {
			c := NewCollector(DisallowedURLFilters(filters...))

			if got, want := c.DisallowedURLFilters, filters; !reflect.DeepEqual(got, want) {
				t.Fatalf("c.DisallowedURLFilters = %v, want %v", got, want)
			}
		}
	},
	"URLFilters": func(t *testing.T) {
		for _, filters := range [][]*regexp.Regexp{
			{regexp.MustCompile(`\w+`)},
			{regexp.MustCompile(`\d+`)},
			{},
			nil,
		} {
			c := NewCollector(URLFilters(filters...))

			if got, want := c.URLFilters, filters; !reflect.DeepEqual(got, want) {
				t.Fatalf("c.URLFilters = %v, want %v", got, want)
			}
		}
	},
	"AllowURLRevisit": func(t *testing.T) {
		c := NewCollector(AllowURLRevisit())

		if !c.AllowURLRevisit {
			t.Fatal("c.AllowURLRevisit = false, want true")
		}
	},
	"MaxBodySize": func(t *testing.T) {
		for _, sizeInBytes := range []int{
			1024 * 1024,
			1024,
			0,
		} {
			c := NewCollector(MaxBodySize(sizeInBytes))

			if got, want := c.MaxBodySize, sizeInBytes; got != want {
				t.Fatalf("c.MaxBodySize = %d, want %d", got, want)
			}
		}
	},
	"CacheDir": func(t *testing.T) {
		for _, path := range []string{
			"/tmp/",
			"/var/cache/",
		} {
			c := NewCollector(CacheDir(path))

			if got, want := c.CacheDir, path; got != want {
				t.Fatalf("c.CacheDir = %q, want %q", got, want)
			}
		}
	},
	"IgnoreRobotsTxt": func(t *testing.T) {
		c := NewCollector(IgnoreRobotsTxt())

		if !c.IgnoreRobotsTxt {
			t.Fatal("c.IgnoreRobotsTxt = false, want true")
		}
	},
	"ID": func(t *testing.T) {
		for _, id := range []uint32{
			0,
			1,
			2,
		} {
			c := NewCollector(ID(id))

			if got, want := c.ID, id; got != want {
				t.Fatalf("c.ID = %d, want %d", got, want)
			}
		}
	},
	"DetectCharset": func(t *testing.T) {
		c := NewCollector(DetectCharset())

		if !c.DetectCharset {
			t.Fatal("c.DetectCharset = false, want true")
		}
	},
	"Debugger": func(t *testing.T) {
		d := &debug.LogDebugger{}
		c := NewCollector(Debugger(d))

		if got, want := c.debugger, d; got != want {
			t.Fatalf("c.debugger = %v, want %v", got, want)
		}
	},
	"CheckHead": func(t *testing.T) {
		c := NewCollector(CheckHead())

		if !c.CheckHead {
			t.Fatal("c.CheckHead = false, want true")
		}
	},
	"Async": func(t *testing.T) {
		c := NewCollector(Async())

		if !c.Async {
			t.Fatal("c.Async = false, want true")
		}
	},
}

func TestNewCollector(t *testing.T) {
	t.Run("Functional Options", func(t *testing.T) {
		for name, test := range newCollectorTests {
			t.Run(name, test)
		}
	})
}

func TestCollectorVisit(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	c := NewCollector()

	onRequestCalled := false
	onResponseCalled := false
	onScrapedCalled := false

	c.OnRequest(func(r *Request) {
		onRequestCalled = true
		r.Ctx.Put("x", "y")
	})

	c.OnResponse(func(r *Response) {
		onResponseCalled = true

		if r.Ctx.Get("x") != "y" {
			t.Error("Failed to retrieve context value for key 'x'")
		}

		if !bytes.Equal(r.Body, serverIndexResponse) {
			t.Error("Response body does not match with the original content")
		}
	})

	c.OnScraped(func(r *Response) {
		if !onResponseCalled {
			t.Error("OnScraped called before OnResponse")
		}

		if !onRequestCalled {
			t.Error("OnScraped called before OnRequest")
		}

		onScrapedCalled = true
	})

	c.Visit(ts.URL)

	if !onRequestCalled {
		t.Error("Failed to call OnRequest callback")
	}

	if !onResponseCalled {
		t.Error("Failed to call OnResponse callback")
	}

	if !onScrapedCalled {
		t.Error("Failed to call OnScraped callback")
	}
}

func TestCollectorVisitWithAllowedDomains(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	c := NewCollector(AllowedDomains("localhost", "127.0.0.1", "::1"))
	err := c.Visit(ts.URL)
	if err != nil {
		t.Errorf("Failed to visit url %s", ts.URL)
	}

	err = c.Visit("http://example.com")
	if err != ErrForbiddenDomain {
		t.Errorf("c.Visit should return ErrForbiddenDomain, but got %v", err)
	}
}

func TestCollectorVisitWithDisallowedDomains(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	c := NewCollector(DisallowedDomains("localhost", "127.0.0.1", "::1"))
	err := c.Visit(ts.URL)
	if err != ErrForbiddenDomain {
		t.Errorf("c.Visit should return ErrForbiddenDomain, but got %v", err)
	}

	c2 := NewCollector(DisallowedDomains("example.com"))
	err = c2.Visit("http://example.com:8080")
	if err != ErrForbiddenDomain {
		t.Errorf("c.Visit should return ErrForbiddenDomain, but got %v", err)
	}
	err = c2.Visit(ts.URL)
	if err != nil {
		t.Errorf("Failed to visit url %s", ts.URL)
	}
}

func TestCollectorVisitResponseHeaders(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	var onResponseHeadersCalled bool

	c := NewCollector()
	c.OnResponseHeaders(func(r *Response) {
		onResponseHeadersCalled = true
		if r.Headers.Get("Content-Type") == "application/octet-stream" {
			r.Request.Abort()
		}
	})
	c.OnResponse(func(r *Response) {
		t.Error("OnResponse was called")
	})
	c.Visit(ts.URL + "/large_binary")
	if !onResponseHeadersCalled {
		t.Error("OnResponseHeaders was not called")
	}
}

func TestCollectorURLRevisit(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	c := NewCollector()

	visitCount := 0

	c.OnRequest(func(r *Request) {
		visitCount++
	})

	c.Visit(ts.URL)
	c.Visit(ts.URL)

	if visitCount != 1 {
		t.Error("URL revisited")
	}

	c.AllowURLRevisit = true

	c.Visit(ts.URL)
	c.Visit(ts.URL)

	if visitCount != 3 {
		t.Error("URL not revisited")
	}
}

func TestCollectorPostRevisit(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	postValue := "hello"
	postData := map[string]string{
		"name": postValue,
	}
	visitCount := 0

	c := NewCollector()
	c.OnResponse(func(r *Response) {
		if postValue != string(r.Body) {
			t.Error("Failed to send data with POST")
		}
		visitCount++
	})

	c.Post(ts.URL+"/login", postData)
	c.Post(ts.URL+"/login", postData)
	c.Post(ts.URL+"/login", map[string]string{
		"name":     postValue,
		"lastname": "world",
	})

	if visitCount != 2 {
		t.Error("URL POST revisited")
	}

	c.AllowURLRevisit = true

	c.Post(ts.URL+"/login", postData)
	c.Post(ts.URL+"/login", postData)

	if visitCount != 4 {
		t.Error("URL POST not revisited")
	}
}

func TestCollectorURLRevisitCheck(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	c := NewCollector()

	visited, err := c.HasVisited(ts.URL)

	if err != nil {
		t.Error(err.Error())
	}

	if visited != false {
		t.Error("Expected URL to NOT have been visited")
	}

	c.Visit(ts.URL)

	visited, err = c.HasVisited(ts.URL)

	if err != nil {
		t.Error(err.Error())
	}

	if visited != true {
		t.Error("Expected URL to have been visited")
	}
}

func TestCollectorPostURLRevisitCheck(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	c := NewCollector()

	postValue := "hello"
	postData := map[string]string{
		"name": postValue,
	}

	posted, err := c.HasPosted(ts.URL+"/login", postData)

	if err != nil {
		t.Error(err.Error())
	}

	if posted != false {
		t.Error("Expected URL to NOT have been visited")
	}

	c.Post(ts.URL+"/login", postData)

	posted, err = c.HasPosted(ts.URL+"/login", postData)

	if err != nil {
		t.Error(err.Error())
	}

	if posted != true {
		t.Error("Expected URL to have been visited")
	}

	postData["lastname"] = "world"
	posted, err = c.HasPosted(ts.URL+"/login", postData)

	if err != nil {
		t.Error(err.Error())
	}

	if posted != false {
		t.Error("Expected URL to NOT have been visited")
	}

	c.Post(ts.URL+"/login", postData)

	posted, err = c.HasPosted(ts.URL+"/login", postData)

	if err != nil {
		t.Error(err.Error())
	}

	if posted != true {
		t.Error("Expected URL to have been visited")
	}
}

func TestCollectorURLRevisitDomainDisallowed(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	parsedURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	c := NewCollector(DisallowedDomains(parsedURL.Hostname()))
	err = c.Visit(ts.URL)
	if got, want := err, ErrForbiddenDomain; got != want {
		t.Fatalf("wrong error on first visit: got=%v want=%v", got, want)
	}
	err = c.Visit(ts.URL)
	if got, want := err, ErrForbiddenDomain; got != want {
		t.Fatalf("wrong error on second visit: got=%v want=%v", got, want)
	}

}

func TestCollectorPost(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	postValue := "hello"
	c := NewCollector()

	c.OnResponse(func(r *Response) {
		if postValue != string(r.Body) {
			t.Error("Failed to send data with POST")
		}
	})

	c.Post(ts.URL+"/login", map[string]string{
		"name": postValue,
	})
}

func TestCollectorPostRaw(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	postValue := "hello"
	c := NewCollector()

	c.OnResponse(func(r *Response) {
		if postValue != string(r.Body) {
			t.Error("Failed to send data with POST")
		}
	})

	c.PostRaw(ts.URL+"/login", []byte("name="+postValue))
}

func TestCollectorPostRawRevisit(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	postValue := "hello"
	postData := "name=" + postValue
	visitCount := 0

	c := NewCollector()
	c.OnResponse(func(r *Response) {
		if postValue != string(r.Body) {
			t.Error("Failed to send data with POST RAW")
		}
		visitCount++
	})

	c.PostRaw(ts.URL+"/login", []byte(postData))
	c.PostRaw(ts.URL+"/login", []byte(postData))
	c.PostRaw(ts.URL+"/login", []byte(postData+"&lastname=world"))

	if visitCount != 2 {
		t.Error("URL POST RAW revisited")
	}

	c.AllowURLRevisit = true

	c.PostRaw(ts.URL+"/login", []byte(postData))
	c.PostRaw(ts.URL+"/login", []byte(postData))

	if visitCount != 4 {
		t.Error("URL POST RAW not revisited")
	}
}

func TestCollectorCookies(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	c := NewCollector()

	if err := c.Visit(ts.URL + "/set_cookie"); err != nil {
		t.Fatal(err)
	}

	if err := c.Visit(ts.URL + "/check_cookie"); err != nil {
		t.Fatalf("Failed to use previously set cookies: %s", err)
	}
}

func TestConnectionErrorOnRobotsTxtResultsInError(t *testing.T) {
	ts := newTestServer()
	ts.Close()

	c := NewCollector()
	c.IgnoreRobotsTxt = false
	err := c.Visit(ts.URL)

	if err == nil {
		t.Fatal("Error expected")
	}
}

func TestUserAgent(t *testing.T) {
	const exampleUserAgent1 = "Example/1.0"
	const exampleUserAgent2 = "Example/2.0"
	const defaultUserAgent = "crawler - https://github.com/flywave/go-tileproxy/crawler"

	ts := newTestServer()
	defer ts.Close()

	var receivedUserAgent string

	func() {
		c := NewCollector()
		c.OnResponse(func(resp *Response) {
			receivedUserAgent = string(resp.Body)
		})
		c.Visit(ts.URL + "/user_agent")
		if got, want := receivedUserAgent, defaultUserAgent; got != want {
			t.Errorf("mismatched User-Agent: got=%q want=%q", got, want)
		}
	}()
	func() {
		c := NewCollector(UserAgent(exampleUserAgent1))
		c.OnResponse(func(resp *Response) {
			receivedUserAgent = string(resp.Body)
		})
		c.Visit(ts.URL + "/user_agent")
		if got, want := receivedUserAgent, exampleUserAgent1; got != want {
			t.Errorf("mismatched User-Agent: got=%q want=%q", got, want)
		}
	}()
	func() {
		c := NewCollector(UserAgent(exampleUserAgent1))
		c.OnResponse(func(resp *Response) {
			receivedUserAgent = string(resp.Body)
		})

		c.Request("GET", ts.URL+"/user_agent", nil, nil, nil)
		if got, want := receivedUserAgent, exampleUserAgent1; got != want {
			t.Errorf("mismatched User-Agent (nil hdr): got=%q want=%q", got, want)
		}
	}()
	func() {
		c := NewCollector(UserAgent(exampleUserAgent1))
		c.OnResponse(func(resp *Response) {
			receivedUserAgent = string(resp.Body)
		})

		c.Request("GET", ts.URL+"/user_agent", nil, nil, http.Header{})
		if got, want := receivedUserAgent, exampleUserAgent1; got != want {
			t.Errorf("mismatched User-Agent (non-nil hdr): got=%q want=%q", got, want)
		}
	}()
	func() {
		c := NewCollector(UserAgent(exampleUserAgent1))
		c.OnResponse(func(resp *Response) {
			receivedUserAgent = string(resp.Body)
		})
		hdr := http.Header{}
		hdr.Set("User-Agent", "")

		c.Request("GET", ts.URL+"/user_agent", nil, nil, hdr)
		if got, want := receivedUserAgent, ""; got != want {
			t.Errorf("mismatched User-Agent (hdr with empty UA): got=%q want=%q", got, want)
		}
	}()
	func() {
		c := NewCollector(UserAgent(exampleUserAgent1))
		c.OnResponse(func(resp *Response) {
			receivedUserAgent = string(resp.Body)
		})
		hdr := http.Header{}
		hdr.Set("User-Agent", exampleUserAgent2)

		c.Request("GET", ts.URL+"/user_agent", nil, nil, hdr)
		if got, want := receivedUserAgent, exampleUserAgent2; got != want {
			t.Errorf("mismatched User-Agent (hdr with UA): got=%q want=%q", got, want)
		}
	}()
}

func TestCollectorVisitWithTrace(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	c := NewCollector(AllowedDomains("localhost", "127.0.0.1", "::1"), TraceHTTP())
	c.OnResponse(func(resp *Response) {
		if resp.Trace == nil {
			t.Error("Failed to initialize trace")
		}
	})

	err := c.Visit(ts.URL)
	if err != nil {
		t.Errorf("Failed to visit url %s", ts.URL)
	}
}

func TestCollectorVisitWithCheckHead(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	c := NewCollector(CheckHead())
	var requestMethodChain []string
	c.OnResponse(func(resp *Response) {
		requestMethodChain = append(requestMethodChain, resp.Request.Method)
	})

	err := c.Visit(ts.URL)
	if err != nil {
		t.Errorf("Failed to visit url %s", ts.URL)
	}
	if requestMethodChain[0] != "HEAD" && requestMethodChain[1] != "GET" {
		t.Errorf("Failed to perform a HEAD request before GET")
	}
}

func TestCollectorContext(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	c := NewCollector(StdlibContext(ctx))

	onErrorCalled := false

	c.OnResponse(func(resp *Response) {
		t.Error("OnResponse was called, expected OnError")
	})

	c.OnError(func(resp *Response, err error) {
		onErrorCalled = true
		if err != context.DeadlineExceeded {
			t.Errorf("OnError got err=%#v, expected context.DeadlineExceeded", err)
		}
	})

	err := c.Visit(ts.URL + "/slow")
	if err != context.DeadlineExceeded {
		t.Errorf("Visit return err=%#v, expected context.DeadlineExceeded", err)
	}

	if !onErrorCalled {
		t.Error("OnError was not called")
	}

}

func BenchmarkOnResponse(b *testing.B) {
	ts := newTestServer()
	defer ts.Close()

	c := NewCollector()
	c.AllowURLRevisit = true
	c.OnResponse(func(_ *Response) {})

	for n := 0; n < b.N; n++ {
		c.Visit(ts.URL)
	}
}