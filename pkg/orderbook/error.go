package orderbook

import "errors"

var (
	errOrderNotFound     = errors.New("order not found")
	errInvalidOrderPrice = errors.New("invalid order price")
)
