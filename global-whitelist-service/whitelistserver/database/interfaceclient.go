package database

type InterfaceClient interface {
	DeleteWhitelistParticipant(participantID, wlParticipant string) error
	AddWhitelistParticipant(participant, wlparticipant string) error
	GetWhiteListParicipants(participantID string) ([]string, error)
}
