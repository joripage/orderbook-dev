package eventstore

import "github.com/joripage/orderbook-dev/pkg/oms/model"

type EventStore interface {
	AddEvent(ev *model.OrderEvent)
	TrackClOrdChain(orderID, clOrdID, origClOrdID string)
	GetLatestGatewayID(orderID string) string
	GetOrigGatewayID(clOrdID string) string
	GetOrderID(clOrdID string) string
	ReconstructChain(clOrdID string) []string
	DeleteChainByOrderID(orderID string)
}
