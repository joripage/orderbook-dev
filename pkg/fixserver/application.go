package fixserver

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"oms-fix/pkg/oms"
	"oms-fix/pkg/oms/model"
	"oms-fix/pkg/orderbook"
	"os"
	"time"

	"github.com/joripage/go_util/pkg/shardqueue"
	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/fix44/newordersingle"
	"github.com/quickfixgo/fix44/ordercancelreplacerequest"
	"github.com/quickfixgo/fix44/ordercancelrequest"
	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/quickfix/log/file"
	"github.com/quickfixgo/tag"
)

// Application implements the quickfix.Application interface
type Application struct {
	*quickfix.MessageRouter
	cfg        AppConfig
	quickEvent chan bool
	dispatcher chan *inboundMsg
	shardQueue *shardqueue.Shardqueue
	// orderBookManager *orderbook.OrderBookManager
	orderManager oms.OrderManager
}

type AppConfig struct {
	enableQueue      bool
	enableShardQueue bool
}

type inboundMsg struct {
	msg       *quickfix.Message
	sessionID quickfix.SessionID
}

const (
	numShards = 16
	queueSize = 1_000_000
)

func newApplication(cfg AppConfig, obm *orderbook.OrderBookManager) *Application {
	fm := oms.NewFixManager()
	app := &Application{
		MessageRouter: quickfix.NewMessageRouter(),
		// orderBookManager: obm,
		orderManager: fm,
		cfg:          cfg,
		quickEvent:   make(chan bool, 1),
	}

	app.AddRoute(newordersingle.Route(app.onNewOrderSingle))
	app.AddRoute(ordercancelrequest.Route(app.onOrderCancelRequest))
	app.AddRoute(ordercancelreplacerequest.Route(app.onOrderCancelReplaceRequest))

	if app.cfg.enableShardQueue {
		app.shardQueue = shardqueue.NewShardQueue(numShards, queueSize)
		app.shardQueue.Start(func(msg interface{}) error {
			if v, ok := msg.(*inboundMsg); ok {
				app.Route(v.msg, v.sessionID)
			}
			return nil
		})
	} else if app.cfg.enableQueue {
		app.dispatcher = make(chan *inboundMsg, queueSize)
		go app.runDispatcher()
	}

	return app
}

func startApp(config_filepath string, obm *orderbook.OrderBookManager) (*Application, error) {
	var cfgFileName = config_filepath

	cfg, err := os.Open(cfgFileName)
	if err != nil {
		return nil, fmt.Errorf("error opening %v, %v", cfgFileName, err)
	}
	defer cfg.Close() // nolint

	stringData, readErr := io.ReadAll(cfg)
	if readErr != nil {
		return nil, fmt.Errorf("error reading cfg: %s,", readErr)
	}

	appSettings, err := quickfix.ParseSettings(bytes.NewReader(stringData))
	if err != nil {
		return nil, fmt.Errorf("error reading cfg: %s,", err)
	}

	app := newApplication(AppConfig{
		enableQueue: true,
		// enableShardQueue: true,
	}, obm)
	logFactory, _ := file.NewLogFactory(appSettings)
	acceptor, err := quickfix.NewAcceptor(app, quickfix.NewMemoryStoreFactory(), appSettings, logFactory)
	if err != nil {
		return nil, fmt.Errorf("unable to create acceptor: %s", err)
	}

	err = acceptor.Start()
	if err != nil {
		return nil, fmt.Errorf("unable to start FIX acceptor: %s", err)
	}

	go func() {
		<-app.quickEvent
		acceptor.Stop()
	}()

	return app, nil
}

func stopApp(a *Application) {
	select {
	case a.quickEvent <- true:
	default:
	}
}

// OnCreate implemented as part of Application interface
func (a Application) OnCreate(sessionID quickfix.SessionID) {}

// OnLogon implemented as part of Application interface
func (a Application) OnLogon(sessionID quickfix.SessionID) {}

// OnLogout implemented as part of Application interface
func (a Application) OnLogout(sessionID quickfix.SessionID) {}

// ToAdmin implemented as part of Application interface
func (a Application) ToAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) {}

// ToApp implemented as part of Application interface
func (a Application) ToApp(msg *quickfix.Message, sessionID quickfix.SessionID) error {
	return nil
}

// FromAdmin implemented as part of Application interface
func (a Application) FromAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	return nil
}

// FromApp implemented as part of Application interface, uses Router on incoming application messages
func (a *Application) FromApp(msg *quickfix.Message, sessionID quickfix.SessionID) (reject quickfix.MessageRejectError) {

	if a.cfg.enableShardQueue {
		a.shardQueue.Shard(getRoutingKey(msg, sessionID), &inboundMsg{msg, sessionID})
		return nil
	} else if a.cfg.enableQueue {
		a.dispatcher <- &inboundMsg{msg, sessionID}
		return nil
	}

	return a.Route(msg, sessionID)
}

func getRoutingKey(msg *quickfix.Message, sessionID quickfix.SessionID) string {
	if clOrdID, err := msg.Body.GetString(tag.ClOrdID); err == nil && clOrdID != "" {
		return clOrdID
	}

	if msgType, err := msg.Header.GetString(tag.MsgType); err == nil {
		return "MSGTYPE:" + msgType
	}

	return sessionID.String()
}

func (a *Application) runDispatcher() {
	for msg := range a.dispatcher {
		if err := a.Route(msg.msg, msg.sessionID); err != nil {
			log.Println("Route error", err)
		}
	}
}

func (a *Application) onNewOrderSingle(msg newordersingle.NewOrderSingle, sessionID quickfix.SessionID) quickfix.MessageRejectError {

	// clOrdID, _ := msg.GetClOrdID()
	symbol, _ := msg.GetSymbol()
	side, _ := msg.GetSide()
	ordType, _ := msg.GetOrdType()
	price, _ := msg.GetPrice()
	orderQty, _ := msg.GetOrderQty()
	timeInForce, _ := msg.GetTimeInForce()
	maxFloor, _ := msg.GetMaxFloor()

	obOrderType := map[enum.OrdType]model.OrderType{
		enum.OrdType_LIMIT:  model.OrderTypeLimit,
		enum.OrdType_MARKET: model.OrderTypeMarket,
		//check iceberg
	}[ordType]
	// var visibleQty int
	if maxFloor.IntPart() != 0 {
		obOrderType = model.OrderTypeIceberg
		// visibleQty = int(maxFloor.IntPart())
	}

	obTimeInForce := map[enum.TimeInForce]model.OrderTimeInForce{
		enum.TimeInForce_DAY:                 model.OrderTimeInForceDAY,
		enum.TimeInForce_FILL_OR_KILL:        model.OrderTimeInForceFOK,
		enum.TimeInForce_GOOD_TILL_CANCEL:    model.OrderTimeInForceGTC,
		enum.TimeInForce_IMMEDIATE_OR_CANCEL: model.OrderTimeInForceIOC,
	}[timeInForce]

	obSide := map[enum.Side]model.OrderSide{
		enum.Side_BUY:  model.OrderSideBuy,
		enum.Side_SELL: model.OrderSideSell,
	}[side]

	// a.orderBookManager.AddOrder(&orderbook.Order{
	// 	ID:     clOrdID,
	// 	Symbol: symbol,
	// 	Side: map[enum.Side]orderbook.Side{
	// 		enum.Side_BUY:  orderbook.BUY,
	// 		enum.Side_SELL: orderbook.SELL,
	// 	}[side],
	// 	Price:       price.InexactFloat64(),
	// 	Qty:         int(orderQty.IntPart()),
	// 	Type:        obOrderType,
	// 	TimeInForce: obTimeInForce,
	// 	VisibleQty:  visibleQty,
	// })
	a.orderManager.AddOrder(&model.AddOrder{
		Account:      "account1",
		Symbol:       symbol,
		Type:         obOrderType,
		Price:        price,
		TimeInForce:  obTimeInForce,
		Side:         obSide,
		TransactTime: time.Now(),
		Quantity:     orderQty,
	})
	return nil
}

func (a *Application) onOrderCancelRequest(msg ordercancelrequest.OrderCancelRequest, sessionID quickfix.SessionID) quickfix.MessageRejectError {

	return nil
}

func (a *Application) onOrderCancelReplaceRequest(msg ordercancelreplacerequest.OrderCancelReplaceRequest, sessionID quickfix.SessionID) quickfix.MessageRejectError {

	return nil
}

// func (a *Application) sendExecutionReport(tradeResult *orderbook.MatchResult) quickfix.MessageRejectError {
// 	execReportMsg := executionreport.New(
// 		field.NewOrderID(report.OrderID),
// 		field.NewExecID(report.ExecID),
// 		field.NewExecTransType(enum.ExecTransType(report.ExecTransType)),
// 		field.NewExecType(enum.ExecType(report.ExecType)),
// 		field.NewOrdStatus(enum.OrdStatus(report.OrdStatus)),
// 		field.NewSymbol(report.Symbol),
// 		field.NewSide(enum.Side(report.Side)),
// 		field.NewLeavesQty(report.LeavesQty, 2),
// 		field.NewCumQty(report.CumQty, 2),
// 		field.NewAvgPx(report.AvgPx, 2),
// 	)

// 	execReportMsg.SetClOrdID(report.ClOrdID)
// 	execReportMsg.SetAccount(report.Account)
// 	execReportMsg.SetOrderQty(report.OrderQty, 0)
// 	execReportMsg.SetPrice(report.Price, 0)
// 	execReportMsg.SetTimeInForce(enum.TimeInForce(report.TimeInForce))
// 	execReportMsg.SetTransactTime(report.TransactTime)
// 	execReportMsg.SetText(report.Text)
// 	execReportMsg.SetMaturityMonthYear(report.MaturityMonthYear)
// 	execReportMsg.SetSecurityType(enum.SecurityType(report.SecurityType))

// 	execReportMsg.SetTargetCompID(report.SenderCompID)
// 	execReportMsg.SetSenderCompID(report.TargetCompID)

// 	return quickfix.Send(execReportMsg)
// }
