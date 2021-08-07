package client

import (
	pr_client "github.com/IBM/world-wire/participant-registry-client/pr-client"
	"github.com/IBM/world-wire/payment/constant"
	"github.com/IBM/world-wire/utility/common"
)

func GetParticipantAccount(prServiceURL, homeDomain, queryStr string) *string {
	prc, prcErr := pr_client.CreateRestPRServiceClient(prServiceURL)
	if prcErr != nil {
		LOGGER.Error("Can not create connection to PR client service, please check if PR service is running")
		return nil
	}

	pr, prcGetErr := prc.GetParticipantForDomain(homeDomain)
	if prcGetErr != nil {
		LOGGER.Error("Could not found participant from PR service")
		return nil
	}

	if pr.Status == "active" {
		if queryStr == common.ISSUING {
			return &pr.IssuingAccount
		} else if queryStr == constant.BIC_STRING {
			return pr.Bic
		} else {
			for _, oa := range pr.OperatingAccounts {
				if oa.Name == queryStr {
					return oa.Address
				}
			}
		}
	} else {
		LOGGER.Errorf("Participant status is inactive")
		return nil
	}

	return nil
}

func GetParticipantRole(prServiceURL, homeDomain string) *string {
	prc, prcErr := pr_client.CreateRestPRServiceClient(prServiceURL)
	if prcErr != nil {
		LOGGER.Error("Can not create connection to PR client service, please check if PR service is running")
		return nil
	}

	pr, prcGetErr := prc.GetParticipantForDomain(homeDomain)
	if prcGetErr != nil {
		LOGGER.Error("Could not found participant from PR service")
		return nil
	}

	if pr.Status == "active" {
		return pr.Role
	} else {
		LOGGER.Errorf("Participant status is inactive")
		return nil
	}

	return nil
}
