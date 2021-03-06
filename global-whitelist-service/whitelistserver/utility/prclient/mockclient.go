package prclient

import "github.ibm.com/gftn/world-wire-services/gftn-models/model"

type MockClient struct {
	GetAllParticipantsFunc func() ([]model.Participant, error)
}

func (mPR MockClient) GetAllParticipants() ([]model.Participant, error) {
	return mPR.GetAllParticipantsFunc()
}

func DefaultMock() MockClient {
	return MockClient{
		GetAllParticipantsFunc: func() ([]model.Participant, error) {
			return nil, nil
		},
	}
}
