package whitelistservice

import "github.ibm.com/gftn/world-wire-services/gftn-models/model"

type InterfaceClient interface {
	GetWhiteListParticipantDomains(participantID string) ([]string, error)
	GetWhiteListParticipants(participantID string) ([]model.Participant, error)
	GetMutualWhiteListParticipants(participantID string) ([]model.Participant, error)
	GetMutualWhiteListParticipantDomains(participantID string) ([]string, error)
}
