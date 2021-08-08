package asset

import (
	"os"

	"github.com/shopspring/decimal"

	gasservice "github.com/IBM/world-wire/gas-service-client"
	pr_client "github.com/IBM/world-wire/participant-registry-client/pr-client"

	"github.com/IBM/world-wire/gftn-models/model"
	util "github.com/IBM/world-wire/utility/common"
	global_environment "github.com/IBM/world-wire/utility/global-environment"
	nc "github.com/IBM/world-wire/utility/nodeconfig"
	b "github.com/stellar/go/build"
)

func GetCreditAsset(ast model.Asset, amount decimal.Decimal, prclient pr_client.PRServiceClient) (creditAsset interface{}, err error) {
	creditAsset = b.CreditAmount{}
	if *ast.AssetCode == "xlm" || *ast.AssetCode == "XLM" {
		creditAsset = b.NativeAmount{Amount: amount.Round(7).String()}
	} else {
		astIA, err := prclient.GetParticipantIssuingAccount(ast.IssuerID)
		if err != nil {
			LOGGER.Error("Error getting asset issuing address from participant registry, issuerID: " + ast.IssuerID)
			return creditAsset, err
		}
		creditAsset = b.CreditAmount{Code: *ast.AssetCode, Issuer: astIA,
			Amount: amount.Round(7).String()}
	}
	return creditAsset, nil
}

func CreateNativePayment(gClient gasservice.GasServiceClient, sourceAccount nc.Account, destinationAddress, amount string) (string, error) {

	stellarNetwork := util.GetStellarNetwork(os.Getenv(global_environment.ENV_KEY_STELLAR_NETWORK))

	//Get IBM gas account
	ibmAccount, sequenceNum, err := gClient.GetAccountAndSequence()
	tx, err := b.Transaction(
		b.SourceAccount{AddressOrSeed: ibmAccount},
		stellarNetwork,
		b.Sequence{Sequence: sequenceNum},
		b.Payment(
			b.SourceAccount{AddressOrSeed: sourceAccount.NodeAddress},
			b.Destination{AddressOrSeed: destinationAddress},
			b.NativeAmount{Amount: amount},
		),
	)

	if err != nil {
		return "error during native transaction", err
	}

	txe, err := tx.Sign(sourceAccount.NodeSeed)
	if err != nil {
		return "error while signing native transaction", err
	}

	txeB64, err := txe.Base64()
	if err != nil {
		return "error getting Base64", err
	}

	//Post to gas service
	hash, ledger, err := gClient.SubmitTxe(txeB64)
	if err != nil {
		LOGGER.Warningf("SubmitPaymentTransaction failed... %v ", err.Error())

		return "SubmitPaymentTransaction failed:", err
	}
	LOGGER.Debugf("Hash:%v, Ledger:%v", hash, ledger)

	return "native payment successful :" + hash, nil
}
