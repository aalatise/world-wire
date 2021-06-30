package client

type PaymentListenerClient interface {
	SubscribePayments(distAccountName string) (err error)
}
