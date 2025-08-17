package orderbook

type Side string

const (
	BUY  Side = "BUY"
	SELL Side = "SELL"
)

type OrderType string

const (
	LIMIT   OrderType = "LIMIT"
	MARKET  OrderType = "MARKET"
	ICEBERG OrderType = "ICEBERG"
)

type TimeInForce string

const (
	DAY TimeInForce = "DAY"
	IOC TimeInForce = "IOC"
	FOK TimeInForce = "FOK"
	GTC TimeInForce = "GTC"
)

type Order struct {
	ID          string
	Symbol      string
	Side        Side
	Price       float64
	Qty         int64
	Type        OrderType
	TimeInForce TimeInForce // IOC, FOK, GTC, etc.
	VisibleQty  int64       // for Iceberg: public visible quantity
	hiddenQty   int64       // for Iceberg: internal qty
}
