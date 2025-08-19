package fixgateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/joripage/go_util/pkg/shardqueue"
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

	fixGateway *FixGateway
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
	numShards = 64
	queueSize = 1_000_000
)

func newApplication(cfg AppConfig, fixGateway *FixGateway) *Application {
	app := &Application{
		MessageRouter: quickfix.NewMessageRouter(),
		cfg:           cfg,
		quickEvent:    make(chan bool, 1),
		fixGateway:    fixGateway,
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

func startApp(config_filepath string, fixGateway *FixGateway) (*Application, error) {
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
		// enableQueue: true,
		enableShardQueue: true,
	}, fixGateway)

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
	// senderCompID, _ := msg.GetSenderCompID()
	// senderSubID, _ := msg.GetSenderSubID()
	// targetCompID, _ := msg.GetTargetCompID()
	// onBehalfOfCompID, _ := msg.GetOnBehalfOfCompID()
	// deliverToCompID, _ := msg.GetDeliverToCompID()

	clOrdID, _ := msg.GetClOrdID()
	symbol, _ := msg.GetSymbol()
	side, _ := msg.GetSide()
	ordType, _ := msg.GetOrdType()
	price, _ := msg.GetPrice()
	orderQty, _ := msg.GetOrderQty()
	account, _ := msg.GetAccount()
	accountType, _ := msg.GetAccountType()
	timeInForce, _ := msg.GetTimeInForce()
	transactTime, _ := msg.GetTransactTime()
	maturityMonthYear, _ := msg.GetMaturityMonthYear()
	securityType, _ := msg.GetSecurityType()
	securityID, _ := msg.GetSecurityID()
	maxFloor, _ := msg.GetMaxFloor()

	m := &NewOrderSingle{
		SessionID: &sessionID,
		// SenderCompID:     senderCompID,
		// SenderSubID:      senderSubID,
		// TargetCompID:     targetCompID,
		// OnBehalfOfCompID: onBehalfOfCompID,
		// DeliverToCompID:  deliverToCompID,

		Account:           account,
		AccountType:       accountType,
		ClOrdID:           clOrdID,
		Symbol:            symbol,
		OrdType:           ordType,
		Price:             price,
		TimeInForce:       timeInForce,
		Side:              side,
		TransactTime:      transactTime,
		OrderQty:          orderQty,
		MaturityMonthYear: maturityMonthYear,
		SecurityType:      securityType,
		SecurityID:        securityID,
		MaxFloor:          maxFloor,
	}
	a.fixGateway.AddOrder(context.Background(), m)

	return nil
}

func (a *Application) onOrderCancelRequest(msg ordercancelrequest.OrderCancelRequest, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	// senderCompID, _ := msg.GetSenderCompID()
	// senderSubID, _ := msg.GetSenderSubID()
	// targetCompID, _ := msg.GetTargetCompID()
	// onBehalfOfCompID, _ := msg.GetOnBehalfOfCompID()
	// deliverToCompID, _ := msg.GetDeliverToCompID()

	origClOrdID, _ := msg.GetOrigClOrdID()
	clOrdID, _ := msg.GetClOrdID()
	account, _ := msg.GetAccount()
	symbol, _ := msg.GetSymbol()
	side, _ := msg.GetSide()
	transactTime, _ := msg.GetTransactTime()
	maturityMonthYear, _ := msg.GetMaturityMonthYear()
	securityType, _ := msg.GetSecurityType()
	securityID, _ := msg.GetSecurityID()

	m := &OrderCancelRequest{
		SessionID: &sessionID,
		// SenderCompID:     senderCompID,
		// SenderSubID:      senderSubID,
		// TargetCompID:     targetCompID,
		// OnBehalfOfCompID: onBehalfOfCompID,
		// DeliverToCompID:  deliverToCompID,

		OrigClOrderID:     origClOrdID,
		ClOrderID:         clOrdID,
		Account:           account,
		Symbol:            symbol,
		Side:              side,
		TransactTime:      transactTime,
		MaturityMonthYear: maturityMonthYear,
		SecurityType:      securityType,
		SecurityID:        securityID,
	}
	a.fixGateway.CancelOrder(context.Background(), m)

	return nil
}

func (a *Application) onOrderCancelReplaceRequest(msg ordercancelreplacerequest.OrderCancelReplaceRequest, sessionID quickfix.SessionID) quickfix.MessageRejectError {

	// senderCompID, _ := msg.GetSenderCompID()
	// senderSubID, _ := msg.GetSenderSubID()
	// targetCompID, _ := msg.GetTargetCompID()
	// onBehalfOfCompID, _ := msg.GetOnBehalfOfCompID()
	// deliverToCompID, _ := msg.GetDeliverToCompID()
	// msgSeqNum, _ := msg.GetMsgSeqNum()

	origClOrderID, _ := msg.GetOrigClOrdID()
	clOrderID, _ := msg.GetClOrdID()
	account, _ := msg.GetAccount()
	symbol, _ := msg.GetSymbol()
	side, _ := msg.GetSide()
	transactTime, _ := msg.GetTransactTime()
	orderQty, _ := msg.GetOrderQty()
	ordType, _ := msg.GetOrdType()
	price, _ := msg.GetPrice()
	timeInForce, _ := msg.GetTimeInForce()
	maturityMonthYear, _ := msg.GetMaturityMonthYear()
	securityType, _ := msg.GetSecurityType()
	securityID, _ := msg.GetSecurityID()

	m := &OrderCancelReplaceRequest{
		OrigClOrderID:     origClOrderID,
		ClOrderID:         clOrderID,
		Account:           account,
		Symbol:            symbol,
		Side:              side,
		TransactTime:      transactTime,
		OrderQty:          orderQty,
		OrdType:           ordType,
		Price:             price,
		TimeInForce:       timeInForce,
		MaturityMonthYear: maturityMonthYear,
		SecurityType:      securityType,
		SecurityID:        securityID,
	}
	// err := a.oms.OnOrderCancelReplaceRequest(m)
	// if err != nil {
	// 	a.logger.Error("OnOrderCancelReplaceRequest", "error", err)
	// }
	a.fixGateway.ModifyOrder(context.Background(), m)

	return nil
}
