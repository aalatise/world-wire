package callback

import (
	"github.com/IBM/world-wire/gftn-models/model"
)

type MockNotifyPaymentClient struct {
	NotifyPaymentURL string
	GetCursorURL     string
}

func CreateMockNotifyPaymentClient() (MockNotifyPaymentClient, error) {
	client := MockNotifyPaymentClient{}
	client.NotifyPaymentURL = ""
	client.GetCursorURL = ""
	return client, nil
}

func (client MockNotifyPaymentClient) NotifyPayment(pNotification model.Receive) (error) {
	LOGGER.Infof("In api-service:callback:mock_client:NotifyPayment")
	return nil
}

func (client MockNotifyPaymentClient) GetLastCursor(string) (cursor model.Cursor, err error) {
	cursor.Cursor = "now"
	return cursor, nil
}
