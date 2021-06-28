package proxy

type dummyResolver struct{}

var DummyResolver = dummyResolver{}

func (dummyResolver) LookupHost(host string) (addrs []string, err error) {
	return []string{host}, nil
}
