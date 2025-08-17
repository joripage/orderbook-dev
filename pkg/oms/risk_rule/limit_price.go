package riskrule

import (
	"fmt"

	"github.com/joripage/orderbook-dev/pkg/oms/model"
)

type limitPrice struct {
	ceil  float64
	floor float64
}

type LimitPriceRule struct {
	prices map[string]*limitPrice
}

func (r *LimitPriceRule) Check(order *model.Order) error {
	if order.Price > r.prices[order.Symbol].ceil || order.Price < r.prices[order.Symbol].floor {
		return fmt.Errorf("price limit violation")
	}
	return nil
}
