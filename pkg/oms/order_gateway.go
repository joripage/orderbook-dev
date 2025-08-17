package oms

import (
	"context"
)

type OrderGateway interface {
	Start(ctx context.Context) error

	// oms to client
	OnOrderReport(ctx context.Context, args ...interface{})
}
