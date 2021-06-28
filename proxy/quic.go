package proxy

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudflare/golibs/lrucache"

	quicconn "github.com/lucas-clemente/quic-go"
)

func QUIC(network, addr string, auth *Auth, forward Dialer, resolver Resolver) (Dialer, error) {
	var hostname string

	if host, _, err := net.SplitHostPort(addr); err == nil {
		hostname = host
	} else {
		hostname = addr
		addr = net.JoinHostPort(addr, "443")
	}

	s := &quic{
		network:  network,
		addr:     addr,
		hostname: hostname,
		forward:  forward,
		resolver: resolver,
		cache:    lrucache.NewLRUCache(128),
	}
	if auth != nil {
		s.user = auth.User
		s.password = auth.Password
	}

	return s, nil
}

type QuicConn struct {
	net.Conn
	session quicconn.Session

	receiveStream quicconn.Stream
	sendStream    quicconn.Stream
}

func newConn(sess quicconn.Session) (*QuicConn, error) {
	stream, err := sess.OpenStream()
	if err != nil {
		return nil, err
	}
	return &QuicConn{
		session:    sess,
		sendStream: stream,
	}, nil
}

func (c *QuicConn) Read(b []byte) (n int, err error) {
	if c.receiveStream == nil {
		var err error
		c.receiveStream, err = c.session.AcceptStream(context.Background())

		if err != nil {
			return 0, err
		}
		err = c.receiveStream.Close()
		if err != nil {
			return 0, err
		}
	}

	return c.receiveStream.Read(b)
}

func (c *QuicConn) Write(b []byte) (n int, err error) {
	return c.sendStream.Write(b)
}

func (c *QuicConn) Close() error {
	return c.session.CloseWithError(0, "")
}

func (c *QuicConn) LocalAddr() net.Addr {
	return c.session.LocalAddr()
}

func (c *QuicConn) RemoteAddr() net.Addr {
	return c.session.RemoteAddr()
}

func (c *QuicConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *QuicConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *QuicConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type quic struct {
	user, password string
	network, addr  string
	hostname       string
	forward        Dialer
	resolver       Resolver
	cache          lrucache.Cache
}

func (h *quic) Dial(network, addr string) (net.Conn, error) {
	switch network {
	case "tcp", "tcp6", "tcp4":
	default:
		return nil, errors.New("proxy: no support for QUIC proxy connections of type " + network)
	}

	var config *tls.Config
	if v, ok := h.cache.GetNotStale(h.addr); ok {
		config = v.(*tls.Config)
	} else {
		config = &tls.Config{
			MinVersion:         tls.VersionTLS10,
			MaxVersion:         tls.VersionTLS13,
			InsecureSkipVerify: true,
			ServerName:         h.hostname,
			ClientSessionCache: tls.NewLRUClientSessionCache(1024),
		}
		h.cache.Set(h.addr, config, time.Now().Add(2*time.Hour))
	}

	conn, err := quicconn.DialAddr(h.addr, config, nil)
	if err != nil {
		return nil, err
	}
	closeConn := &conn
	defer func() {
		if closeConn != nil {
			(*closeConn).CloseWithError(0, err.Error())
		}
	}()

	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, errors.New("proxy: failed to parse port number: " + portStr)
	}
	if port < 1 || port > 0xffff {
		return nil, errors.New("proxy: port number out of range: " + portStr)
	}

	if h.resolver != nil {
		hosts, err := h.resolver.LookupHost(host)
		if err == nil && len(hosts) > 0 {
			host = hosts[0]
		}
	}

	b := bufPool.Get().(*bytes.Buffer)
	b.Reset()

	fmt.Fprintf(b, "CONNECT %s:%s HTTP/1.1\r\nHost: %s:%s\r\n", host, portStr, host, portStr)
	if h.user != "" {
		fmt.Fprintf(b, "Proxy-Authorization: Basic %s\r\n", base64.StdEncoding.EncodeToString([]byte(h.user+":"+h.password)))
	}
	io.WriteString(b, "\r\n")

	bb := b.Bytes()
	bufPool.Put(b)

	if err := conn.SendMessage(bb); err != nil {
		return nil, errors.New("proxy: failed to write greeting to HTTP proxy at " + h.addr + ": " + err.Error())
	}

	var ncon net.Conn

	buf := make([]byte, 2048)
	b0 := buf
	total := 0

	for {
		buf, err := conn.ReceiveMessage()
		if err != nil {
			return nil, err
		}
		n := len(buf)
		total += n
		buf = buf[n:]

		if i := bytes.Index(b0, CRLFCRLF); i > 0 {
			ncon, err = newConn(conn)
			if err == nil {
				ncon = &preReaderConn{ncon, b0[i+4 : total]}
				b0 = b0[:i+4]
			}
			break
		}
	}

	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(b0)), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("proxy: failed to read greeting from HTTP proxy at " + h.addr + ": " + resp.Status)
	}

	closeConn = nil
	return ncon, nil
}
