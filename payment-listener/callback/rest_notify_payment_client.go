package callback

import (
	"encoding/json"

	"github.com/go-resty/resty"
	"github.com/IBM/world-wire/gftn-models/model"
)

type RestNotifyPaymentClient struct {
	GetCursorURL string
}

func (client RestNotifyPaymentClient) GetLastCursor(accountName string) (cursor model.Cursor, err error) {

	cursor = model.Cursor{}
	LOGGER.Infof("In api-service:callback:GetLastCursor:GetLastCursor")

	response, err := resty.R().SetHeader("Content-type", "application/json").Get(client.GetCursorURL + "/" + accountName + "/cursor")
	if err != nil {
		LOGGER.Errorf("Error while making request to Notify Payment:  %v", err)
		return
	}

	if response.StatusCode() != 200 {
		LOGGER.Infof("Notify Payment returned a non-200 response")
		return
	}

	responseBody := response.Body()
	err = json.Unmarshal(responseBody, &cursor)

	if err != nil {
		LOGGER.Warningf("Error while unmarshalling callback GetLastCursor:  %v", err)
		return cursor, err
	}

	return cursor, nil
}
