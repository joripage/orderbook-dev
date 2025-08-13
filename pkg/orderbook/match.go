package orderbook

type MatchResult struct {
	BuyOrderID  string
	SellOrderID string
	Price       float64
	Qty         int
}
