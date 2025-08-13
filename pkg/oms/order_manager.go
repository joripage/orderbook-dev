package oms

import "oms-fix/pkg/oms/model"

type OrderManager interface {
	// client to oms
	AddOrder(addOrder *model.AddOrder)
	ModifyOrder()
	CancelOrder()

	// oms to client
	OnOrderReport(args ...interface{})
}
