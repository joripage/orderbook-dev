package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	fix44nos "github.com/quickfixgo/fix44/newordersingle"
	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/quickfix/log/file"
	"github.com/shopspring/decimal"
)

type InitiatorApp struct {
	sessionID *quickfix.SessionID
}

func (a *InitiatorApp) OnCreate(sessionID quickfix.SessionID) {
	a.sessionID = &sessionID
}

func (a *InitiatorApp) OnLogon(sessionID quickfix.SessionID) {
	log.Println("Logon success", sessionID)
	sendMessageMatchLimit(sessionID)
	// sendMessageMatchMarket(sessionID)
	// sendMessageMatchAmend(sessionID)
	// sendMessageCancelOrder(sessionID)
}

func (a *InitiatorApp) OnLogout(sessionID quickfix.SessionID)                       {}
func (a *InitiatorApp) ToAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) {}
func (a *InitiatorApp) ToApp(msg *quickfix.Message, sessionID quickfix.SessionID) error {
	return nil
}
func (a *InitiatorApp) FromAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	return nil
}
func (a *InitiatorApp) FromApp(msg *quickfix.Message, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	return nil
}

// === Message sender ===
func sendMessageMatchLimit(sessionID quickfix.SessionID) {
	// New(clordid field.ClOrdIDField, side field.SideField, transacttime field.TransactTimeField, ordtype field.OrdTypeField)
	orderBuy := fix44nos.New(
		field.NewClOrdID(""),
		field.NewSide(enum.Side_BUY),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(enum.OrdType_LIMIT))
	orderBuy.SetSymbol("ABC")
	orderBuy.SetAccount("011C399158")
	orderBuy.SetPrice(decimal.NewFromInt(14700), 0)
	orderBuy.SetOrderQty(decimal.NewFromInt(10000), 0)
	orderBuy.SetTimeInForce("0")
	orderBuy.SetSenderCompID(sessionID.SenderCompID)
	orderBuy.SetTargetCompID(sessionID.TargetCompID)
	orderBuy.SetClOrdID(randSeq(17))
	// orderBuy.SetMaxFloor(decimal.NewFromInt(1000), 0)
	err := quickfix.Send(orderBuy)
	log.Println(err)

	orderSell := fix44nos.New(
		field.NewClOrdID(""),
		field.NewSide(enum.Side_SELL),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(enum.OrdType_LIMIT))
	orderSell.SetSymbol("ABC")
	orderSell.SetAccount("011C399157")
	orderSell.SetPrice(decimal.NewFromInt(14700), 0)
	orderSell.SetOrderQty(decimal.NewFromInt(50000), 0)
	orderSell.SetTimeInForce("0")
	orderSell.SetMaxFloor(decimal.NewFromInt(1000), 0)
	orderSell.SetSenderCompID(sessionID.SenderCompID)
	orderSell.SetTargetCompID(sessionID.TargetCompID)
	orderSell.SetClOrdID(randSeq(17))
	err = quickfix.Send(orderSell)
	log.Println(err)
}

func main() {
	cfgPath := os.Args[1]
	log.Println("cfgPath:", cfgPath)
	app := &InitiatorApp{}

	cfg, err := os.Open(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	defer cfg.Close() // nolint

	settings, err := quickfix.ParseSettings(cfg)
	if err != nil {
		log.Fatal(err)
	}

	storeFactory := quickfix.NewMemoryStoreFactory()
	logFactory, _ := file.NewLogFactory(settings)
	initiator, err := quickfix.NewInitiator(app, storeFactory, settings, logFactory)
	if err != nil {
		log.Fatal(err)
	}
	err = initiator.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Initiator started...")
	select {}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
