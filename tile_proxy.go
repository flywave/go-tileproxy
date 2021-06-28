package tileproxy

import (
	"net/http"
	"time"

	"github.com/flywave/go-tileproxy/backend"
	"github.com/flywave/go-tileproxy/filters"
	"github.com/flywave/go-tileproxy/utils"
)

type ProxyConfig struct {
	Enabled          bool
	Address          string
	KeepAlivePeriod  int
	ReadTimeout      int
	WriteTimeout     int
	RequestFilters   []string
	RoundTripFilters []string
	ResponseFilters  []string
}

func (c *ProxyConfig) fromKVS(conf map[string]backend.KVS) error {
	return nil
}

func ServeProfile(conf backend.Config, branding string) error {
	pconfig := new(ProxyConfig)
	pconfig.fromKVS(conf[backend.EnvProxy])
	listenOpts := &utils.ListenOptions{TLSConfig: nil}

	ln, err := utils.ListenTCP("tcp", pconfig.Address, listenOpts)
	if err != nil {
		//glog.Fatalf("ListenTCP(%s, %#v) error: %s", config.Address, listenOpts, err)
	}

	h := Handler{
		Listener:         ln,
		RequestFilters:   []filters.RequestFilter{},
		RoundTripFilters: []filters.RoundTripFilter{},
		ResponseFilters:  []filters.ResponseFilter{},
		Branding:         branding,
	}

	for _, name := range pconfig.RequestFilters {
		f, _ := filters.GetFilter(name, conf)
		f1, ok := f.(filters.RequestFilter)
		if !ok {
			//glog.Fatalf("%#v is not a RequestFilter, err=%+v", f, err)
		}
		h.RequestFilters = append(h.RequestFilters, f1)
	}

	for _, name := range pconfig.RoundTripFilters {
		f, _ := filters.GetFilter(name, conf)
		f1, ok := f.(filters.RoundTripFilter)
		if !ok {
			//glog.Fatalf("%#v is not a RoundTripFilter, err=%+v", f, err)
		}
		h.RoundTripFilters = append(h.RoundTripFilters, f1)
	}

	for _, name := range pconfig.ResponseFilters {
		f, _ := filters.GetFilter(name, conf)
		f1, ok := f.(filters.ResponseFilter)
		if !ok {
			//glog.Fatalf("%#v is not a ResponseFilter, err=%+v", f, err)
		}
		h.ResponseFilters = append(h.ResponseFilters, f1)
	}

	s := &http.Server{
		Handler:        h,
		ReadTimeout:    time.Duration(pconfig.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(pconfig.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return s.Serve(h.Listener)
}
