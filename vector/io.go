package vector

import "io"

type VectorIO interface {
	Decode(r io.Reader) (Vector, error)
	Encode(data Vector) ([]byte, error)
}
