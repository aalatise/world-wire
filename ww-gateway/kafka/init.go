package kafka

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafka_utils "github.com/IBM/world-wire/utility/kafka"
)

func Initialize() (*kafka.Consumer, error) {

	var kafkaActor *kafka_utils.KafkaOpreations

	/*
		Retrieve kafka settings
	*/

	if len(os.Getenv(kafka_utils.ENV_KEY_KAFKA_BROKER_URL)) == 0 {
		LOGGER.Errorf("Kafka broker URL is empty")
		return &kafka.Consumer{}, errors.New("kafka broker url is empty")
	}

	kafkaActor = &kafka_utils.KafkaOpreations{}
	kafkaActor.BrokerURL = os.Getenv(kafka_utils.ENV_KEY_KAFKA_BROKER_URL)
	kafkaActor.GroupId = os.Getenv(kafka_utils.ENV_KEY_KAFKA_GROUP_ID)
	kafkaActor.AutoOffReset = os.Getenv(kafka_utils.ENV_KEY_KAFKA_AUTO_OFF_RESET)
	authMode := strings.ToLower(os.Getenv(kafka_utils.ENV_KEY_KAFKA_AUTH_MODE))
	// Check if the environment variable KAFKA_ENABLE_SSL was set to true. If it's true, setting up certificate that
	// will be use by the Kafka producer and consumer.
	if authMode == kafka_utils.KAFKA_SSL {
		kafkaActor.SecurityProtocol = kafka_utils.KAFKA_SSL
		kafkaActor.SslCaLocation = os.Getenv(kafka_utils.ENV_KEY_KAFKA_CA_LOCATION)
		kafkaActor.SslCertificateLocation = os.Getenv(kafka_utils.ENV_KEY_KAFKA_CERTIFICATE_LOCATION)
		kafkaActor.SslKeyLocation = os.Getenv(kafka_utils.ENV_KEY_KAFKA_KEY_LOCATION)
		pw, _ := ioutil.ReadFile(os.Getenv(kafka_utils.ENV_KEY_KAFKA_KEY_PASSWORD))
		kafkaActor.SslKeyPassword = string(pw)
	} else if authMode == kafka_utils.KAFKA_SASL {
		kafkaActor.SecurityProtocol = kafka_utils.KAFKA_SASL
		kafkaActor.SaslUsername = os.Getenv(kafka_utils.ENV_KEY_KAFKA_KEY_SASL_USER)
		kafkaActor.SaslPassword = os.Getenv(kafka_utils.ENV_KEY_KAFKA_KEY_SASL_PASSWORD)
		kafkaActor.SaslMechanism = os.Getenv(kafka_utils.ENV_KEY_KAFKA_KEY_SASL_MECHANISM)
	} else {
		kafkaActor.SecurityProtocol = "false"
	}

	/*
		Init consumer settings
	*/

	consumer, err := InitConsumer(
		kafkaActor.BrokerURL,
		kafkaActor.GroupId,
		kafkaActor.AutoOffReset,
		kafkaActor.SecurityProtocol,
		kafkaActor.SslCaLocation,
		kafkaActor.SslCertificateLocation,
		kafkaActor.SslKeyLocation,
		kafkaActor.SslKeyPassword,
		kafkaActor.SaslUsername,
		kafkaActor.SaslPassword,
		kafkaActor.SaslMechanism)
	if err != nil {
		LOGGER.Errorf("Error creating the Kafka consumer: %s", err.Error())
		return &kafka.Consumer{}, err
	}
	return consumer, nil
}

func InitConsumer(brokerURL, groupId, autoOffReset, securityProtocol, caLocation, certLocation, keyLocation, keyPassword, saslUsername, saslPassword, saslMechanism string) (*kafka.Consumer, error) {

	c := &kafka.Consumer{}
	var err error
	if securityProtocol == kafka_utils.KAFKA_SSL {
		c, err = kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":         brokerURL,
			"group.id":                  groupId,
			"auto.offset.reset":         autoOffReset,
			"security.protocol":         securityProtocol,
			"ssl.ca.location":           caLocation,
			"ssl.certificate.location":  certLocation,
			"ssl.key.location":          keyLocation,
			"ssl.key.password":          keyPassword,
			"session.timeout.ms":        60000,
			"max.partition.fetch.bytes": 3000000,
			"enable.auto.commit":        false,
			"enable.auto.offset.store":  false,
			"max.poll.interval.ms":      86400000,
		})
	} else if securityProtocol == kafka_utils.KAFKA_SASL {
		c, err = kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":         brokerURL,
			"group.id":                  groupId,
			"auto.offset.reset":         autoOffReset,
			"sasl.username":             saslUsername,
			"sasl.password":             saslPassword,
			"sasl.mechanisms":           saslMechanism,
			"security.protocol":         securityProtocol,
			"session.timeout.ms":        60000,
			"max.partition.fetch.bytes": 3000000,
			"enable.auto.commit":        false,
			"enable.auto.offset.store":  false,
			"max.poll.interval.ms":      86400000,
		})
	} else {
		c, err = kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":         brokerURL,
			"group.id":                  groupId,
			"auto.offset.reset":         autoOffReset,
			"session.timeout.ms":        60000,
			"max.partition.fetch.bytes": 3000000,
			"enable.auto.commit":        false,
			"enable.auto.offset.store":  false,
			"max.poll.interval.ms":      86400000,
		})
	}

	return c, err
}
