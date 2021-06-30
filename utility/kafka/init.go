package kafka

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.ibm.com/gftn/world-wire-services/utility/database"

	global_environment "github.ibm.com/gftn/world-wire-services/utility/global-environment"
	"github.ibm.com/gftn/world-wire-services/utility/payment/utils/parse"
	whitelist_handler "github.ibm.com/gftn/world-wire-services/utility/payment/utils/whitelist-handler"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.ibm.com/gftn/world-wire-services/utility/payment/constant"
	"github.ibm.com/gftn/world-wire-services/utility/payment/utils/signing"
	"github.ibm.com/gftn/world-wire-services/utility/payment/utils/transaction"
)

type KafkaOpreations struct {
	BrokerURL              string
	AutoOffReset           string
	SecurityProtocol       string
	SslCaLocation          string
	SslCertificateLocation string
	SslKeyLocation         string
	SslKeyPassword         string
	SaslUsername           string
	SaslPassword           string
	SaslMechanism          string
	Producer               *kafka.Producer
	Consumers              []*kafka.Consumer
	GroupId                string
	//used only by send-service
	FundHandler      transaction.CreateFundingOpereations
	SignHandler      signing.CreateSignOperations
	WhitelistHandler whitelist_handler.ParticipantWhiteList
	DbClient         *database.PostgreDatabaseClient
	ResponseHandler  *parse.ResponseHandler
}

func Initialize() (*KafkaOpreations, error) {

	var ops = &KafkaOpreations{}

	ops.BrokerURL = os.Getenv(ENV_KEY_KAFKA_BROKER_URL)
	if len(ops.BrokerURL) == 0 {
		LOGGER.Errorf("Kafka broker URL is empty")
		return &KafkaOpreations{}, errors.New("kafka broker url is empty")
	}

	authMode := strings.ToLower(os.Getenv(ENV_KEY_KAFKA_AUTH_MODE))
	if authMode == "" {
		LOGGER.Errorf("Kafka authentication mode is empty, please choose either sasl or ssl")
		return &KafkaOpreations{}, errors.New("Kafka authentication mode is empty, please choose either sasl or ssl")
	}
	// Check if the environment variable ENV_KEY_KAFKA_AUTH_MODE was set to ssl or sasl. If it's ssl, setting up certificate that
	// will be use by the Kafka producer and consumer.
	if authMode == KAFKA_SSL {
		ops.SecurityProtocol = KAFKA_SSL
		ops.SslCaLocation = os.Getenv(ENV_KEY_KAFKA_CA_LOCATION)
		ops.SslCertificateLocation = os.Getenv(ENV_KEY_KAFKA_CERTIFICATE_LOCATION)
		ops.SslKeyLocation = os.Getenv(ENV_KEY_KAFKA_KEY_LOCATION)
		pw, _ := ioutil.ReadFile(os.Getenv(ENV_KEY_KAFKA_KEY_PASSWORD))
		ops.SslKeyPassword = string(pw)
	} else if authMode == KAFKA_SASL {
		ops.SecurityProtocol = KAFKA_SASL
		ops.SaslUsername = os.Getenv(ENV_KEY_KAFKA_KEY_SASL_USER)
		ops.SaslPassword = os.Getenv(ENV_KEY_KAFKA_KEY_SASL_PASSWORD)
		ops.SaslMechanism = os.Getenv(ENV_KEY_KAFKA_KEY_SASL_MECHANISM)
	} else {
		ops.SecurityProtocol = "false"
	}

	var err error
	ops.ResponseHandler, err = parse.Initialize()
	if err != nil {
		return nil, err
	}
	err = ops.InitProducer()
	if err != nil {
		LOGGER.Errorf("Error initializing Kafka producer: %v", err.Error())
		return &KafkaOpreations{}, err
	}
	return ops, nil
}

func (ops *KafkaOpreations) InitProducer() error {

	p := &kafka.Producer{}
	var err error

	if ops.SecurityProtocol == KAFKA_SSL {
		p, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers":        ops.BrokerURL,
			"security.protocol":        ops.SecurityProtocol,
			"ssl.ca.location":          ops.SslCaLocation,
			"ssl.certificate.location": ops.SslCertificateLocation,
			"ssl.key.location":         ops.SslKeyLocation,
			"ssl.key.password":         ops.SslKeyPassword,
		})
	} else if ops.SecurityProtocol == KAFKA_SASL {
		p, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers": ops.BrokerURL,
			"sasl.username":     ops.SaslUsername,
			"sasl.password":     ops.SaslPassword,
			"sasl.mechanism":    ops.SaslMechanism,
			"security.protocol": ops.SecurityProtocol,
		})
	} else {
		p, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers": ops.BrokerURL,
		})
	}

	if err != nil {
		return err
	}

	ops.Producer = p
	return nil

}

func (ops *KafkaOpreations) InitPaymentConsumer(groupId string, router func(string, []byte, *KafkaOpreations)) error {

	listeningTopics := []string{REQUEST_TOPIC, RESPONSE_TOPIC}

	for consumerIndex, topicType := range listeningTopics {

		c := &kafka.Consumer{}
		var err error

		if groupId != "" {
			ops.GroupId = groupId
		} else {
			return errors.New("Group ID is empty")
		}
		LOGGER.Infof("Initiating Kafka consumer at %v, consumer index: %v", ops.GroupId, consumerIndex)
		if ops.SecurityProtocol == constant.KAFKA_SSL {
			c, err = kafka.NewConsumer(&kafka.ConfigMap{
				"bootstrap.servers":        ops.BrokerURL,
				"group.id":                 ops.GroupId,
				"auto.offset.reset":        "latest",
				"security.protocol":        ops.SecurityProtocol,
				"ssl.ca.location":          ops.SslCaLocation,
				"ssl.certificate.location": ops.SslCertificateLocation,
				"ssl.key.location":         ops.SslKeyLocation,
				"ssl.key.password":         ops.SslKeyPassword,
			})
		} else if ops.SecurityProtocol == KAFKA_SASL {
			c, err = kafka.NewConsumer(&kafka.ConfigMap{
				"bootstrap.servers": ops.BrokerURL,
				"group.id":          ops.GroupId,
				"auto.offset.reset": "latest",
				"sasl.username":     ops.SaslUsername,
				"sasl.password":     ops.SaslPassword,
				"sasl.mechanism":    ops.SaslMechanism,
				"security.protocol": ops.SecurityProtocol,
			})
		} else {
			c, err = kafka.NewConsumer(&kafka.ConfigMap{
				"bootstrap.servers": ops.BrokerURL,
				"group.id":          ops.GroupId,
				"auto.offset.reset": "latest",
			})
		}

		if err != nil {
			LOGGER.Errorf("Error creating the Kafka consumer: %s", err.Error())
			return err
		}

		ops.Consumers = append(ops.Consumers, c)

		homeDomainName := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
		kafkaReqTopic := homeDomainName + topicType
		go ops.consumerStartListening(kafkaReqTopic, topicType, consumerIndex, router)

	}

	// initialize admin client

	homeDomain := os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME)
	prServiceURL := os.Getenv(global_environment.ENV_KEY_PARTICIPANT_REGISTRY_URL)
	ops.FundHandler = transaction.InitiateFundingOperations(prServiceURL, homeDomain)

	// initailize sign handler
	ops.SignHandler = signing.InitiateSignOperations(prServiceURL)

	// initialize whitelist handler
	ops.WhitelistHandler = whitelist_handler.CreateWhiteListServiceOperations()

	ops.DbClient = &database.PostgreDatabaseClient{}
	err := ops.DbClient.CreateConnection()
	if err != nil {
		return err
	}

	return nil
}
