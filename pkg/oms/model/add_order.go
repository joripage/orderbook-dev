package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type AddOrder struct {
	ID           string
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
