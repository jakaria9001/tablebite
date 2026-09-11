package notification

import "context"

type OrderSummary struct {
	OrderNumber string
	TableNumber string
	Body        string
}

type Sender interface {
	Send(ctx context.Context, order OrderSummary) error
}
