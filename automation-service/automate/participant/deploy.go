package participant

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.ibm.com/gftn/world-wire-services/automation-service/environment"

	"github.ibm.com/gftn/world-wire-services/automation-service/constant"

	"github.com/go-openapi/strfmt"
	"github.com/op/go-logging"
	"github.ibm.com/gftn/world-wire-services/automation-service/internal_model"
	"github.ibm.com/gftn/world-wire-services/automation-service/model/model"
	"github.ibm.com/gftn/world-wire-services/automation-service/utility"
	gftn_model "github.ibm.com/gftn/world-wire-services/gftn-models/model"
	"github.ibm.com/gftn/world-wire-services/utility/database"
	global_environment "github.ibm.com/gftn/world-wire-services/utility/global-environment"
	"github.ibm.com/gftn/world-wire-services/utility/participant"
	"github.ibm.com/gftn/world-wire-services/utility/response"
)

var LOGGER = logging.MustGetLogger("participant")

type DeploymentOperations struct {
	dockerImageVersion string
	vaultToken         string
	vaultAddress       string
	ibmIKSId           string
	ibmESName          string
	// Export this because we need to re-use throughout the packages
	Session      *database.MongoClient
	VaultSession *VaultSession
}

func InitiateDeploymentOperations() (DeploymentOperations, error) {
	op := DeploymentOperations{}

	op.initVaultSession()
	// Pull Vault secrets for automation service to function
	op.retrieveVaultSecrets()

	op.dockerImageVersion = os.Getenv(environment.ENV_DOCKER_IMAGE_VERSION)
	op.vaultAddress = os.Getenv("VAULT_ADDR")
	op.vaultToken = os.Getenv("VAULT_TOKEN")
	op.ibmIKSId = os.Getenv("IBMIKSID")
	op.ibmESName = os.Getenv("IBMCLOUD_ES_NAME")

	LOGGER.Debugf("version: %s, vault addr: %s, iks id: %s, es name: %s", op.dockerImageVersion, op.vaultAddress, op.ibmIKSId, op.ibmESName)

	// Connect to MongoDB and save client context
	LOGGER.Infof("\t* Automation Service connecting Mongo DB ")
	client, err := database.InitializeIbmCloudConnection()
	if err != nil {
		LOGGER.Errorf("IBM Cloud Mongo DB connection failed! %s", err)
		panic("IBM Cloud Mongo DB connection failed! " + err.Error())
	}
	op.Session = client

	// Login to IBM Cloud with API key
	LOGGER.Infof("\t* Login to IBM CLoud")
	loginScript := constant.K8sBasePath + "/script/ibm_cloud_login.sh"
	// IBMAPIKey := os.Getenv("IBMCLOUD_API_KEY")
	IBMAPIKey, exists := os.LookupEnv("IBMCLOUD_API_KEY")
	if !exists {
		return op, errors.New("failed to login to the IBM Cloud: IBMCLOUD_API_KEY env var Does not exist")
	}
	err = utility.RunBashCmd(loginScript, IBMAPIKey, op.ibmIKSId)
	if err != nil {
		LOGGER.Errorf("Failed to login to the IBM Cloud: %s", err.Error())
		return op, errors.New("failed to login to the IBM Cloud: " + err.Error())
	}

	return op, nil
}

func (op *DeploymentOperations) DeployParticipantServicesAndConfigs(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("X-XSS-Protection", "1")
	var input model.Automation
	err := json.NewDecoder(req.Body).Decode(&input)
	if err != nil {
		LOGGER.Errorf("Error decoding the payload: %s", err.Error())
		response.NotifyWWError(w, req, http.StatusBadRequest, "DEPLOYMENT-1001", err)
		return
	}

	err = input.Validate(strfmt.Default)
	if err != nil {
		LOGGER.Errorf("Error while validating Payload: %s", err.Error())
		response.NotifyWWError(w, req, http.StatusBadRequest, "DEPLOYMENT-1002", err)
		return
	}

	institutionID := *input.InstitutionID
	participantID := *input.ParticipantID
	imageVersion := op.dockerImageVersion
	participantRole := *input.Role

	LOGGER.Infof("Participant: %s, Docker Image Tag: %s", participantID, imageVersion)

	// Check if the participant was recorded in the firebase
	// if not, Write the participant information into MongoDB: configuring
	// if yes, check which process was failed and redo it again
	statuses, err := op.checkParticipantRecord(input)
	if err != nil {
		response.NotifyWWError(w, req, http.StatusBadRequest, "DEPLOYMENT-1100", err)
	}

	if statuses == nil || len(statuses) == 0 {
		LOGGER.Infof("Create participant entry in the participant-registry: %s", participantID)
		createEntryErr := createParticipantEntry(input)
		if createEntryErr != nil {
			op.updateParticipantStatus(institutionID, participantID, constant.StatusCreatePREntryFailed, "failed", nil, true)
			response.NotifyWWError(w, req, http.StatusBadRequest, "DEPLOYMENT-1100", createEntryErr)
			return
		}

		LOGGER.Info("Deploy participant backend micro services")
		deployErr := op.deployNodesProcess(participantRole, imageVersion, input)
		if deployErr != nil {
			response.NotifyWWError(w, req, http.StatusBadRequest, "DEPLOYMENT-1100", deployErr)
			return
		}

		// If the role of the participant is a market maker, create the issuing and operating account for it
		if participantRole == constant.MarketMakerParticipant {
			LOGGER.Info("Create participant accounts")
			go op.createAccountsProcess(institutionID, participantID)
		} else {
			go op.updateParticipantStatus(institutionID, participantID, constant.StatusComplete, "done", nil, true)
		}

		response.NotifySuccess(w, req, "Successfully deployed")
		return
	} else if statuses[0] == constant.StatusComplete {
		response.NotifySuccess(w, req, "Successfully deployed")
		return
	} else {
		for _, s := range statuses {
			LOGGER.Debugf("Doing retry, status is %s", s)
			switch s {
			case constant.StatusCreatePREntryFailed:
				LOGGER.Debug("**Retry**: Create Participant entry")
				err := createParticipantEntry(input)
				if err != nil {
					op.updateParticipantStatus(institutionID, participantID, constant.StatusCreatePREntryFailed, "failed", nil, true)
					response.NotifyWWError(w, req, http.StatusBadRequest, "DEPLOYMENT-1100", err)
					return
				}
				// else update that pr entry error was resolved
				go op.updateParticipantStatus(institutionID, participantID, constant.StatusCreatePREntryFailed, "resolve", nil, false)

				LOGGER.Debug("**Retry**: Deploy participant backend micro services")
				deployErr := op.deployNodesProcess(participantRole, imageVersion, input)
				if deployErr != nil {
					response.NotifyWWError(w, req, http.StatusBadRequest, "DEPLOYMENT-1100", deployErr)
					return
				}

				if participantRole == constant.MarketMakerParticipant {
					LOGGER.Debug("**Retry**: Create participant accounts")
					go op.createAccountsProcess(institutionID, participantID)
				} else {
					go op.updateParticipantStatus(institutionID, participantID, constant.StatusComplete, "done", nil, false)
				}

			case constant.StatusCreateKafkaTopicFailed:
				createKafkaTopicsChan := make(chan error)
				LOGGER.Debug("**Retry**: Create Kafka certificate and topics")
				go op.createKafkaTopics(createKafkaTopicsChan, participantRole, institutionID, participantID)
				createTopicError := <-createKafkaTopicsChan
				if createTopicError != nil {
					op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateKafkaTopicFailed, "failed", nil, true)
					response.NotifyWWError(w, req, http.StatusBadRequest, "DEPLOYMENT-1100", err)
					return
				}
				// Else Update that the re-creation has succeeded
				go op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateKafkaTopicFailed, "resolve", nil, false)
			case constant.StatusCreateVaultSecretsFailed:
				createSecretChan := make(chan error)
				LOGGER.Debug("**Retry**: Create AWS secret")
				go op.createVaultSecrets(createSecretChan, participantRole, institutionID, participantID)
				createSecretError := <-createSecretChan
				if createSecretError != nil {
					op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateVaultSecretsFailed, "failed", nil, true)
					response.NotifyWWError(w, req, http.StatusBadRequest, "DEPLOYMENT-1100", createSecretError)
					return
				}
				go op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateVaultSecretsFailed, "resolve", nil, false)
			case constant.StatusCreateMicroServicesFailed:
				LOGGER.Debug("**Retry**: Deploy participant backend micro services")
				deployParticipantServiceError := op.deployParticipantServices(participantRole, participantID, imageVersion, input.Replica)
				if deployParticipantServiceError != nil {
					op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateMicroServicesFailed, "failed", nil, true)
					response.NotifyWWError(w, req, http.StatusBadRequest, "DEPLOYMENT-1100", deployParticipantServiceError)
					return
				}
				go op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateMicroServicesFailed, "resolve", nil, false)

				if participantRole == constant.MarketMakerParticipant {
					LOGGER.Debug("**Retry**: Create participant accounts")
					go op.createAccountsProcess(institutionID, participantID)
				} else {
					go op.updateParticipantStatus(institutionID, participantID, constant.StatusComplete, "done", nil, false)
				}
			case constant.StatusCreateIssuingAccountFailed:
				LOGGER.Debug("**Retry**: Create issuing account")
				apiSvcClient, getErr := getAPISVCClient(participantID)
				if getErr != nil {
					op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateIssuingAccountFailed, "failed", nil, true)
					response.NotifyWWError(w, req, http.StatusBadRequest, "DEPLOYMENT-1100", getErr)
					return
				}

				createIssuingAccountErr := createIssuingAccount(apiSvcClient)
				if createIssuingAccountErr != nil {
					op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateIssuingAccountFailed, "failed", nil, true)
					response.NotifyWWError(w, req, http.StatusBadRequest, "DEPLOYMENT-1100", createIssuingAccountErr)
					return
				}

				go op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateIssuingAccountFailed, "resolve", nil, false)

				createOperatingAccountErr := createOperatingAccount(apiSvcClient)
				if createOperatingAccountErr != nil {
					op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateOperatingAccountFailed, "failed", nil, true)
					response.NotifyWWError(w, req, http.StatusBadRequest, "DEPLOYMENT-1100", createOperatingAccountErr)
					return
				} else {
					go op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateOperatingAccountFailed, "done", nil, false)
				}
			case constant.StatusCreateOperatingAccountFailed:
				LOGGER.Debug("**Retry**: Create operating account")
				apiSvcClient, getErr := getAPISVCClient(participantID)
				if getErr != nil {
					op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateOperatingAccountFailed, "failed", nil, true)
					response.NotifyWWError(w, req, http.StatusBadRequest, "DEPLOYMENT-1100", getErr)
					return
				}

				createOperatingAccountErr := createOperatingAccount(apiSvcClient)
				if createOperatingAccountErr != nil {
					op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateOperatingAccountFailed, "failed", nil, true)
					response.NotifyWWError(w, req, http.StatusBadRequest, "DEPLOYMENT-1100", createOperatingAccountErr)
					return
				} else {
					go op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateOperatingAccountFailed, "done", nil, false)
				}
			default:
				LOGGER.Warningf("Unknown status: %s", s)
			}
		}
		response.NotifySuccess(w, req, "Successfully deployed")
		return
	}
}

func (op *DeploymentOperations) deployNodesProcess(role, imageVersion string, input model.Automation) error {
	//institutionID := *input.InstitutionID
	participantID := *input.ParticipantID
	institutionID := *input.InstitutionID
	replica := input.Replica

	// createIAMPolicyChan := make(chan error)
	// createSecretChan := make(chan error)

	LOGGER.Debugf("Deploying participant: %s", role)

	// Create Kafka topics
	createKafkaTopicsChan := make(chan error)
	go op.createKafkaTopics(createKafkaTopicsChan, role, institutionID, participantID)

	// Create secret in Vault
	createVaultSecretsChan := make(chan error)
	go op.createVaultSecrets(createVaultSecretsChan, role, institutionID, participantID)

	LOGGER.Debug("---------- Waiting for all the pre-process to be finished ----------")

	createKafkaTopicErr := <-createKafkaTopicsChan
	createVaultSecretsErr := <-createVaultSecretsChan

	if createVaultSecretsErr != nil || createKafkaTopicErr != nil {
		LOGGER.Errorf("Failed to create Vault secret: %+v and Kafka topics: %+v", createVaultSecretsErr, createKafkaTopicErr)
		op.updateParticipantStatus(*input.InstitutionID, *input.ParticipantID, constant.StatusCreateMicroServicesFailed, "failed", nil, true)
		return errors.New("failed to create Vault secret or Kafka topics")
	}

	LOGGER.Debug("---------- All the pre-process were successfully finished ----------")

	deployParticipantServiceError := op.deployParticipantServices(role, participantID, imageVersion, replica)
	if deployParticipantServiceError != nil {
		// Update the status in FireBase: create_micro_services_failed
		op.updateParticipantStatus(*input.InstitutionID, *input.ParticipantID, constant.StatusCreateMicroServicesFailed, "failed", nil, true)
		return deployParticipantServiceError
	}

	LOGGER.Info("===> Successfully deployed participant services")

	return nil
}

func (op DeploymentOperations) createAccountsProcess(institutionID, participantID string) {
	apiSvcClient, getErr := getAPISVCClient(participantID)
	if getErr != nil {
		LOGGER.Errorf("Failed to create api-service http client: %s", getErr.Error())
		go op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateIssuingAccountFailed, "failed", nil, true)
		return
	}

	// Wait for the api-service to start up
	LOGGER.Debug("Create issuing account")
	time.Sleep(time.Second * 60)

	createIssuingAccountErr := createIssuingAccount(apiSvcClient)
	if createIssuingAccountErr != nil {
		LOGGER.Errorf("Failed to create issuing account: %s", createIssuingAccountErr.Error())
		// Update the status in FireBase: create_issuing_account_failed
		go op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateIssuingAccountFailed, "failed", nil, true)
		return
	}

	// Wait for the api-service to start up
	LOGGER.Debug("Create operating account")
	time.Sleep(time.Second * 10)

	createOperatingAccountErr := createOperatingAccount(apiSvcClient)
	if createOperatingAccountErr != nil {
		LOGGER.Errorf("Failed to create operating account: %s", createOperatingAccountErr.Error())
		// Update the status in FireBase: create_operating_account_failed
		go op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateOperatingAccountFailed, "failed", nil, true)
		return
	}

	LOGGER.Info("===> Successfully create participant accounts")
	go op.updateParticipantStatus(institutionID, participantID, constant.StatusComplete, "done", nil, false)
	return
}

func (op DeploymentOperations) createKafkaTopics(ch chan error, role, institutionID, participantID string) {
	switch role {
	case constant.MarketMakerParticipant:
		// Create Kafka certificate for each participant
		LOGGER.Infof("Create Kafka topics for participant: %s", participantID)
		createPTopicScriptPath := constant.IbmESPath + "/create_cert_and_topic.sh"
		err := utility.RunBashCmd(createPTopicScriptPath, participantID, op.ibmESName)
		if err != nil {
			LOGGER.Errorf("Failed to deploy participant Kafka topics")
			// Update the status in FireBase: create_kafka_topic_failed
			go op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateKafkaTopicFailed, "failed", nil, false)
			ch <- err
			return
		}
		LOGGER.Debugf("Participant Kafka topics successfully generated")
	case constant.AssetIssuerParticipant:
		// Create Kafka certificate for anchor
		LOGGER.Infof("Create Kafka cert and topics for anchor: %s", participantID)
		createATopicScriptPath := constant.IbmESPath + "/create_is_cert_and_topic.sh"
		err := utility.RunBashCmd(createATopicScriptPath, participantID, op.ibmESName)
		if err != nil {
			LOGGER.Errorf("Failed to deploy anchor Kafka certificate and topics")
			// Update the status in FireBase: create_kafka_topic_failed
			go op.updateParticipantStatus(institutionID, participantID, constant.StatusCreateKafkaTopicFailed, "failed", nil, false)
			ch <- err
			return
		}

		LOGGER.Debugf("Anchor Kafka topics successfully generated")
	}
	ch <- nil
	return
}

func createParticipantEntry(input model.Automation) error {
	// Construct a participant record
	prServiceURL := os.Getenv(global_environment.ENV_KEY_PARTICIPANT_REGISTRY_URL)
	participantModel := &gftn_model.Participant{
		Bic:         input.Bic,
		CountryCode: input.CountryCode,
		ID:          input.ParticipantID,
		Role:        input.Role,
	}

	participantPayload, _ := json.Marshal(participantModel)

	prClient := &internal_model.Client{
		HTTPClient: &http.Client{Timeout: time.Second * 30},
		URL:        prServiceURL,
	}

	// call the /internal/pr/domain/input.ParticipantID "GET" endpoint to check if the participant already exist in the participant registry
	rGet, err := http.NewRequest(http.MethodGet, prClient.URL+"/internal/pr/domain/"+*input.ParticipantID, bytes.NewBuffer(participantPayload))
	if err != nil {
		LOGGER.Errorf("Error creating HTTP request: %s", err.Error())
		return err
	}

	resGet, restGetErr := http.DefaultClient.Do(rGet)
	if restGetErr != nil {
		LOGGER.Errorf("Unable to get participant from pr-serivce: %s", restGetErr.Error())
		return restGetErr
	} else if resGet.StatusCode != http.StatusOK {
		if resGet.StatusCode == http.StatusBadRequest {
			for i := 1; i < 6; i++ {
				LOGGER.Warningf("Service not ready, retry it %d", i)
				time.Sleep(10 * time.Second)
				rGet, err := http.NewRequest(http.MethodGet, prClient.URL+"/internal/pr/domain/"+*input.ParticipantID, bytes.NewBuffer(participantPayload))
				if err != nil {
					LOGGER.Errorf("Error creating HTTP request: %s", err.Error())
					return err
				}

				resGet, restErr := http.DefaultClient.Do(rGet)
				if restErr != nil {
					return restErr
				} else if resGet.StatusCode != http.StatusOK {
					if resGet.StatusCode == http.StatusBadRequest {
						if i == 5 {
							return errors.New("unable to get participant from pr-serivce")
						} else {
							continue
						}
					} else if resGet.StatusCode == http.StatusNotFound {
						LOGGER.Warning("Participant did not exist")
						break
					}
				} else {
					LOGGER.Error("Participant already exist")
					return errors.New("participant already exist")
				}
			}
		} else if resGet.StatusCode == http.StatusNotFound {
			LOGGER.Warning("Participant did not exist")
		}
	} else {
		LOGGER.Error("Participant already exist")
		return errors.New("participant already exist")
	}

	LOGGER.Debug("Create participant entry")
	// call the /internal/pr "POST" endpoint to create a participant entry into participant registry
	r, reqErr := http.NewRequest(http.MethodPost, prClient.URL+"/internal/pr", bytes.NewBuffer(participantPayload))
	if reqErr != nil {
		LOGGER.Errorf("Error creating HTTP request: %s", reqErr.Error())
		return reqErr
	}

	res, restErr := http.DefaultClient.Do(r)
	if restErr != nil {
		LOGGER.Errorf("Unable to create participant into pr-serivce: %s", restErr.Error())
		return restErr
	} else if res.StatusCode != http.StatusOK {
		responseBody, responseErr := ioutil.ReadAll(resGet.Body)
		if responseErr != nil {
			return errors.New("error trying to read the response from pr-service: " + responseErr.Error())
		}
		LOGGER.Errorf("%s", string(responseBody))
		LOGGER.Errorf("Unable to create participant into pr-serivce: %d", res.StatusCode)
		return errors.New("unable to create participant entry, http response status code from pr-service endpoint is not 200")
	}

	// Activate the participant using pr-service endpoint
	participantStatus := "active"
	participantStatusModel := &gftn_model.ParticipantStatus{
		Status: &participantStatus,
	}

	participantStatusPayload, _ := json.Marshal(participantStatusModel)

	LOGGER.Debug("Activate participant")
	// call the internal/pr/{participantID}/status endpoint to update the status to `active`
	statusReq, reqErr := http.NewRequest(http.MethodPut, prClient.URL+"/internal/pr/"+*input.ParticipantID+"/status", bytes.NewBuffer(participantStatusPayload))
	if reqErr != nil {
		LOGGER.Errorf("Error creating HTTP request: %s", reqErr.Error())
		return reqErr
	}

	statusRes, restErr := http.DefaultClient.Do(statusReq)
	if restErr != nil {
		LOGGER.Errorf("Unable to update the status: %s", restErr.Error())
		return restErr
	} else if statusRes.StatusCode != http.StatusOK {
		responseBody, responseErr := ioutil.ReadAll(resGet.Body)
		if responseErr != nil {
			return errors.New("error trying to read the response from pr-service: " + responseErr.Error())
		}
		LOGGER.Errorf("%s", string(responseBody))
		LOGGER.Errorf("Unable to update the status: %d", statusRes.StatusCode)
		return errors.New("unable to update the participant status, http response status code from pr-service endpoint is not 200")
	}

	LOGGER.Info("===> Successfully create participant entry in participant registry")

	return nil
}

func (op DeploymentOperations) deployParticipantServices(role, participantID, imageVersion, replica string) error {
	LOGGER.Infof("Create Role in the Vault")
	script := constant.K8sBasePath + "/script/setup_vault_token.sh"
	vErr := utility.RunBashCmd(script, participantID, op.vaultToken, op.vaultAddress, op.ibmIKSId)
	if vErr != nil {
		LOGGER.Errorf("Failed to setup role in the vault")
		return errors.New("failed to setup role in the vaults")
	}

	// Create whitelist entry in GFTN/whitelist collection
	err := op.createPartsicipantWhitelistEntry(participantID)
	if err != nil {
		return errors.New("Failed to create whitelist entry in MongoDB for participant of ID: " + participantID)
	}

	switch role {
	case constant.MarketMakerParticipant:
		// Deploy participant services
		LOGGER.Infof("Deploy participant micro services on Kubernetes cluster")
		pScript := constant.K8sBasePath + "/script/create_participant_msvc.sh"
		pErr := utility.RunBashCmd(pScript, participantID, imageVersion, replica, op.ibmIKSId)
		if pErr != nil {
			LOGGER.Errorf("Failed to deploy participant services")
			return errors.New("failed to deploy participant services")
		}

		LOGGER.Debugf("Participant services successfully generated")
	case constant.AssetIssuerParticipant:
		// Deploy anchor's local services
		LOGGER.Infof("Deploy anchor local micro services on Kubernetes cluster")
		pScript := constant.K8sBasePath + "/script/create_anchor_msvc.sh"
		pErr := utility.RunBashCmd(pScript, participantID, imageVersion, replica, op.ibmIKSId)
		if pErr != nil {
			LOGGER.Errorf("Failed to deploy anchor local micro services")
			return errors.New("failed to deploy anchor local micro services")
		}

		LOGGER.Debugf("Anchor local micro services successfully generated")
	}

	return nil
}

func getAPISVCClient(participantID string) (*internal_model.Client, error) {
	// Call the api-service internal endpoint
	url := os.Getenv(global_environment.ENV_KEY_API_SVC_URL)
	apiServiceURL, convertErr := participant.GetServiceUrl(url, participantID)
	if convertErr != nil {
		LOGGER.Error(convertErr.Error())
		return nil, convertErr
	}

	apiSvcClient := &internal_model.Client{
		HTTPClient: &http.Client{Timeout: time.Second * 80},
		URL:        apiServiceURL,
	}

	return apiSvcClient, nil
}

func createIssuingAccount(apiSvcClient *internal_model.Client) error {
	// call the /internal/accounts/issuing endpoint to create issuing account
	issuingR, err := http.NewRequest(http.MethodPost, apiSvcClient.URL+"/internal/accounts/issuing", nil)
	if err != nil {
		LOGGER.Errorf("Error creating HTTP request: %s", err.Error())
		return err
	}

	resIssuing, restErr := http.DefaultClient.Do(issuingR)
	if restErr != nil || resIssuing.StatusCode != http.StatusOK {
		LOGGER.Errorf("Unable to create participant issuing account")
		for i := 1; i < 6; i++ {
			LOGGER.Warningf("Service not ready, retry it %d", i)
			time.Sleep(15 * time.Second)
			issuingR, err := http.NewRequest(http.MethodPost, apiSvcClient.URL+"/internal/accounts/issuing", nil)
			if err != nil {
				LOGGER.Errorf("Error creating HTTP request: %s", err.Error())
				return err
			}

			resIssuing, restErr := http.DefaultClient.Do(issuingR)
			if restErr != nil || resIssuing.StatusCode != http.StatusOK {
				LOGGER.Errorf("Unable to create participant issuing account")
				if i == 5 {
					LOGGER.Errorf("Unable to create participant issuing account")
					return errors.New("unable to create participant issuing account")
				} else {
					continue
				}
			} else {
				break
			}
		}
	}

	return nil
}

func createOperatingAccount(apiSvcClient *internal_model.Client) error {
	// call the /internal/accounts/default endpoint to create issuing account
	defaultR, err := http.NewRequest(http.MethodPost, apiSvcClient.URL+"/internal/accounts/default", nil)
	if err != nil {
		LOGGER.Errorf("Error creating HTTP request: %s", err.Error())
		return err
	}

	resDefault, restErr := http.DefaultClient.Do(defaultR)
	if restErr != nil || resDefault.StatusCode != http.StatusOK {
		LOGGER.Errorf("Unable to create participant operating account")
		for i := 1; i < 6; i++ {
			LOGGER.Warningf("Service not ready, retry it %d", i)
			time.Sleep(15 * time.Second)
			defaultR, err := http.NewRequest(http.MethodPost, apiSvcClient.URL+"/internal/accounts/default", nil)
			if err != nil {
				LOGGER.Errorf("Error creating HTTP request: %s", err.Error())
				return err
			}

			resDefault, restErr := http.DefaultClient.Do(defaultR)
			if restErr != nil || resDefault.StatusCode != http.StatusOK {
				LOGGER.Errorf("Unable to create participant operating account")
				if i == 5 {
					LOGGER.Errorf("Unable to create participant operating account")
					return errors.New("unable to create participant operating account")
				} else {
					continue
				}
			} else {
				break
			}
		}
	}

	return nil
}
