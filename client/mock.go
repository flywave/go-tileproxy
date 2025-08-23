package client

import "net/http"

type mockClient struct {
	HttpClient
	data []byte
	url  string
	body []byte
	code int
}

func (c *mockClient) Open(url string, data []byte, hdr http.Header) (statusCode int, body []byte) {
	c.data = data
	c.url = url
	return c.code, c.body
}

type mockContext struct {
	Context
	c *mockClient
}

func (c *mockContext) Client() HttpClient {
	return c.c
}

func (c *mockContext) Sync() {
}
