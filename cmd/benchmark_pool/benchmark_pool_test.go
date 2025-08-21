package benchmarkpool

import (
	"sync"
	"testing"
	"time"

	"github.com/joripage/orderbook-dev/pkg/oms/model"
)

var orderPool = sync.Pool{
	New: func() interface{} {
		return &model.Order{}
	},
}

func BenchmarkNewOrder(b *testing.B) {
	arr := make([]*model.Order, 0, b.N)
	for i := 0; i < b.N; i++ {
		o := &model.Order{
			ID:            "ID",
			GatewayID:     "GatewayID",
			OrigGatewayID: "OrigGatewayID",

			// init info
			Symbol:       "Symbol",
			SecurityID:   "SecurityID",
			Exchange:     "Exchange",
			Side:         "Side",
			Type:         "Type",
			TimeInForce:  "TimeInForce",
			Price:        1000,
			Quantity:     100,
			Account:      "Account",
			TransactTime: time.Now(),
		}
		arr = append(arr, o)
		_ = o
	}
}

func BenchmarkPoolOrder(b *testing.B) {
	arr := make([]*model.Order, 0, b.N)
	for i := 0; i < b.N; i++ {
		s := orderPool.Get().(*model.Order)
		s.ID = "ID"
		s.GatewayID = "GatewayID"
		s.OrigGatewayID = "OrigGatewayID"
		s.Symbol = "Symbol"
		s.SecurityID = "SecurityID"
		s.Exchange = "Exchange"
		s.Side = "Side"
		s.Type = "Type"
		s.TimeInForce = "TimeInForce"
		s.Price = 1000
		s.Quantity = 100
		s.Account = "Account"
		s.TransactTime = time.Now()

		arr = append(arr, s)

		// reset trước khi trả về pool
		s.ID = ""
		s.GatewayID = ""
		s.OrigGatewayID = ""
		s.Symbol = ""
		s.SecurityID = ""
		s.Exchange = ""
		s.Side = ""
		s.Type = ""
		s.TimeInForce = ""
		s.Price = 0
		s.Quantity = 0
		s.Account = ""
		s.TransactTime = time.Now()
		orderPool.Put(s)
	}
}

var bufPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 64*1024) // 64KB buffer
		return &b
	},
}

func BenchmarkNewBuffer(b *testing.B) {
	buffers := make([][]byte, 0, b.N)
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 64*1024)
		buffers = append(buffers, buf)
		if len(buffers) > 1000 {
			// giữ lại nhiều buffer để ép GC
			buffers = buffers[:0]
		}
	}
}

func BenchmarkPoolBuffer(b *testing.B) {
	buffers := make([]*[]byte, 0, b.N)
	for i := 0; i < b.N; i++ {
		buf := bufPool.Get().(*[]byte)
		buffers = append(buffers, buf)
		if len(buffers) > 1000 {
			// reset về pool
			for _, bb := range buffers {
				bufPool.Put(bb)
			}
			buffers = buffers[:0]
		}
	}
}
