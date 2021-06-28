package proxy

import (
	"net"
	"strings"
)

type PerHost struct {
	def, bypass Dialer

	bypassNetworks []*net.IPNet
	bypassIPs      []net.IP
	bypassZones    []string
	bypassHosts    []string
}

func NewPerHost(defaultDialer, bypass Dialer) *PerHost {
	return &PerHost{
		def:    defaultDialer,
		bypass: bypass,
	}
}

func (p *PerHost) Dial(network, addr string) (c net.Conn, err error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	return p.dialerForRequest(host).Dial(network, addr)
}

func (p *PerHost) dialerForRequest(host string) Dialer {
	if ip := net.ParseIP(host); ip != nil {
		for _, net := range p.bypassNetworks {
			if net.Contains(ip) {
				return p.bypass
			}
		}
		for _, bypassIP := range p.bypassIPs {
			if bypassIP.Equal(ip) {
				return p.bypass
			}
		}
		return p.def
	}
	for _, zone := range p.bypassZones {
		if strings.HasSuffix(host, zone) {
			return p.bypass
		}
		if host == zone[1:] {
			return p.bypass
		}
	}
	for _, bypassHost := range p.bypassHosts {
		if bypassHost == host {
			return p.bypass
		}
	}
	return p.def
}

func (p *PerHost) AddFromString(s string) {
	hosts := strings.Split(s, ",")
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		if len(host) == 0 {
			continue
		}
		if strings.Contains(host, "/") {
			if _, net, err := net.ParseCIDR(host); err == nil {
				p.AddNetwork(net)
			}
			continue
		}
		if ip := net.ParseIP(host); ip != nil {
			p.AddIP(ip)
			continue
		}
		if strings.HasPrefix(host, "*.") {
			p.AddZone(host[1:])
			continue
		}
		p.AddHost(host)
	}
}

func (p *PerHost) AddIP(ip net.IP) {
	p.bypassIPs = append(p.bypassIPs, ip)
}

func (p *PerHost) AddNetwork(net *net.IPNet) {
	p.bypassNetworks = append(p.bypassNetworks, net)
}

func (p *PerHost) AddZone(zone string) {
	if strings.HasSuffix(zone, ".") {
		zone = zone[:len(zone)-1]
	}
	if !strings.HasPrefix(zone, ".") {
		zone = "." + zone
	}
	p.bypassZones = append(p.bypassZones, zone)
}

func (p *PerHost) AddHost(host string) {
	if strings.HasSuffix(host, ".") {
		host = host[:len(host)-1]
	}
	p.bypassHosts = append(p.bypassHosts, host)
}
