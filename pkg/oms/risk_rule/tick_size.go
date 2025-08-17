package riskrule

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/joripage/orderbook-dev/pkg/oms/model"
)

type tickSizeConfig struct {
	MaxPrice int64 `json:"maxPrice"` // 0 = no limit
	Step     int64 `json:"step"`
}

// TickSizeRule chứa toàn bộ config cho nhiều symbol
type TickSizeRule struct {
	Config map[string][]tickSizeConfig
}

// Load config từ file JSON
func NewTickSizeRuleFromFile(path string) (*TickSizeRule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg map[string][]tickSizeConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &TickSizeRule{Config: cfg}, nil
}

func (r *TickSizeRule) Check(order *model.Order) error {
	rules, ok := r.Config[order.Exchange]
	if !ok { // no config -> no rule
		return nil
	}

	price := int64(order.Price)
	for _, rule := range rules {
		if rule.MaxPrice == 0 || price <= rule.MaxPrice {
			if price%rule.Step != 0 {
				return fmt.Errorf("invalid tick size")
			}
			return nil
		}
	}

	return nil
}
