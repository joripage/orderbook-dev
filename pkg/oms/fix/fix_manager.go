package fixmanager

import (
	"context"
	"log"
	"sync"

	"github.com/joripage/orderbook-dev/pkg/oms"
	"github.com/joripage/orderbook-dev/pkg/oms/model"
	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/quickfix"
)

type FixManager struct {
	cfg         *FixManagerConfig
	app         *Application
	omsInstance oms.IOMS

	orderMapping sync.Map
}

type FixManagerConfig struct {
	ConfigFilepath string
}

func NewFixManager(cfg *FixManagerConfig) *FixManager {
	fm := &FixManager{
		cfg: cfg,
	}

	return fm
}

func (s *FixManager) AddOmsInstance(o oms.IOMS) {
	s.omsInstance = o
}

func (s *FixManager) Start(ctx context.Context) error {
	app, err := startApp(s.cfg.ConfigFilepath, s)
	if err != nil {
		log.Printf("start app err=%v", err)
		return err
	}
	s.app = app
	return nil
}

func (s *FixManager) AddOrder(ctx context.Context, newOrderSingle *NewOrderSingle) {
	orderType := map[enum.OrdType]model.OrderType{
		enum.OrdType_LIMIT:  model.OrderTypeLimit,
		enum.OrdType_MARKET: model.OrderTypeMarket,
		//check iceberg
	}[enum.OrdType(newOrderSingle.OrdType)]
	// var visibleQty int
	if newOrderSingle.MaxFloor.IntPart() != 0 {
		orderType = model.OrderTypeIceberg
		// visibleQty = int(maxFloor.IntPart())
	}

	timeInForce := map[enum.TimeInForce]model.OrderTimeInForce{
		enum.TimeInForce_DAY:                 model.OrderTimeInForceDAY,
		enum.TimeInForce_FILL_OR_KILL:        model.OrderTimeInForceFOK,
		enum.TimeInForce_GOOD_TILL_CANCEL:    model.OrderTimeInForceGTC,
		enum.TimeInForce_IMMEDIATE_OR_CANCEL: model.OrderTimeInForceIOC,
	}[enum.TimeInForce(newOrderSingle.TimeInForce)]

	side := map[enum.Side]model.OrderSide{
		enum.Side_BUY:  model.OrderSideBuy,
		enum.Side_SELL: model.OrderSideSell,
	}[enum.Side(newOrderSingle.Side)]

	s.AddOrderToMap(newOrderSingle)

	s.omsInstance.AddOrder(ctx, &model.AddOrder{
		ID:           newOrderSingle.ClOrdID,
		Account:      newOrderSingle.Account,
		Symbol:       newOrderSingle.Symbol,
		Type:         orderType,
		Price:        newOrderSingle.Price,
		TimeInForce:  timeInForce,
		Side:         side,
		TransactTime: newOrderSingle.TransactTime,
		Quantity:     newOrderSingle.OrderQty,
	})
}

func (s *FixManager) ModifyOrder(ctx context.Context) {

}

func (s *FixManager) CancelOrder(ctx context.Context, orderCancelRequest *OrderCancelRequest) {

}

func (s *FixManager) OnOrderReport(ctx context.Context, args ...interface{}) {
	if len(args) == 0 {
		return
	}

	if order, ok := args[0].(*model.Order); ok {
		newOrderSingle, err := s.GetNewOrderSingleByOrderID(order.GatewayID)
		if err != nil {
			log.Printf("match OrderID=%s not found", order.OrderID)
			return
		}
		// todo: need to review if we need to send via goroutine
		orderBK, newOrnewOrderSingleBK := *order, *newOrderSingle
		go func() {
			msg := orderReportToExecutionReport(&orderBK, &newOrnewOrderSingleBK)
			quickfix.Send(msg)
		}()
	}
}
