package common

import (
	"errors"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/miekg/pkcs11"
	"github.com/stellar/go/keypair"
	"github.com/IBM/world-wire/crypto-service/environment"
	"github.com/IBM/world-wire/crypto-service/utility/constant"
	"github.com/IBM/world-wire/crypto-service/utility/hsm"

	"github.com/IBM/world-wire/crypto-service/utility/hsmclient"
	"github.com/IBM/world-wire/utility/nodeconfig"
)

type HsmObject struct {
	C       *hsm.Crypto
	Slot    uint
	Session pkcs11.SessionHandle
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func RandStringBytes(n int) string {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[r1.Intn(len(letterBytes))]
	}
	return string(b)
}

//create a new account using either HSM or Stellar SDK
func (obj *HsmObject) GenericGenerateAccount() (nodeconfig.Account, error) {
	source := strings.ToUpper(os.Getenv(environment.ENV_KEY_ACCOUNT_SOURCE))
	if source == constant.ACCOUNT_FROM_STELLAR {
		fullkey, err := keypair.Random()
		if err != nil {
			LOGGER.Errorf("Error Creating a new keypair: %v", err)
		}
		return nodeconfig.Account{NodeAddress: fullkey.Address(), NodeSeed: fullkey.Seed(), PrivateKeyLabel: "null", PublicKeyLabel: "null"}, nil
	} else if source == constant.ACCOUNT_FROM_HSM {
		publicKeyLabel, privateKeyLabel, _, _, ecPoints, err := hsmclient.GenerateKeyPair(obj.C, obj.Slot, RandStringBytes(15), RandStringBytes(15))
		if err != nil {
			LOGGER.Errorf("Error while creating account, re-initializing")
			obj.C, obj.Session, err = ReinitializeHSM(obj.C, obj.Session)
			if err != nil {
				LOGGER.Errorf("Encounter error while re-intializing HSM login session: %v", err)
				return nodeconfig.Account{}, err
			}
			publicKeyLabel, privateKeyLabel, _, _, ecPoints, err = hsmclient.GenerateKeyPair(obj.C, obj.Slot, RandStringBytes(15), RandStringBytes(15))
			if err != nil {
				return nodeconfig.Account{}, err
			}
		}
		stellarAccount := hsmclient.GenerateStellarAccount(ecPoints)
		// vault safe cannot accept empty string, so we will input `null`
		return nodeconfig.Account{NodeAddress: stellarAccount, NodeSeed: "null", PublicKeyLabel: publicKeyLabel, PrivateKeyLabel: privateKeyLabel}, nil
	}
	return nodeconfig.Account{}, errors.New("Cannot fetch correct env variables for GenericGenerateAccount function")
}
