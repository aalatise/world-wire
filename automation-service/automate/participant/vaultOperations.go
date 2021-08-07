package participant

// Token will be read from VAULT_TOKEN env var

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/IBM/world-wire/automation-service/constant"
)

// VaultSession : Reusable Vault connection
type VaultSession struct {
	Client  *vault.Client
	Logical *vault.Logical
}

func (op *DeploymentOperations) initVaultSession() {
	tlsConfig := &tls.Config{}
	rootCAs := x509.NewCertPool()

	cacert := os.Getenv("VAULT_CERT")
	sDec, err := base64.StdEncoding.DecodeString(cacert)
	if err != nil {
		LOGGER.Panicf("Certificate malformed\nError: %v", err)
	}
	// Append our cert to the system pool
	if ok := rootCAs.AppendCertsFromPEM([]byte(sDec)); !ok {
		LOGGER.Panicf("Encounter error while appending Root CA cert\nError: %v", err)
	}

	// append cert as RootCAs
	tlsConfig.RootCAs = rootCAs

	transport := &http.Transport{TLSClientConfig: tlsConfig}

	var httpClient = &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	config := &vault.Config{}

	if vaultAddress, exists := os.LookupEnv("VAULT_ADDR"); exists {
		config.Address = vaultAddress
	} else {
		config.Address = constant.VaultAddress
	}
	config.HttpClient = httpClient

	client, err := vault.NewClient(config)
	if err != nil {
		LOGGER.Panicf("Failed to create Vault Client\nError: %v", err)
	}
	tokenBytes, err := ioutil.ReadFile(constant.VaultTokenPodPath)
	if err != nil {
		LOGGER.Panicf("Failed to get vault token file\nError: %v", err)
	}
	token := string(tokenBytes)
	// See https://github.com/hashicorp/vault/issues/7288
	token = strings.TrimSuffix(token, "\n")

	// Set token to auth requests
	client.SetToken(token)
	logical := client.Logical()

	op.VaultSession = &VaultSession{}

	op.VaultSession.Client = client
	op.VaultSession.Logical = logical
}

func (op *DeploymentOperations) createVaultSecrets(ch chan error, role, iid, pid string) {
	var err error
	var sendVaultData map[string]interface{}

	// Grab participant's secret template
	logical := op.VaultSession.Logical
	data, err := logical.Read("ww/data/sandbox/template/participant-template")
	if err != nil {
		LOGGER.Errorf("Failed to get participant secret template\nError: %v", err)
		go op.updateParticipantStatus(iid, pid, constant.StatusCreateVaultSecretsFailed, "failed", nil, false)
		ch <- err
		return
	}
	if data == nil {
		LOGGER.Errorf("Failed to find any participant secret template\n")
		go op.updateParticipantStatus(iid, pid, constant.StatusCreateVaultSecretsFailed, "failed", nil, false)
		ch <- errors.New("Failed to find any participant secret template")
		return
	}
	vaultData := data.Data["data"].(map[string]interface{})

	// If asset issuer then only create ww-gateway secret
	if role == constant.AssetIssuerParticipant && err == nil {
		sendVaultData = make(map[string]interface{})
		sendVaultData["data"] = vaultData["ww-gateway"]
		bytes, err := json.Marshal(sendVaultData)
		if err != nil {
			LOGGER.Errorf("Failed to create %s's secret for %s\nError: %v", "ww-gateway", pid, err)
			go op.updateParticipantStatus(iid, pid, constant.StatusCreateVaultSecretsFailed, "failed", nil, false)
			ch <- err
			return
		}
		logical.WriteBytes(constant.VaultBasePath+pid+"/ww-gateway/initialize", bytes)

		sendVaultData = make(map[string]interface{})
		sendVaultData["data"] = vaultData["participant"]
		bytes, err = json.Marshal(sendVaultData)
		if err != nil {
			LOGGER.Errorf("Failed to create %s's secret for %s\nError: %v", "participant", pid, err)
			go op.updateParticipantStatus(iid, pid, constant.StatusCreateVaultSecretsFailed, "failed", nil, false)
			ch <- err
			return
		}
		logical.WriteBytes(constant.VaultBasePath+pid+"/participant/initialize", bytes)
	} else if err == nil {

		// Response contains json object with keys as the service names
		// Iterate through each service
		for svcName, v := range vaultData {
			sendVaultData = make(map[string]interface{})
			sendVaultData["data"] = v
			bytes, err := json.Marshal(sendVaultData)
			if err != nil {
				LOGGER.Errorf("Failed to create %s's secret for %s\nError: %v", svcName, pid, err)
				go op.updateParticipantStatus(iid, pid, constant.StatusCreateVaultSecretsFailed, "failed", nil, false)
				ch <- err
				return
			}
			logical.WriteBytes(constant.VaultBasePath+pid+"/"+svcName+"/initialize", bytes)
		}
	}

	if err != nil {
		// Update the status in FireBase: create_aws_secret_failed
		LOGGER.Errorf("Failed to create %s's secret for %s\nError: %v", "participant", pid, err)
		go op.updateParticipantStatus(iid, pid, constant.StatusCreateVaultSecretsFailed, "failed", nil, false)
		ch <- err
		return
	}
	LOGGER.Infof("Participant services secret successfully generated")
	ch <- nil
}

func (op DeploymentOperations) retrieveVaultSecrets() error {
	var err error

	// Grab participant's secret template
	logical := op.VaultSession.Logical
	data, err := logical.Read("ww/data/automation-service")
	if err != nil || data == nil {
		LOGGER.Errorf("Failed to get automation-service secrets\nError: %v", err)
		return err
	}
	vaultData := data.Data["data"].(map[string]interface{})
	bytes, err := json.Marshal(vaultData)
	if err != nil {
		LOGGER.Errorf("Failed to get marshal retrieved vault data\nError: %v", err)
		return err
	}
	var envVars map[string]string
	err = json.Unmarshal(bytes, &envVars)
	if err != nil {
		LOGGER.Errorf("Failed to get unmarshal retrieved vault data\nError: %v", err)
		return err
	}
	for k, v := range envVars {
		os.Setenv(k, v)
	}
	LOGGER.Infof("Environment variables retrieved from vault and set")
	return err
}
