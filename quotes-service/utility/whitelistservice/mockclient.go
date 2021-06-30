package whitelistservice

import (
	"net/http"

	"github.ibm.com/gftn/world-wire-services/gftn-models/model"
	"github.ibm.com/gftn/world-wire-services/quotes-service/utility/participantregistry"
)

type MockClient struct {
	HTTP  *http.Client
	WLURL string
}

func (client *MockClient) GetWhiteListParticipantDomains(participantOFIID string) ([]string, error) {
	return []string{"hk.one.payments.worldwire.io", "ie.one.payments.worldwire.io", "nz.one.payments.gftn.io", "au.one.payments.gftn.io"}, nil
}

func (client *MockClient) GetWhiteListParticipants(participantOFIID string) ([]model.Participant, error) {
	var whiteLists []string
	if participantOFIID == "hk.one.payments.worldwire.io" {
		whiteLists = []string{"hk.one.payments.worldwire.io", "ie.one.payments.worldwire.io", "au.one.payments.gftn.io"}
	}
	if participantOFIID == "ie.one.payments.worldwire.io" {
		whiteLists = []string{"hk.one.payments.worldwire.io", "ie.one.payments.worldwire.io", "nz.one.payments.gftn.io", "au.one.payments.gftn.io"}
	}

	var whiteListParticipants []model.Participant
	prc, _ := participantregistry.CreateMockClient()
	participants, _ := prc.GetAllParticipants()
	for idx, participant := range participants {
		for _, whiteList := range whiteLists {
			if *participant.ID == whiteList {
				whiteListParticipants = append(whiteListParticipants, participants[idx])
			}
		}
	}
	return whiteListParticipants, nil

}
