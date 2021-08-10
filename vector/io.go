package vector

import "io"

type VectorIO interface {
	Decode(r io.Reader) (interface{}, error)
	Encode(data interface{}) ([]byte, error)
}
