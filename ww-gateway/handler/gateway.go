package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.ibm.com/gftn/world-wire-services/utility/common"
	"github.ibm.com/gftn/world-wire-services/ww-gateway/environment"

	go_kafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.ibm.com/gftn/world-wire-services/gftn-models/model"
	global_environment "github.ibm.com/gftn/world-wire-services/utility/global-environment"
	kafka_utils "github.ibm.com/gftn/world-wire-services/utility/kafka"
	"github.ibm.com/gftn/world-wire-services/utility/response"
	"github.ibm.com/gftn/world-wire-services/ww-gateway/kafka"
	kafkaHandler "github.ibm.com/gftn/world-wire-services/ww-gateway/kafka"
	"github.ibm.com/gftn/world-wire-services/ww-gateway/utility"
)

type GatewayOperations struct {
	Consumer          map[string]*go_kafka.Consumer
	homeDomain        string
	StartingPartition map[string]*int32
}

func InitGatewayOperation() (GatewayOperations, error) {

	operation := GatewayOperations{
		homeDomain:        os.Getenv(global_environment.ENV_KEY_HOME_DOMAIN_NAME),
		StartingPartition: make(map[string]*int32, len(kafka_utils.SUPPORT_MESSAGE_TYPES)),
	}

	LOGGER.Infof("* Initiate Kafka consumer")

	/*
		Init consumer settings & starting partition index
	*/
	operation.Consumer = make(map[string]*go_kafka.Consumer, len(kafka_utils.SUPPORT_MESSAGE_TYPES))

	for _, msg := range kafka_utils.SUPPORT_MESSAGE_TYPES {
		err := operation.InitializeConsumer(msg)
		if err != nil {
			LOGGER.Errorf("Error creating the Kafka consumer: %s", err.Error())
			return GatewayOperations{}, err
		}
		LOGGER.Infof("* Kafka consumer successfully subscribed to topic %v", msg)
		var initPartition = int32(0)
		operation.StartingPartition[msg] = &initPartition
	}

	LOGGER.Infof("* InitGatewayOperations finished")
	return operation, nil
}

func (operation GatewayOperations) ServiceCheck(w http.ResponseWriter, req *http.Request) {
	LOGGER.Infof("Performing service check")
	response.Respond(w, http.StatusOK, []byte(`{"status":"Alive"}`))
	return
}

func (operation GatewayOperations) GetMessage(w http.ResponseWriter, request *http.Request) {
	LOGGER.Infof("gateway-service:Gateway Operations :get message")

	/*
		parsing request parameters
	*/

	queryParams := request.URL.Query()
	var topic string

	if len(queryParams["type"]) > 0 {
		topic = strings.ToUpper(strings.TrimSpace(queryParams["type"][0]))
	}

	if !utility.Contains(kafka_utils.SUPPORT_MESSAGE_TYPES, topic) {
		LOGGER.Errorf("Error while parsing message parameter : Specified message type is not supported")
		response.NotifyWWError(w, request, http.StatusBadRequest, "GATEWAY-1004", nil)
		return
	}

	topicName := operation.homeDomain + "_" + topic

	var timeout int
	var retryInterval time.Duration
	var retryTimes float64

	// default value of retry time interval is 3 seconds
	if _, exists := os.LookupEnv(global_environment.ENV_KEY_SERVICE_RETRY_INTERVAL); !exists {
		retryInterval = time.Duration(3)
	} else {
		temp, _ := strconv.Atoi(os.Getenv(global_environment.ENV_KEY_SERVICE_RETRY_INTERVAL))
		retryInterval = time.Duration(temp)
	}

	// default value of retry times is infinite
	if _, exists := os.LookupEnv(global_environment.ENV_KEY_SERVICE_RETRY_TIMES); !exists {
		retryTimes = math.Inf(1)
	} else {
		retryTimes, _ = strconv.ParseFloat(os.Getenv(global_environment.ENV_KEY_SERVICE_RETRY_TIMES), 64)
	}

	// default value of timeout is 3 seconds
	if _, exists := os.LookupEnv(environment.ENV_KEY_MESSAGE_RETRIEVE_TIMEOUT); !exists {
		timeout = 3000
	} else {
		timeout, _ = strconv.Atoi(os.Getenv(environment.ENV_KEY_MESSAGE_RETRIEVE_TIMEOUT))
	}
	/*
		reading meta data from Kafka
	*/

	var res *go_kafka.Metadata
	var err error
	err = common.Retry(retryTimes, retryInterval*time.Second, func() error {
		res, err = operation.Consumer[topic].GetMetadata(&topicName, false, timeout)
		if err != nil {
			LOGGER.Errorf("Encounter error while getting topic meta data %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		LOGGER.Errorf("Encounter error while retrying retrieve topic meta data %v", err)
		response.NotifyWWError(w, request, http.StatusInternalServerError, "GATEWAY-1005", err)
		return
	}
	partitionCount := len(res.Topics[topicName].Partitions)
	if partitionCount <= 0 {
		LOGGER.Errorf("The partition number for topic %v is %v", topicName, partitionCount)
		response.NotifyWWError(w, request, http.StatusInternalServerError, "GATEWAY-1005", err)
		return
	}

	/*
		start reading message from Kafka
	*/
	var messages []*go_kafka.Message
	var newStart go_kafka.TopicPartitions
	*operation.StartingPartition[topic] = *operation.StartingPartition[topic] % int32(partitionCount)
	visitingPartition := *operation.StartingPartition[topic]
	LOGGER.Infof("----- Detected %+v partitions for topic %v -----", partitionCount, topicName)
	LOGGER.Debugf("Starting from partitions %v for topic %v", *operation.StartingPartition[topic], topicName)

	for i := 0; i < partitionCount; i++ {
		visitingPartition := (visitingPartition + int32(i)) % int32(partitionCount)
		LOGGER.Infof("Now visiting partition %v of %v", visitingPartition, topicName)
		newStart, messages, err = kafkaHandler.ReqConsumer(operation.Consumer[topic], topicName, visitingPartition)
		if err != nil && err.Error() != utility.STATUS_QUEUE_EMPTY {
			LOGGER.Errorf("Encounter error while consuming message: %v", err)

			// entering retry loop
			err = common.Retry(retryTimes, retryInterval*time.Second, func() error {

				// reinitialize kafka consumer first
				err := operation.InitializeConsumer(topic)
				if err != nil {
					LOGGER.Errorf("Failed re-intializing Kafka consumer: %v", err)
					return err
				}

				// resend a request to retrieve message
				newStart, messages, err = kafkaHandler.ReqConsumer(operation.Consumer[topic], topicName, visitingPartition)
				if err != nil && err.Error() != utility.STATUS_QUEUE_EMPTY {
					LOGGER.Errorf("Failed re-fetching message from kafka: %v", err)
					return err
				}
				return nil
			})
			if err != nil {
				LOGGER.Errorf("Failed retrieving message from Kafka consumer: %v", err)
				response.NotifyWWError(w, request, http.StatusInternalServerError, "GATEWAY-1005", err)
				return
			}
		}

		if len(newStart) != 0 {
			break
		} else {
			LOGGER.Infof("No new message at partition %v", visitingPartition)
		}
	}

	/*
		gathering necessary info to return to client
	*/
	var results = model.GatewayResponse{}
	for _, elem := range messages {
		//LOGGER.Debugf("msg key %+v", string(elem.Value))
		var temp map[string]interface{}
		_ = json.Unmarshal(elem.Value, &temp)
		results.Data = append(results.Data, temp)
	}

	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	results.Timestamp = &timestamp
	b, _ := json.Marshal(&results)

	LOGGER.Infof("Get message from Kafka successfully")
	response.Respond(w, http.StatusOK, b)

	/*
		update offset after successfully return the message result
	*/

	// no new message discovered, skip offset commit part
	if len(newStart) == 0 {
		LOGGER.Infof("No offset await commit, end of the handler")
		*operation.StartingPartition[topic]++
		return
	}

	LOGGER.Infof("Now committing offset to %v[%v]@%v", topicName, newStart[0].Partition, newStart[0].Offset)

	part, err := operation.Consumer[topic].CommitOffsets(newStart)
	if err != nil {
		LOGGER.Warningf("error committing %s", err.Error())
	}

	/*
		check if offset is successfully commited
	*/
	part, err = operation.Consumer[topic].Committed(newStart, timeout)
	if err != nil {
		LOGGER.Errorf("error fetching committed offset: %v", err)
	}

	if len(part) == 0 {
		LOGGER.Infof("Nothing to commit, move to next partition")
	} else {
		LOGGER.Infof("Offset %+v commited to topic: %v at partition %v", part[0].Offset, topicName, part[0].Partition)
	}
	*operation.StartingPartition[topic] = part[0].Partition + 1
	return

}

func (operation GatewayOperations) InitializeConsumer(topic string) error {
	var err error
	LOGGER.Debugf("****** Initializing Kafka consumer ******")
	if operation.Consumer[topic] != nil {
		operation.Consumer[topic].Unsubscribe()
		operation.Consumer[topic].Unassign()
		operation.Consumer[topic].Close()
	}
	operation.Consumer[topic], err = kafka.Initialize()
	if err != nil {
		LOGGER.Errorf("Failed intializing Kafka consumer: %v", err)
		return err
	}

	LOGGER.Debugf("****** Kafka consumer successfully initialized ******")
	return nil
}

/*
func (operation GatewayOperations) ResetOffset(w http.ResponseWriter, request *http.Request) {
	LOGGER.Infof("gateway-service:Gateway Operations :reset offset")

	queryParams := request.URL.Query()
	var topic string

	if len(queryParams["type"]) > 0 {
		topic = strings.ToUpper(queryParams["type"][0])
	}

	if !utility.Contains(kafka_utils.SUPPORT_MESSAGE_TYPES, topic) {
		LOGGER.Errorf("Error while parsing message parameter : Specified message type is not supported")
		response.NotifyWWError(w, request, http.StatusBadRequest, "GATEWAY-1004", nil)
		return
	}

	topicName := operation.homeDomain + "_" + topic

	LOGGER.Infof("Resetting offset of topic %v to 0", topicName)
	err := kafkaHandler.ResetOffset(operation.Consumer[topic], topicName)
	if err != nil {
		LOGGER.Errorf("Error while resetting offset of Kafka :  %v", err)
		return
	}

	LOGGER.Infof("Reset offset of Kafka successfully")
	response.Respond(w, http.StatusOK, []byte(`{"status":"Success"}`))
	return

}
*/
