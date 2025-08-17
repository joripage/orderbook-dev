package fixmanager

import "errors"

func (s *FixManager) GetNewOrderSingleByOrderID(orderID string) (*NewOrderSingle, error) {
	var newOrderSingle any
	var ok bool
	if newOrderSingle, ok = s.orderMapping.Load(orderID); !ok {
		return nil, errors.New("orderID not found")
	}

	return newOrderSingle.(*NewOrderSingle), nil
}

// func (s *FixManager) SetOrderInfoByOrderID(orderID string, orderInfo *OrderInfo) error {
// 	s.orderInfoMapping.Store(orderID, orderInfo)
// 	return nil
// }

func (s *FixManager) AddOrderToMap(order *NewOrderSingle) {
	s.orderMapping.Store(order.ClOrdID, order)
}
