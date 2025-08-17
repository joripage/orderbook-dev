package riskrule

import "github.com/joripage/orderbook-dev/pkg/oms/model"

type RiskRule interface {
	Check(order *model.Order) error
}
