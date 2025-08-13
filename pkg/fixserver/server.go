package fixserver

import (
	"log"
	"oms-fix/pkg/orderbook"
)

type Server struct {
	app              *Application
	orderBookManager *orderbook.OrderBookManager
	configFilepath   string
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Init(configFilepath string) error {
	s.configFilepath = configFilepath
	return nil
}

func (s *Server) Start() error {
	app, err := startApp(s.configFilepath, s.orderBookManager)
	if err != nil {
		log.Println("start app err=", err)
		return err
	}
	s.app = app
	return nil
}

func (s *Server) Stop() error {
	if s.app != nil {
		stopApp(s.app)
	}
	return nil
}

func (s *Server) SetOrderbookManager(obm *orderbook.OrderBookManager) {
	s.orderBookManager = obm
}
