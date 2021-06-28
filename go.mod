module github.com/flywave/go-tileproxy

go 1.16

require (
	boringssl.googlesource.com/boringssl v0.0.0-20210514210023-ddecaabdc8c9
	github.com/bitly/go-simplejson v0.5.0
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/cheekybits/is v0.0.0-20150225183255-68e9c0620927 // indirect
	github.com/cloudflare/golibs v0.0.0-20201113145655-eb7a42c5e0be
	github.com/flywave/go-geom v0.0.0-20210628091419-7d55414ca1d3
	github.com/flywave/go-geos v0.0.0-20210608062102-81fc76bf44c4
	github.com/flywave/go-lerc v0.0.0-20191107120907-ad19c08efa45
	github.com/flywave/go-mapbox v0.0.0-20210624012915-79a7911c3c00
	github.com/flywave/go-proj v0.0.0-20210605132706-d05d11f58021
	github.com/flywave/go-raster v0.0.0-20210526065301-f50e348f662e // indirect
	github.com/flywave/go3d v0.0.0-20210521003526-9185b600148d
	github.com/flywave/imaging v1.6.5
	github.com/fogleman/gg v1.3.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/jonas-p/go-shp v0.1.1
	github.com/juju/ratelimit v1.0.1
	github.com/kr/pretty v0.2.0 // indirect
	github.com/lucas-clemente/quic-go v0.21.0-rc.2
	github.com/miekg/dns v1.1.42
	github.com/paulmach/go.geo v0.0.0-20180829195134-22b514266d33 // indirect
	github.com/rubenv/topojson v0.0.0-20180822134236-13be738db397
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/image v0.0.0-20210607152325-775e3b0c77b9
	golang.org/x/net v0.0.0-20210510120150-4163338589ed
	golang.org/x/sys v0.0.0-20210423082822-04245dca01da
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace boringssl.googlesource.com/boringssl => github.com/google/boringssl v0.0.0-20210514210023-ddecaabdc8c9
