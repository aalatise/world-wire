package whitelistclient

import "github.com/IBM/world-wire/gftn-models/model"

type InterfaceClient interface {
	GetWhiteListParticipantDomains(participantID string) ([]string, error)
	GetWhiteListParticipants(participantID string) ([]model.Participant, error)
	CreateWhiteListParticipants(participantID, wlparitcipantID string) error
	IsParticipantWhiteListed(participantID string, targetDomain string) (bool, error)
	DeleteWhiteListParticipants(participantID, wlparitcipantID string) error
	GetMutualWhiteListParticipantDomains(participantID string) ([]string, error)
	GetMutualWhiteListParticipants(participantID string) ([]model.Participant, error)
}
