package fixgateway

import (
	"time"

	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/quickfix"
	"github.com/shopspring/decimal"
)

type NewOrderSingle struct {
	SessionID *quickfix.SessionID
	// SenderCompID     string
	// SenderSubID      string
	// TargetCompID     string
	// OnBehalfOfCompID string
	// DeliverToCompID  string

	Account           string
	AccountType       enum.AccountType
	ClOrdID           string
	Symbol            string
	SecurityID        string
	SecurityType      enum.SecurityType
	OrdType           enum.OrdType
	Price             decimal.Decimal
	TimeInForce       enum.TimeInForce
	Side              enum.Side
	TransactTime      time.Time
	OrderQty          decimal.Decimal
	MaturityMonthYear string

	MaxFloor decimal.Decimal
}

type OrderCancelRequest struct {
	SessionID *quickfix.SessionID
	// SenderCompID     string
	// SenderSubID      string
	// TargetCompID     string
	// OnBehalfOfCompID string
	// DeliverToCompID  string

	OrigClOrderID     string
	ClOrderID         string
	Account           string
	Symbol            string
	SecurityID        string
	SecurityType      enum.SecurityType
	Side              enum.Side
	TransactTime      time.Time
	MaturityMonthYear string
}

type OrderCancelReplaceRequest struct {
	SessionID *quickfix.SessionID
	// SenderCompID     string
	// SenderSubID      string
	// TargetCompID     string
	// OnBehalfOfCompID string
	// DeliverToCompID  string

	OrigClOrderID     string
	ClOrderID         string
	Account           string
	Symbol            string
	SecurityID        string
	SecurityType      enum.SecurityType
	Side              enum.Side
	TransactTime      time.Time
	OrderQty          decimal.Decimal
	OrdType           enum.OrdType
	Price             decimal.Decimal
	TimeInForce       enum.TimeInForce
	MaturityMonthYear string
}
