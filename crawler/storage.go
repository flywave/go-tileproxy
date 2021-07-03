package crawler

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
)

type Storage interface {
	Init() error
	Visited(requestID uint64) error
	IsVisited(requestID uint64) (bool, error)
	Cookies(u *url.URL) string
	SetCookies(u *url.URL, cookies string)
}

type InMemoryStorage struct {
	visitedURLs map[uint64]bool
	lock        *sync.RWMutex
	jar         *cookiejar.Jar
}

func (s *InMemoryStorage) Init() error {
	if s.visitedURLs == nil {
		s.visitedURLs = make(map[uint64]bool)
	}
	if s.lock == nil {
		s.lock = &sync.RWMutex{}
	}
	if s.jar == nil {
		var err error
		s.jar, err = cookiejar.New(nil)
		return err
	}
	return nil
}

func (s *InMemoryStorage) Visited(requestID uint64) error {
	s.lock.Lock()
	s.visitedURLs[requestID] = true
	s.lock.Unlock()
	return nil
}

func (s *InMemoryStorage) IsVisited(requestID uint64) (bool, error) {
	s.lock.RLock()
	visited := s.visitedURLs[requestID]
	s.lock.RUnlock()
	return visited, nil
}

func (s *InMemoryStorage) Cookies(u *url.URL) string {
	return StringifyCookies(s.jar.Cookies(u))
}

func (s *InMemoryStorage) SetCookies(u *url.URL, cookies string) {
	s.jar.SetCookies(u, UnstringifyCookies(cookies))
}

func (s *InMemoryStorage) Close() error {
	return nil
}

func StringifyCookies(cookies []*http.Cookie) string {
	cs := make([]string, len(cookies))
	for i, c := range cookies {
		cs[i] = c.String()
	}
	return strings.Join(cs, "\n")
}

func UnstringifyCookies(s string) []*http.Cookie {
	h := http.Header{}
	for _, c := range strings.Split(s, "\n") {
		h.Add("Set-Cookie", c)
	}
	r := http.Response{Header: h}
	return r.Cookies()
}

func ContainsCookie(cookies []*http.Cookie, name string) bool {
	for _, c := range cookies {
		if c.Name == name {
			return true
		}
	}
	return false
}
