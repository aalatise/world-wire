package pr_client

import "github.ibm.com/gftn/world-wire-services/gftn-models/model"

const (
	PR_ACTIVE    = "active"
	PR_SUSPENDED = "suspended"
	PR_INACTIVE  = "inactive"
)

type PRServiceClient interface {
	GetParticipantForDomain(domain string) (model.Participant, error)
	GetParticipantForIssuingAddress(domain string) (model.Participant, error)
	GetAllParticipants() ([]model.Participant, error)
	GetParticipantsByCountry(countryCode string) ([]model.Participant, error)
	GetParticipantDistAccount(domain string, account string) (string, error)
	GetParticipantIssuingAccount(domain string) (string, error)
	PostParticipantDistAccount(domain string, account model.Account) error
	PostParticipantIssuingAccount(domain string, account model.Account) error
	GetParticipantAccount(domain string, account string) (string, error)
	GetParticipantByAddress(address string) (model.Participant, error)
}
