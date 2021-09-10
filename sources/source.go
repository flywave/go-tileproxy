package sources

import (
	"errors"
)

var ErrInvalidBBOX = errors.New("invalid bbox")

var ErrInvalidSourceQuery = errors.New("invalid source query")

type Source interface {
}
