package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type AddOrder struct {
	GatewayID    string
	Account      string
	Symbol       string
	SecurityID   string
	Exchange     string
	Type         OrderType
	Price        decimal.Decimal
	TimeInForce  OrderTimeInForce
	Side         OrderSide
	TransactTime time.Time
	Quantity     decimal.Decimal
}

type CancelOrder struct {
	GatewayID     string
	OrigGatewayID string
}

type ModifyOrder struct {
	NewPrice      decimal.Decimal
	NewQuantity   decimal.Decimal
	GatewayID     string
	OrigGatewayID string
}
