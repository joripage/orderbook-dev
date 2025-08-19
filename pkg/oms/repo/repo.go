package repo

import (
	"gorm.io/gorm"
)

type IRepo interface {
	Order() IOrder
	OrderEvent() IOrderEvent
}

type Repo struct {
	omsDB *gorm.DB
}

func NewRepo(omsDB *gorm.DB) IRepo {
	return &Repo{
		omsDB: omsDB,
	}
}

func (r *Repo) Order() IOrder {
	return NewOrderSQLRepo(r.omsDB)
}

func (r *Repo) OrderEvent() IOrderEvent {
	return NewOrderEventSQLRepo(r.omsDB)
}
