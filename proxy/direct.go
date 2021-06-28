package proxy

import (
	"net"
)

type direct struct{}

var Direct = direct{}

func (direct) Dial(network, addr string) (net.Conn, error) {
	return net.Dial(network, addr)
}
