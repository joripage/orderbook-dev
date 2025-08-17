package fixmanager

import (
	"time"

	"github.com/quickfixgo/enum"
	"github.com/shopspring/decimal"
)

type NewOrderSingle struct {
	SenderCompID     string
	SenderSubID      string
	TargetCompID     string
	OnBehalfOfCompID string
	DeliverToCompID  string

	Account           string
	AccountType       enum.AccountType
	ClOrdID           string
	Symbol            string
	OrdType           enum.OrdType
	Price             decimal.Decimal
	TimeInForce       enum.TimeInForce
	Side              enum.Side
	TransactTime      time.Time
	OrderQty          decimal.Decimal
	MaturityMonthYear string
	SecurityType      enum.SecurityType
	SecurityID        string
	MaxFloor          decimal.Decimal
}

type OrderCancelRequest struct {
	SenderCompID     string
	SenderSubID      string
	TargetCompID     string
	OnBehalfOfCompID string
	DeliverToCompID  string

	OrigClOrderID     string
	ClOrderID         string
	Account           string
	Symbol            string
	Side              string
	TransactTime      time.Time
	MaturityMonthYear string
	SecurityType      string
	SecurityID        string
}
