package vault

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"time"

	global_environment "github.ibm.com/gftn/world-wire-services/utility/global-environment"

	"github.com/hashicorp/vault/api"
)

type Vault struct {
	ServiceName   string
	token         string
	vaultAddr     string
	client        *api.Client
	ParticipantId string
	ENV           string
}

// hard-codede vaule for now, To be changed
const envTokenName = "TOKEN"
const envAddrName = "VAULT_ADDR"
const serviceName = "send-service"
const participantId = "ibm01"
const envStage = "sandbox"

func InitializeVault() (*Vault, error) {
	LOGGER.Infof("Initializing vault connection")
	var token, vaultAddr string
	var exists bool

	if token, exists = os.LookupEnv(envTokenName); !exists {
		return nil, errors.New("No Vault API token detected")
	}

	if vaultAddr, exists = os.LookupEnv(envAddrName); !exists {
		return nil, errors.New("No Vault API address detected")
	}

	var vaultClient = &Vault{
		token:         token,
		vaultAddr:     vaultAddr,
		ServiceName:   serviceName,
		ParticipantId: participantId,
		ENV:           envStage,
	}
	var err error
	// TLS config
	tlsConfig := &tls.Config{}
	rootCAs := x509.NewCertPool()

	cacert := os.Getenv(global_environment.ENV_KEY_VAULT_CERT)
	sDec, err := base64.StdEncoding.DecodeString(cacert)
	if err != nil {
		return nil, errors.New("Certificate malformed: " + err.Error())
	}
	// Append our cert to the system pool
	if ok := rootCAs.AppendCertsFromPEM([]byte(sDec)); !ok {
		return nil, errors.New("Encounter error while appending Root CA cert")
	}

	// append cert as RootCAs
	tlsConfig.RootCAs = rootCAs

	transport := &http.Transport{TLSClientConfig: tlsConfig}

	//----------------
	/*
		tlsConfig := &tls.Config{InsecureSkipVerify: true}
		transport := &http.Transport{TLSClientConfig: tlsConfig}
	*/
	//--------
	var httpClient = &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	config := &api.Config{
		Address:    vaultClient.vaultAddr,
		HttpClient: httpClient,
	}
	vaultClient.client, err = api.NewClient(config)
	if err != nil {
		LOGGER.Error(err)
		return nil, err
	}
	vaultClient.client.SetToken(vaultClient.token)
	LOGGER.Infof("Vault connection established")

	return vaultClient, nil
}
