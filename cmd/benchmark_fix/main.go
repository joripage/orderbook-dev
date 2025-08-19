package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	fix42nos "github.com/quickfixgo/fix42/newordersingle"
	fix42ocrr "github.com/quickfixgo/fix42/ordercancelreplacerequest"
	fix42ocr "github.com/quickfixgo/fix42/ordercancelrequest"
	fix44nos "github.com/quickfixgo/fix44/newordersingle"
	fix44ocrr "github.com/quickfixgo/fix44/ordercancelreplacerequest"
	fix44ocr "github.com/quickfixgo/fix44/ordercancelrequest"
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
	log.Println("Logon success")

	// go sendMessageMatchLimitSoftly(sessionID)
	// go sendMessageMatchLimit(sessionID)
	// go sendMessageMatchLimit44(sessionID)
	// sendMessageMatchMarket(sessionID)
	// sendMessageMatchAmend(sessionID)
	// sendMessageCancelOrder(sessionID)
	// go sendMessageCancelOrder44(sessionID)
	go sendMessageMatchAmend44(sessionID)
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
func sendMessageMatchLimitSoftly(sessionID quickfix.SessionID) {
	total := 250
	// total := 1

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	i := 0
	for t := range ticker.C {
		i += 1
		if i >= 201 {
			break
		}
		fmt.Printf("sending %d at %s\n", total, t.Format("15-04-05"))
		go func() {
			start := time.Now()
			for i := 0; i < total; i++ {
				orderBuy := fix42nos.New(
					field.NewClOrdID(""),
					field.NewHandlInst("1"),
					field.NewSymbol("HCM"),
					field.NewSide(enum.Side_BUY),
					field.NewTransactTime(time.Now()),
					field.NewOrdType(enum.OrdType_LIMIT))
				orderBuy.SetAccount("TMT")
				orderBuy.SetSecurityID("VN000000HCM0")
				orderBuy.SetPrice(decimal.NewFromInt(28000), 0)
				orderBuy.SetOrderQty(decimal.NewFromInt(100), 0)
				orderBuy.SetTimeInForce("0")
				orderBuy.SetSenderCompID(sessionID.SenderCompID)
				orderBuy.SetTargetCompID(sessionID.TargetCompID)
				orderBuy.SetClOrdID(randSeq(17))
				// orderBuy.SetMaxFloor(decimal.NewFromInt(1000), 0)
				err := quickfix.Send(orderBuy)
				_ = err
				// log.Println(err)

				sellID := randSeq(17)
				orderSell := fix42nos.New(
					field.NewClOrdID(""),
					field.NewHandlInst("1"),
					field.NewSymbol("HCM"),
					field.NewSide(enum.Side_SELL),
					field.NewTransactTime(time.Now()),
					field.NewOrdType(enum.OrdType_LIMIT))
				orderSell.SetAccount("TMT")
				orderSell.SetSecurityID("VN000000HCM0")
				orderSell.SetPrice(decimal.NewFromInt(28000), 0)
				orderSell.SetOrderQty(decimal.NewFromInt(100), 0)
				orderSell.SetTimeInForce("0")
				orderSell.SetSenderCompID(sessionID.SenderCompID)
				orderSell.SetTargetCompID(sessionID.TargetCompID)
				orderSell.SetClOrdID(sellID)
				err = quickfix.Send(orderSell)
				_ = err
				// log.Println(err)
				if i == total-1 {
					log.Println(i, sellID)
				}
			}

			elapsed := time.Since(start)
			msgsPerSec := float64(total) / elapsed.Seconds()

			log.Printf("Sent %d messages in %v", total, elapsed)
			log.Printf("Throughput: %.2f messages/sec", msgsPerSec)
		}()
	}
}

func sendMessageMatchLimit(sessionID quickfix.SessionID) {
	total := 250_000
	start := time.Now()

	log.Printf("Sending %d orders", total*2)
	totalBuy, totalSell := int64(0), int64(0)
	min, max := 10, 50
	for i := 0; i < total; i++ {
		// Wait for rate limiter to allow the buy order
		fmt.Println(i)

		buyQty := int64(rand.Intn(max-min) + min)
		totalBuy += buyQty
		orderBuy := fix42nos.New(
			field.NewClOrdID(""),
			field.NewHandlInst("1"),
			field.NewSymbol("VN000000HAG6"),
			field.NewSide(enum.Side_BUY),
			field.NewTransactTime(time.Now()),
			field.NewOrdType(enum.OrdType_LIMIT))
		orderBuy.SetAccount("TMT")
		orderBuy.SetSecurityID("VN000000HAG6")
		orderBuy.SetPrice(decimal.NewFromInt(15600), 0)
		orderBuy.SetOrderQty(decimal.NewFromInt(buyQty), 0)
		orderBuy.SetTimeInForce("0")
		orderBuy.SetSenderCompID(sessionID.SenderCompID)
		orderBuy.SetTargetCompID(sessionID.TargetCompID)
		orderBuy.SetClOrdID(randSeq(17))
		// orderBuy.SetMaxFloor(decimal.NewFromInt(1000), 0)
		sendErr := quickfix.Send(orderBuy)
		log.Println(sendErr)

		sellQty := int64(rand.Intn(max-min) + min)
		totalSell += sellQty
		sellID := randSeq(17)
		orderSell := fix42nos.New(
			field.NewClOrdID(""),
			field.NewHandlInst("1"),
			field.NewSymbol("VN000000HAG6"),
			field.NewSide(enum.Side_SELL),
			field.NewTransactTime(time.Now()),
			field.NewOrdType(enum.OrdType_LIMIT))
		orderSell.SetAccount("TMT")
		orderSell.SetSecurityID("VN000000HAG6")
		orderSell.SetPrice(decimal.NewFromInt(15600), 0)
		orderSell.SetOrderQty(decimal.NewFromInt(sellQty), 0)
		orderSell.SetTimeInForce("0")
		orderSell.SetSenderCompID(sessionID.SenderCompID)
		orderSell.SetTargetCompID(sessionID.TargetCompID)
		orderSell.SetClOrdID(sellID)
		sendErr = quickfix.Send(orderSell)
		log.Println(sendErr)

		if i == total-1 {
			log.Println(i, sellID)
		}

	}

	elapsed := time.Since(start)
	msgsPerSec := float64(total) / elapsed.Seconds()

	log.Printf("Sent %d messages in %v", total, elapsed)
	log.Printf("Throughput: %.2f messages/sec", msgsPerSec)

}

func sendMessageMatchMarket(sessionID quickfix.SessionID) {
	orderBuy := fix42nos.New(
		field.NewClOrdID(""),
		field.NewHandlInst("1"),
		field.NewSymbol("VN000000HAG6"),
		field.NewSide(enum.Side_BUY),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(enum.OrdType_LIMIT))
	orderBuy.SetAccount("011C399158")
	orderBuy.SetPrice(decimal.NewFromInt(13000), 0)
	orderBuy.SetOrderQty(decimal.NewFromInt(1000), 0)
	orderBuy.SetTimeInForce("0")
	orderBuy.SetSenderCompID(sessionID.SenderCompID)
	orderBuy.SetTargetCompID(sessionID.TargetCompID)
	orderBuy.SetClOrdID(randSeq(17))
	err := quickfix.Send(orderBuy)
	log.Println(err)

	orderSell := fix42nos.New(
		field.NewClOrdID(""),
		field.NewHandlInst("1"),
		field.NewSymbol("VN000000HAG6"),
		field.NewSide(enum.Side_SELL),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(enum.OrdType_MARKET))
	orderSell.SetAccount("011C399157")
	orderSell.SetPrice(decimal.NewFromInt(13000), 0)
	orderSell.SetOrderQty(decimal.NewFromInt(500), 0)
	orderSell.SetTimeInForce("9")
	orderSell.SetSenderCompID(sessionID.SenderCompID)
	orderSell.SetTargetCompID(sessionID.TargetCompID)
	orderSell.SetClOrdID(randSeq(17))
	err = quickfix.Send(orderSell)
	log.Println(err)
}

func sendMessageMatchAmend(sessionID quickfix.SessionID) {

	orderBuy := fix42nos.New(
		field.NewClOrdID(""),
		field.NewHandlInst("1"),
		field.NewSymbol("VN000000HAG6"),
		field.NewSide(enum.Side_BUY),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(enum.OrdType_LIMIT))
	orderBuy.SetAccount("011C399158")
	orderBuy.SetPrice(decimal.NewFromInt(13000), 0)
	orderBuy.SetOrderQty(decimal.NewFromInt(1000), 0)
	orderBuy.SetTimeInForce("0")
	orderBuy.SetSenderCompID(sessionID.SenderCompID)
	orderBuy.SetTargetCompID(sessionID.TargetCompID)
	orderBuy.SetClOrdID(randSeq(17))
	err := quickfix.Send(orderBuy)
	log.Println(err)

	orderSellID := randSeq(17)
	orderSell := fix42nos.New(
		field.NewClOrdID(orderSellID),
		field.NewHandlInst("1"),
		field.NewSymbol("VN000000HAG6"),
		field.NewSide(enum.Side_SELL),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(enum.OrdType_LIMIT))
	orderSell.SetAccount("011C399157")
	orderSell.SetPrice(decimal.NewFromInt(13500), 0)
	orderSell.SetOrderQty(decimal.NewFromInt(500), 0)
	orderSell.SetTimeInForce("0")
	orderSell.SetSenderCompID(sessionID.SenderCompID)
	orderSell.SetTargetCompID(sessionID.TargetCompID)
	err = quickfix.Send(orderSell)
	log.Println(err)

	go func() {
		select { // nolint
		case <-time.After(5 * time.Second):
			orderSellReplace := fix42ocrr.New(
				field.NewOrigClOrdID(orderSellID),
				field.NewClOrdID(randSeq(17)),
				field.NewHandlInst("1"),
				field.NewSymbol("VN000000HAG6"),
				field.NewSide(enum.Side_SELL),
				field.NewTransactTime(time.Now()),
				field.NewOrdType(enum.OrdType_LIMIT))
			orderSellReplace.SetAccount("011C399157")
			orderSellReplace.SetPrice(decimal.NewFromInt(13000), 0)
			orderSellReplace.SetOrderQty(decimal.NewFromInt(500), 0)
			orderSellReplace.SetTimeInForce("0")
			orderSellReplace.SetSenderCompID(sessionID.SenderCompID)
			orderSellReplace.SetTargetCompID(sessionID.TargetCompID)
			err = quickfix.Send(orderSellReplace)
			log.Println(err)
		}
	}()
}

func sendMessageCancelOrder(sessionID quickfix.SessionID) {
	clOrderID := randSeq(17)
	orderBuy := fix42nos.New(
		field.NewClOrdID(clOrderID),
		field.NewHandlInst("1"),
		field.NewSymbol("VN000000HAG6"),
		field.NewSide(enum.Side_BUY),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(enum.OrdType_LIMIT))
	orderBuy.SetAccount("011C399158")
	orderBuy.SetPrice(decimal.NewFromInt(13000), 0)
	orderBuy.SetOrderQty(decimal.NewFromInt(1000), 0)
	orderBuy.SetTimeInForce("0")
	orderBuy.SetSenderCompID(sessionID.SenderCompID)
	orderBuy.SetTargetCompID(sessionID.TargetCompID)
	err := quickfix.Send(orderBuy)
	log.Println(err)

	go func() {
		select { // nolint
		case <-time.After(5 * time.Second):
			orderBuyCancel := fix42ocr.New(
				field.NewOrigClOrdID(clOrderID),
				field.NewClOrdID(randSeq(17)),
				field.NewSymbol("VN000000HAG6"),
				field.NewSide(enum.Side_BUY),
				field.NewTransactTime(time.Now()))
			orderBuyCancel.SetOrderQty(decimal.NewFromInt(1000), 0)
			orderBuyCancel.SetSenderCompID(sessionID.SenderCompID)
			orderBuyCancel.SetTargetCompID(sessionID.TargetCompID)
			err = quickfix.Send(orderBuyCancel)
			log.Println(err)
		}
	}()
}

// fix 4.4

func sendMessageMatchLimit44(sessionID quickfix.SessionID) {
	total := 125_000 // 500_000 / 4
	// total = 62500    // 500_000 / 8
	// total = 31250    // 500_000 / 16
	start := time.Now()

	for i := 0; i < total; i++ {

		orderBuy := fix44nos.New(
			field.NewClOrdID(""),
			field.NewSide(enum.Side_BUY),
			field.NewTransactTime(time.Now()),
			field.NewOrdType(enum.OrdType_LIMIT))
		orderBuy.SetAccount("TMT")
		orderBuy.SetSymbol("VN000000HAG6")
		orderBuy.SetSecurityID("VN000000HAG6")
		orderBuy.SetPrice(decimal.NewFromInt(15600), 0)
		orderBuy.SetOrderQty(decimal.NewFromInt(100), 0)
		orderBuy.SetTimeInForce("0")
		orderBuy.SetSenderCompID(sessionID.SenderCompID)
		orderBuy.SetTargetCompID(sessionID.TargetCompID)
		orderBuy.SetClOrdID(randSeq(17))
		// orderBuy.SetMaxFloor(decimal.NewFromInt(1000), 0)

		go func() {
			sendErr := quickfix.Send(orderBuy)
			_ = sendErr
			// log.Println(sendErr)
		}()

		sellID := randSeq(17)
		orderSell := fix44nos.New(
			field.NewClOrdID(""),
			field.NewSide(enum.Side_SELL),
			field.NewTransactTime(time.Now()),
			field.NewOrdType(enum.OrdType_LIMIT))
		orderSell.SetAccount("TMT")
		orderSell.SetSymbol("VN000000HAG6")
		orderSell.SetSecurityID("VN000000HAG6")
		orderSell.SetPrice(decimal.NewFromInt(15600), 0)
		orderSell.SetOrderQty(decimal.NewFromInt(100), 0)
		orderSell.SetTimeInForce("0")
		orderSell.SetSenderCompID(sessionID.SenderCompID)
		orderSell.SetTargetCompID(sessionID.TargetCompID)
		orderSell.SetClOrdID(sellID)
		go func() {
			sendErr := quickfix.Send(orderSell)
			_ = sendErr
			// log.Println(sendErr)
			if i == total-1 {
				log.Println(i, sellID)
			}
		}()

	}

	elapsed := time.Since(start)
	msgsPerSec := float64(total) / elapsed.Seconds()

	log.Printf("Sent %d messages in %v", total, elapsed)
	log.Printf("Throughput: %.2f messages/sec", msgsPerSec)

}

func sendMessageCancelOrder44(sessionID quickfix.SessionID) {
	buyClOrderID := randSeq(17)
	orderBuy := fix44nos.New(
		field.NewClOrdID(""),
		field.NewSide(enum.Side_BUY),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(enum.OrdType_LIMIT))
	orderBuy.SetAccount("TMT")
	orderBuy.SetSymbol("VN000000HAG6")
	orderBuy.SetSecurityID("VN000000HAG6")
	orderBuy.SetPrice(decimal.NewFromInt(15600), 0)
	orderBuy.SetOrderQty(decimal.NewFromInt(100), 0)
	orderBuy.SetTimeInForce("0")
	orderBuy.SetSenderCompID(sessionID.SenderCompID)
	orderBuy.SetTargetCompID(sessionID.TargetCompID)
	orderBuy.SetClOrdID(buyClOrderID)

	err := quickfix.Send(orderBuy)
	log.Println(err)

	sellID := randSeq(17)
	orderSell := fix44nos.New(
		field.NewClOrdID(""),
		field.NewSide(enum.Side_SELL),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(enum.OrdType_LIMIT))
	orderSell.SetAccount("TMT")
	orderSell.SetSymbol("VN000000HAG6")
	orderSell.SetSecurityID("VN000000HAG6")
	orderSell.SetPrice(decimal.NewFromInt(15600), 0)
	orderSell.SetOrderQty(decimal.NewFromInt(20), 0)
	orderSell.SetTimeInForce("0")
	orderSell.SetSenderCompID(sessionID.SenderCompID)
	orderSell.SetTargetCompID(sessionID.TargetCompID)
	orderSell.SetClOrdID(sellID)

	err = quickfix.Send(orderSell)
	log.Println(err)

	go func() {
		select { // nolint
		case <-time.After(5 * time.Second):
			orderBuyCancel := fix44ocr.New(
				field.NewOrigClOrdID(buyClOrderID),
				field.NewClOrdID(randSeq(17)),
				field.NewSide(enum.Side_BUY),
				field.NewTransactTime(time.Now()))
			orderBuyCancel.SetAccount("TMT")
			orderBuyCancel.SetSymbol("VN000000HAG6")
			orderBuyCancel.SetSecurityID("VN000000HAG6")
			orderBuyCancel.SetOrderQty(decimal.NewFromInt(80), 0)
			orderBuyCancel.SetSenderCompID(sessionID.SenderCompID)
			orderBuyCancel.SetTargetCompID(sessionID.TargetCompID)
			err = quickfix.Send(orderBuyCancel)
			log.Println(err)
		}
	}()
}

func sendMessageMatchAmend44(sessionID quickfix.SessionID) {

	orderBuy := fix44nos.New(
		field.NewClOrdID(""),
		field.NewSide(enum.Side_BUY),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(enum.OrdType_LIMIT))
	orderBuy.SetAccount("011C399158")
	orderBuy.SetPrice(decimal.NewFromInt(13000), 0)
	orderBuy.SetOrderQty(decimal.NewFromInt(1000), 0)
	orderBuy.SetTimeInForce("0")
	orderBuy.SetSenderCompID(sessionID.SenderCompID)
	orderBuy.SetTargetCompID(sessionID.TargetCompID)
	orderBuy.SetClOrdID(randSeq(17))
	err := quickfix.Send(orderBuy)
	log.Println(err)

	orderSellID := randSeq(17)
	orderSell := fix44nos.New(
		field.NewClOrdID(orderSellID),
		field.NewSide(enum.Side_SELL),
		field.NewTransactTime(time.Now()),
		field.NewOrdType(enum.OrdType_LIMIT))
	orderSell.SetAccount("011C399157")
	orderSell.SetPrice(decimal.NewFromInt(13500), 0)
	orderSell.SetOrderQty(decimal.NewFromInt(500), 0)
	orderSell.SetTimeInForce("0")
	orderSell.SetSenderCompID(sessionID.SenderCompID)
	orderSell.SetTargetCompID(sessionID.TargetCompID)
	err = quickfix.Send(orderSell)
	log.Println(err)

	go func() {
		select { // nolint
		case <-time.After(5 * time.Second):
			orderSellReplace := fix44ocrr.New(
				field.NewOrigClOrdID(orderSellID),
				field.NewClOrdID(randSeq(17)),
				field.NewSide(enum.Side_SELL),
				field.NewTransactTime(time.Now()),
				field.NewOrdType(enum.OrdType_LIMIT))
			orderSellReplace.SetAccount("011C399157")
			orderSellReplace.SetPrice(decimal.NewFromInt(13000), 0)
			orderSellReplace.SetOrderQty(decimal.NewFromInt(500), 0)
			orderSellReplace.SetTimeInForce("0")
			orderSellReplace.SetSenderCompID(sessionID.SenderCompID)
			orderSellReplace.SetTargetCompID(sessionID.TargetCompID)
			err = quickfix.Send(orderSellReplace)
			log.Println(err)
		}
	}()
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

	appSettings, err := quickfix.ParseSettings(cfg)
	if err != nil {
		log.Fatal(err)
	}

	storeFactory := quickfix.NewMemoryStoreFactory()
	logFactory, _ := file.NewLogFactory(appSettings)

	initiator, err := quickfix.NewInitiator(app, storeFactory, appSettings, logFactory)
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
