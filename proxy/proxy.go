package proxy

import (
	"errors"
	"net"
	"net/url"
	"os"
)

type Dialer interface {
	Dial(network, addr string) (c net.Conn, err error)
}

type Resolver interface {
	LookupHost(host string) (addrs []string, err error)
}

type Auth struct {
	User, Password string
}

func FromEnvironment() Dialer {
	allProxy := os.Getenv("all_proxy")
	if len(allProxy) == 0 {
		return Direct
	}

	proxyURL, err := url.Parse(allProxy)
	if err != nil {
		return Direct
	}
	proxy, err := FromURL(proxyURL, Direct, DummyResolver)
	if err != nil {
		return Direct
	}

	noProxy := os.Getenv("no_proxy")
	if len(noProxy) == 0 {
		return proxy
	}

	perHost := NewPerHost(proxy, Direct)
	perHost.AddFromString(noProxy)
	return perHost
}

var proxySchemes map[string]func(*url.URL, Dialer) (Dialer, error)

func RegisterDialerType(scheme string, f func(*url.URL, Dialer) (Dialer, error)) {
	if proxySchemes == nil {
		proxySchemes = make(map[string]func(*url.URL, Dialer) (Dialer, error))
	}
	proxySchemes[scheme] = f
}

func FromURL(u *url.URL, forward Dialer, resolver Resolver) (Dialer, error) {
	var auth *Auth
	if u.User != nil {
		auth = new(Auth)
		auth.User = u.User.Username()
		if p, ok := u.User.Password(); ok {
			auth.Password = p
		}
	}

	switch u.Scheme {
	case "socks5", "socks":
		return SOCKS5("tcp", u.Host, auth, forward, resolver)
	case "socks4":
		return SOCKS4("tcp", u.Host, false, forward, resolver)
	case "socks4a":
		return SOCKS4("tcp", u.Host, true, forward, resolver)
	case "http":
		return HTTP1("tcp", u.Host, auth, forward, resolver)
	case "https":
		return HTTPS("tcp", u.Host, auth, forward, resolver)
	case "https+h2":
		return HTTP2("tcp", u.Host, auth, forward, resolver)
	case "ssh", "ssh2":
		return SSH2("tcp", u.Host, auth, forward, resolver)
	case "quic":
		return QUIC("udp", u.Host, auth, forward, resolver)
	}

	if proxySchemes != nil {
		if f, ok := proxySchemes[u.Scheme]; ok {
			return f(u, forward)
		}
	}

	return nil, errors.New("proxy: unknown scheme: " + u.Scheme)
}
