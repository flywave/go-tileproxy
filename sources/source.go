package sources

import (
	"errors"
)

var ErrInvalidBBOX = errors.New("Invalid BBOX")

var ErrInvalidSourceQuery = errors.New("Invalid source query")

type Source interface {
}
