package fixgateway

import "github.com/quickfixgo/quickfix"

// func (s *FixGateway) AddNewOrderSingleToMap(order *NewOrderSingle) {
// 	s.newOrderSingleMapping.Store(order.ClOrdID, order)
// }

// func (s *FixGateway) GetNewOrderSingleByClOrdID(clOrdID string) (*NewOrderSingle, error) {
// 	var newOrderSingle any
// 	var ok bool
// 	if newOrderSingle, ok = s.newOrderSingleMapping.Load(clOrdID); !ok {
// 		return nil, errClOrdIDNotFound
// 	}

// 	return newOrderSingle.(*NewOrderSingle), nil
// }

// func (s *FixGateway) AddOrderCancelRequestToMap(cancelRequest *OrderCancelRequest) {
// 	s.orderCancelRequestMapping.Store(cancelRequest.ClOrderID, cancelRequest)
// }

// func (s *FixGateway) GetOrderCancelRequestByClOrdID(clOrdID string) (*OrderCancelRequest, error) {
// 	var cancelRequest any
// 	var ok bool
// 	if cancelRequest, ok = s.newOrderSingleMapping.Load(clOrdID); !ok {
// 		return nil, errClOrdIDNotFound
// 	}

// 	return cancelRequest.(*OrderCancelRequest), nil
// }

func (s *FixGateway) AddRequestToMap(id string, request interface{}) {
	s.requestMapping.Store(id, request)
}

func (s *FixGateway) GetRequestByClOrdID(clOrdID string) (*quickfix.SessionID, error) {
	var request any
	var ok bool
	if request, ok = s.requestMapping.Load(clOrdID); !ok {
		return nil, errClOrdIDNotFound
	}

	return request.(*quickfix.SessionID), nil
}
