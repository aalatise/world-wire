package callback

import "github.com/IBM/world-wire/gftn-models/model"

type NotifyPaymentClient interface {
	NotifyPayment(model.Receive) (error)
	GetLastCursor(string) (cursor model.Cursor, err error)
}

func PostPayment(cl NotifyPaymentClient, pNotification model.Receive) (err error) {
	err = cl.NotifyPayment(pNotification)
	return
}

func GetCursor(cl NotifyPaymentClient, accountName string) (cursor model.Cursor, err error) {
	cursor, err = cl.GetLastCursor(accountName)
	LOGGER.Infof("Received Cursor:%v", cursor.Cursor)
	return
}
