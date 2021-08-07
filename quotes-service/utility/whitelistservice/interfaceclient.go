package whitelistservice

import "github.com/IBM/world-wire/gftn-models/model"

type InterfaceClient interface {
	GetWhiteListParticipantDomains(participantID string) ([]string, error)
	GetWhiteListParticipants(participantID string) ([]model.Participant, error)
	GetMutualWhiteListParticipants(participantID string) ([]model.Participant, error)
	GetMutualWhiteListParticipantDomains(participantID string) ([]string, error)
}
