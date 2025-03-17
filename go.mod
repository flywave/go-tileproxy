module github.com/flywave/go-tileproxy

go 1.16

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/beevik/etree v1.1.0
	github.com/deckarep/golang-set v1.8.0
	github.com/flywave/gg v1.3.1-0.20210910115449-fe7fd154baa2
	github.com/flywave/go-cog v0.0.0-20250314092301-4673589220b8
	github.com/flywave/go-geo v0.0.0-20250314091853-e818cb9de299
	github.com/flywave/go-geoid v0.0.0-20220306024153-21126c4758a2
	github.com/flywave/go-geom v0.0.0-20220210023939-86f84322e71f
	github.com/flywave/go-geos v0.0.0-20220312005430-b3e54ee96ed7
	github.com/flywave/go-gpkg v0.0.0-20211215020551-c4e5f7d53419
	github.com/flywave/go-gpx v1.2.2-0.20211027141055-7fa376dde073
	github.com/flywave/go-lerc v0.0.0-20210724083528-61c70a1b0bc9
	github.com/flywave/go-mapbox v0.0.0-20250314092441-27874854ad1b
	github.com/flywave/go-mbgeom v0.0.0-20220407004130-9a9ce7702726
	github.com/flywave/go-obj v0.0.0-20210526030750-7674effc90f7
	github.com/flywave/go-proj v0.0.0-20211220121303-46dc797a5cd0
	github.com/flywave/go-quantized-mesh v0.0.0-20220602084751-eb72cc1f9c21
	github.com/flywave/go-tin v0.0.0-20220223031304-eac1b215d1cb
	github.com/flywave/go-xslt v0.0.0-20210730032627-a21173f9ee67
	github.com/flywave/go3d v0.0.0-20231213061711-48d3c5834480
	github.com/flywave/imaging v1.6.5
	github.com/flywave/ogc-osgeo v0.0.0-20220121133505-c3a428aee8fc
	github.com/flywave/webp v1.1.1
	github.com/gobwas/glob v0.2.3
	github.com/gofrs/uuid v4.2.0+incompatible
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/tiff v0.0.0-20161109161721-4b31f3041d9a
	github.com/hpinc/go3mf v0.24.0
	github.com/hschendel/stl v1.0.4
	github.com/kennygrant/sanitize v1.2.4
	github.com/klauspost/compress v1.14.2 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0
	github.com/mattn/go-sqlite3 v1.14.10 // indirect
	github.com/mholt/archiver/v3 v3.5.1
	github.com/nlnwa/whatwg-url v0.1.0
	github.com/nwaples/rardecode v1.1.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.12 // indirect
	github.com/qmuntal/gltf v0.20.3
	github.com/qmuntal/opc v0.7.11 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/ulikunitz/xz v0.5.10 // indirect
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce // indirect
	golang.org/x/image v0.14.0
	golang.org/x/net v0.6.0
	golang.org/x/sys v0.5.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
)

replace github.com/flywave/go-mbgeom => ../go-mbgeom

replace github.com/flywave/go-geom => ../go-geom

replace github.com/flywave/go-geo => ../go-geo
