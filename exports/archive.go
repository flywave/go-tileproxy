package exports

import _ "github.com/mholt/archiver/v3"

const (
	TARGZ = ".tar.gz"
	ZIP   = ".zip"
)

type ArchiveOptions struct {
	DirectoryLayout  string
	Ext              string
	CompressionLevel int
}
