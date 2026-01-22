module github.com/flywave/go-tileproxy

go 1.24

toolchain go1.24.4

require (
	github.com/beevik/etree v1.1.0
	github.com/deckarep/golang-set v1.8.0
	github.com/flywave/gg v1.3.1-0.20210910115449-fe7fd154baa2
	github.com/flywave/go-cog v0.0.0-20250314092301-4673589220b8
	github.com/flywave/go-geo v0.0.0-20250314091853-e818cb9de299
	github.com/flywave/go-geoid v0.0.0-20220306024153-21126c4758a2
	github.com/flywave/go-geom v0.0.0-20250607125323-f685bf20f12c
	github.com/flywave/go-geos v0.0.0-20220312005430-b3e54ee96ed7
	github.com/flywave/go-gpkg v0.0.0-20220505053116-3682bbf4ea48
	github.com/flywave/go-lerc v0.0.0-20210724083528-61c70a1b0bc9
	github.com/flywave/go-mapbox v0.0.0-20250314092441-27874854ad1b
	github.com/flywave/go-mbgeom v0.0.0-20220407004130-9a9ce7702726
	github.com/flywave/go-proj v0.0.0-20211220121303-46dc797a5cd0
	github.com/flywave/go-quantized-mesh v0.0.0-20220602084751-eb72cc1f9c21
	github.com/flywave/go-tin v0.0.0-20220223031304-eac1b215d1cb
	github.com/flywave/go-xslt v0.0.0-20210730032627-a21173f9ee67
	github.com/flywave/go3d v0.0.0-20250816053852-aed5d825659f
	github.com/flywave/imaging v1.6.5
	github.com/flywave/ogc-osgeo v0.0.0-20220121133505-c3a428aee8fc
	github.com/flywave/webp v1.1.2
	github.com/gobwas/glob v0.2.3
	github.com/gofrs/uuid v4.2.0+incompatible
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/google/tiff v0.0.0-20161109161721-4b31f3041d9a
	github.com/kennygrant/sanitize v1.2.4
	github.com/lucasb-eyer/go-colorful v1.2.0
	github.com/mholt/archiver/v3 v3.5.1
	github.com/spf13/cobra v1.10.2
	github.com/stretchr/testify v1.10.0
	golang.org/x/image v0.28.0
	golang.org/x/net v0.40.0
	golang.org/x/sys v0.33.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/devork/geom v0.0.5 // indirect
	github.com/dsnet/compress v0.0.2-0.20210315054119-f66993602bf5 // indirect
	github.com/flywave/go-pbf v0.0.0-20210701015929-a3bdb1f6728e // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/hhrutter/lzw v0.0.0-20190829144645-6f07a24e8650 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/gorm v1.9.16 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/klauspost/compress v1.14.2 // indirect
	github.com/klauspost/pgzip v1.2.5 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.10 // indirect
	github.com/nwaples/rardecode v1.1.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.12 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/flywave/go-mbgeom => ../go-mbgeom

replace github.com/flywave/go-geom => ../go-geom

replace github.com/flywave/go-geo => ../go-geo

replace github.com/flywave/go-geoid => ../go-geoid

replace github.com/flywave/go-proj => ../go-proj

replace github.com/flywave/go-xslt => ../go-xslt

replace github.com/flywave/go-geos => ../go-geos

replace github.com/flywave/go-tin => ../go-tin
