package orderbook

type MatchResult struct {
	// BuyOrderID  string
	// SellOrderID string
	OrderID        string
	CounterOrderID string
	Price          float64
	Qty            int64
	Side           Side
}
