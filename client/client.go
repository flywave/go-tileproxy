package client

import "github.com/flywave/go-tileproxy/crawler"

type Client interface {
}

type BaseClient struct {
	Client
	Collector *crawler.Collector
}
