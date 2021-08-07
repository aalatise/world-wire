package participantregistry

import "github.com/IBM/world-wire/gftn-models/model"

type InterfaceClient interface {
	GetAllParticipants() ([]model.Participant, error)
	GetParticipantForDomain(participantID string) (model.Participant, error)
	GetParticipantAccount(domain string, account string) (string, error)
}
