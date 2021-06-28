package backend

type ProxyType int32

const (
	PROXY_DIRECT ProxyType = 0
	PROXY_HTTP1  ProxyType = 1
	PROXY_HTTP2  ProxyType = 2
	PROXY_HTTPS  ProxyType = 3
	PROXY_QUIC   ProxyType = 4
	PROXY_SOCKS4 ProxyType = 5
	PROXY_SOCKS5 ProxyType = 6
	PROXY_SSH2   ProxyType = 7
)

type Proxy interface {
	Type() ProxyType
	Address() string
	Auth() bool
	Username() string
	Password() string
	RequestFilters() []string
	RoundTripFilters() []string
	ResponseFilters() []string
}
