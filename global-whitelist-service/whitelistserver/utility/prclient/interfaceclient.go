package prclient

import "github.com/IBM/world-wire/gftn-models/model"

type InterfaceClient interface {
	GetAllParticipants() ([]model.Participant, error)
}
