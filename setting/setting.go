package setting

type ProxyDataset struct {
	Identifier string
	Service    interface{}
	Coverages  map[string]Coverage
	Grids      map[string]GridOpts
	Sources    map[string]interface{}
	Caches     map[string]interface{}
	Seeds      map[string]Seed
	Cleanups   map[string]Cleanup
}

func NewProxyDataset(name string) *ProxyDataset {
	return &ProxyDataset{
		Identifier: name,
		Coverages:  make(map[string]Coverage),
		Sources:    make(map[string]interface{}),
		Caches:     make(map[string]interface{}),
		Seeds:      make(map[string]Seed),
		Cleanups:   make(map[string]Cleanup),
	}
}
