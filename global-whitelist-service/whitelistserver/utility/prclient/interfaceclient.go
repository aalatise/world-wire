package prclient

import "github.ibm.com/gftn/world-wire-services/gftn-models/model"

type InterfaceClient interface {
	GetAllParticipants() ([]model.Participant, error)
}
