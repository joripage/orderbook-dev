package fixgateway

import (
	"context"
	"log"
	"sync"

	"github.com/joripage/orderbook-dev/pkg/oms"
	"github.com/joripage/orderbook-dev/pkg/oms/model"
	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/quickfix"
)

type FixGateway struct {
	cfg         *FixGatewayConfig
	app         *Application
	omsInstance oms.IOMS

	// newOrderSingleMapping     sync.Map
	// orderCancelRequestMapping sync.Map

	requestMapping sync.Map
	sessionMapping sync.Map
}

type FixGatewayConfig struct {
	ConfigFilepath string
}

func NewFixGateway(cfg *FixGatewayConfig) *FixGateway {
	fm := &FixGateway{
		cfg: cfg,
		// newOrderSingleMapping:     sync.Map{},
		// orderCancelRequestMapping: sync.Map{},
		requestMapping: sync.Map{},
		sessionMapping: sync.Map{},
	}

	return fm
}

func (s *FixGateway) AddOmsInstance(o oms.IOMS) {
	s.omsInstance = o
}

func (s *FixGateway) Start(ctx context.Context) error {
	app, err := startApp(s.cfg.ConfigFilepath, s)
	if err != nil {
		log.Printf("start app err=%v", err)
		return err
	}
	s.app = app
	return nil
}

func (s *FixGateway) AddOrder(ctx context.Context, newOrderSingle *NewOrderSingle) {

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

	s.AddRequestToMap(newOrderSingle.ClOrdID, newOrderSingle.SessionID)

	s.omsInstance.AddOrder(ctx, &model.AddOrder{
		GatewayID:  newOrderSingle.ClOrdID,
		Account:    newOrderSingle.Account,
		Symbol:     newOrderSingle.Symbol,
		SecurityID: newOrderSingle.SecurityID,
		// Exchange:     newOrderSingle.Exchange,
		Type:         orderType,
		Price:        newOrderSingle.Price,
		TimeInForce:  timeInForce,
		Side:         side,
		TransactTime: newOrderSingle.TransactTime,
		Quantity:     newOrderSingle.OrderQty,
	})
}

func (s *FixGateway) ModifyOrder(ctx context.Context, req *OrderCancelReplaceRequest) {
	s.AddRequestToMap(req.ClOrdID, req.SessionID)

	s.omsInstance.ModifyOrder(ctx,
		&model.ModifyOrder{
			NewPrice:      req.Price,
			NewQuantity:   req.OrderQty,
			GatewayID:     req.ClOrdID,
			OrigGatewayID: req.OrigClOrdID,
		})
}

func (s *FixGateway) CancelOrder(ctx context.Context, orderCancelRequest *OrderCancelRequest) {
	s.AddRequestToMap(orderCancelRequest.ClOrdID, orderCancelRequest.SessionID)

	s.omsInstance.CancelOrder(ctx, orderCancelRequest.ClOrdID, orderCancelRequest.OrigClOrdID)
}

func (s *FixGateway) OnOrderReport(ctx context.Context, args ...interface{}) {
	if len(args) == 0 {
		return
	}

	if order, ok := args[0].(*model.Order); ok {

		sessionID, err := s.GetRequestByClOrdID(order.GatewayID)
		if err != nil {
			log.Printf("match OrderID=%s not found", order.OrderID)
			return
		}
		// todo: need to review if we need to send via goroutine
		orderBK := *order
		go func() {
			msg := orderReportToExecutionReport(&orderBK)
			quickfix.SendToTarget(msg, *sessionID)
		}()
	}
}
