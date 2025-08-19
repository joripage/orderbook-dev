package oms

import "errors"

var (
	errDuplicateOrder     = errors.New("dupplicate order")
	errOrderIDNotFound    = errors.New("orderID not found")
	errGatewayIDNotFound  = errors.New("gatewayID not found")
	errInvalidOrderStatus = errors.New("invalid order status")
)
