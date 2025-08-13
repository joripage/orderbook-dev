package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type AddOrder struct {
	Account      string
	Symbol       string
	Type         OrderType
	Price        decimal.Decimal
	TimeInForce  OrderTimeInForce
	Side         OrderSide
	TransactTime time.Time
	Quantity     decimal.Decimal
}
