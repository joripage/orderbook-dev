package orderbook

// PriceHeap implements heap.Interface
type PriceHeap struct {
	prices []float64
	less   func(i, j float64) bool
	index  map[float64]bool
}

func NewPriceHeap(less func(i, j float64) bool) *PriceHeap {
	return &PriceHeap{
		prices: []float64{},
		less:   less,
		index:  make(map[float64]bool),
	}
}

func (h PriceHeap) Len() int {
	return len(h.prices)
}

func (h PriceHeap) Less(i, j int) bool {
	return h.less(h.prices[i], h.prices[j])
}

func (h PriceHeap) Swap(i, j int) {
	h.prices[i], h.prices[j] = h.prices[j], h.prices[i]
}

func (h *PriceHeap) Push(x any) {
	price := x.(float64)
	if !h.index[price] {
		h.index[price] = true
		h.prices = append(h.prices, price)
	}
}

func (h *PriceHeap) Pop() any {
	n := len(h.prices)
	price := h.prices[n-1]
	h.prices = h.prices[:n-1]
	delete(h.index, price)
	return price
}

func (h *PriceHeap) Peek() (float64, bool) {
	if len(h.prices) == 0 {
		return 0, false
	}
	return h.prices[0], true
}
