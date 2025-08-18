package oms

import "errors"

var (
	errOrderIDNotFound   = errors.New("orderID not found")
	errGatewayIDNotFound = errors.New("gatewayID not found")
)
