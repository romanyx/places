package broker

import "errors"

var (
	// ErrBadRequest returns when aviasales responds
	// with bad request status.
	ErrBadRequest = errors.New("bad request")
)
