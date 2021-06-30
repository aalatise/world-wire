package kafka

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.ibm.com/gftn/world-wire-services/ww-gateway/environment"
	"github.ibm.com/gftn/world-wire-services/ww-gateway/utility"
)

func ReqConsumer(c *kafka.Consumer, topic string, partition int32) (kafka.TopicPartitions, []*kafka.Message, error) {

	LOGGER.Debug("\n*******************")
	LOGGER.Infof("Requesting message from Kafka topic %v[%v]", topic, partition)
	var startOffset int64

	var timeout int
	// default value of timeout is 3 seconds
	if _, exists := os.LookupEnv(environment.ENV_KEY_MESSAGE_RETRIEVE_TIMEOUT); !exists {
		timeout = 3000
	} else {
		timeout, _ = strconv.Atoi(os.Getenv(environment.ENV_KEY_MESSAGE_RETRIEVE_TIMEOUT))
	}

	/*
		retrieving lowest/highest offset in the partition
	*/
	lowOffset, highOffset, err := c.QueryWatermarkOffsets(topic, partition, timeout)
	if err != nil {
		LOGGER.Errorf("Encounter error while querying the last commit offset %v", err)
		return kafka.TopicPartitions{}, []*kafka.Message{}, err
	}

	LOGGER.Infof("Detect %v messages in %v[%v]", highOffset-lowOffset, topic, partition)

	if lowOffset == highOffset {
		LOGGER.Warningf("Topic %v[%v] is empty", topic, partition)
		return kafka.TopicPartitions{}, []*kafka.Message{}, errors.New(utility.STATUS_QUEUE_EMPTY)
	}

	/*
		retrieving the latest read offset in the partition
	*/
	part, err := c.Committed(kafka.TopicPartitions{{Topic: &topic, Partition: partition}}, timeout)
	if err != nil {
		LOGGER.Errorf("error fetching commit %v", err)
		return kafka.TopicPartitions{}, []*kafka.Message{}, err
	}
	if len(part) == 0 {
		LOGGER.Infof("No committed offset detected, starting from low offset: %v", lowOffset)
		startOffset = lowOffset
	} else {
		commitedOffset := int64(part[0].Offset)
		if commitedOffset < lowOffset {
			startOffset = lowOffset
		} else {
			startOffset = commitedOffset
		}
	}

	// if starting offset is the same as the highest offset in this partition, it means no new message
	if startOffset == highOffset {
		LOGGER.Infof("No new message to consume from %v[%v]@%v", topic, partition, startOffset)
		return kafka.TopicPartitions{}, []*kafka.Message{}, errors.New(utility.STATUS_QUEUE_EMPTY)
	}

	LOGGER.Infof("Continue receiving message at offset: %v for partition %v", startOffset, partition)

	/*
		calculate which offset to stop for this round
	*/
	count, _ := strconv.Atoi(os.Getenv(environment.ENV_KEY_BATCH_LIMIT))
	count64 := int64(count)

	var upperBound int64
	if upperBound = startOffset + count64; upperBound >= highOffset {
		upperBound = highOffset
	}

	var partitions kafka.TopicPartitions

	/*
		assigning starting offset & topic/partition
	*/

	tempOffSet, err := kafka.NewOffset(fmt.Sprintf("%v", startOffset))
	if err != nil {
		return kafka.TopicPartitions{}, []*kafka.Message{}, err
	}
	partitions = append(partitions, kafka.TopicPartition{
		Topic:     &topic,
		Partition: partition,
		Offset:    tempOffSet,
		Error:     err,
	})

	LOGGER.Infof("Assgining partition & offset to consumer manually")

	err = c.Assign(partitions)
	if err != nil {
		LOGGER.Errorf("Assign failed: %s", err)
		return kafka.TopicPartitions{}, []*kafka.Message{}, err
	}
	defer c.Unassign()

	/*
		assignment, err := c.Assignment()
		if err != nil {
			LOGGER.Errorf("Assignment() failed: %s", err)
			return kafka.TopicPartitions{}, []*kafka.Message{}, err
		}
		LOGGER.Infof("Assignment %v\n", assignment)
	*/
	LOGGER.Debugf("****** Start receiving message from Kafka offset ******")

	/*
		start getting message from starting offset to upperbound
	*/

	var actualVisitedIndex int64
	var messages []*kafka.Message
Loop:
	for actualVisitedIndex = startOffset; actualVisitedIndex < upperBound; actualVisitedIndex++ {
		ev := c.Poll(timeout)

		switch e := ev.(type) {
		case *kafka.Message:
			if e.TopicPartition.Error != nil {
				LOGGER.Infof("Encounter error: %+v", e)
				return kafka.TopicPartitions{}, []*kafka.Message{}, e.TopicPartition.Error
			}
			LOGGER.Infof("Received message from Kafka: %+v", e)
			/*
				_, marshalErr := retrieveInstructionId(e.Value)
				if marshalErr != nil {
					LOGGER.Errorf("%v", marshalErr)
				}
			*/
			messages = append(messages, e)
		case kafka.Error:
			LOGGER.Errorf("Error: %+v", e)
			return kafka.TopicPartitions{}, []*kafka.Message{}, e
		case kafka.PartitionEOF:
			LOGGER.Warningf("End of partition")
			break Loop
		default:
			LOGGER.Errorf("Message retrieval timeout")
			return kafka.TopicPartitions{}, []*kafka.Message{}, errors.New("Message retrieval timeout")
		}

	}
	LOGGER.Debug("****** Complete receiving message ******")

	/*
		Recording the last offset being read, and commit it manually to Kafka after respond to the client
	*/

	var finalPartition kafka.TopicPartitions

	LOGGER.Infof("Newly visited offset: %v at partition %v", actualVisitedIndex, partition)
	tempOffSet, err = kafka.NewOffset(fmt.Sprintf("%v", actualVisitedIndex))
	if err != nil {
		LOGGER.Errorf(err.Error())
		return kafka.TopicPartitions{}, []*kafka.Message{}, err
	}
	finalPartition = append(finalPartition, kafka.TopicPartition{
		Topic:     &topic,
		Partition: partition,
		Offset:    tempOffSet,
		Error:     err,
	})

	return finalPartition, messages, err

}

/* for testing purpose
func ResetOffset(c *kafka.Consumer, topic string) error {
	LOGGER.Infof("Resetting message offset of Kafka topic %v", topic)
	part, err := c.Committed(kafka.TopicPartitions{{Topic: &topic, Partition: 0}}, -1)
	if err != nil {
		LOGGER.Errorf("error fetching commit %v", err)
	}

	if len(part) == 0 {
		err := errors.New("Cannot find the specified kafka topic")
		LOGGER.Errorf(err.Error())
		return err
	}
	LOGGER.Infof("Current offset is %+v", part[0].Offset)

	var resetPartition kafka.TopicPartitions

	tempOffSet, err := kafka.NewOffset(fmt.Sprintf("%v", 0))
	if err != nil {
		return err
	}
	resetPartition = append(resetPartition, kafka.TopicPartition{
		Topic:     &topic,
		Partition: 0,
		Offset:    tempOffSet,
		Error:     err,
	})

	part, err = c.CommitOffsets(resetPartition)
	if err != nil {
		LOGGER.Errorf("error committing %v", err)
		return err
	}
	LOGGER.Infof("commiting offset %+v", part)

	part, err = c.Committed(kafka.TopicPartitions{{Topic: &topic, Partition: 0}}, -1)
	if err != nil {
		LOGGER.Errorf("error fetching commit %v", err)
		return err
	}
	LOGGER.Infof("The offset of %v topic is now %+v", topic, part)
	return nil
}
*/
/*
func retrieveInstructionId(raw []byte) (string, error) {

	var rawMsg model.SendPacs
	_ = json.Unmarshal(raw, &rawMsg)
	if rawMsg.Message == nil {
		LOGGER.Errorf("%+v", string(raw))
		return "", errors.New("error marshaling")
	}

	targetMsgName := constant.PACS002

	//decode base 64
	data, err := base64.StdEncoding.DecodeString(*rawMsg.Message)
	if err != nil {
		LOGGER.Errorf("Error retrieving instruction id of the message", err.Error())
		return "", err
	}

	// retrieve message name

	payloadDocument := etree.NewDocument()
	err = payloadDocument.ReadFromString(string(data))
	if err != nil {
		return "", errors.New("Error while parsing XML into DOM")
	}

	appHeaderElement := xmldsig.GetElementByName(payloadDocument.Root(), "AppHdr")
	if appHeaderElement == nil {
		LOGGER.Infof("AppHdr is missing")
		return "", errors.New("Payload doesn't have AppHdr")
	}

	messageNameElement := xmldsig.GetElementByName(appHeaderElement, "MsgDefIdr")
	if messageNameElement == nil {
		LOGGER.Infof("MsgDefIdr is missing")
		return "", errors.New("MsgDefIdr is missing")
	}

	msgName := messageNameElement.Text()
	if msgName != targetMsgName {
		LOGGER.Infof("Not target message type %v, skipped", targetMsgName)
		return "", nil
	}

	// get tx status

	fIToFIPmtStsRptElement := xmldsig.GetElementByName(payloadDocument.Root(), "FIToFIPmtStsRpt")
	if fIToFIPmtStsRptElement == nil {
		LOGGER.Infof("FIToFIPmtStsRpt is missing")
		return "", errors.New("Payload doesn't have FIToFIPmtStsRpt")
	}

	txInfAndStsElement := xmldsig.GetElementByName(fIToFIPmtStsRptElement, "TxInfAndSts")
	if txInfAndStsElement == nil {
		LOGGER.Infof("TxInfAndSts is missing")
		return "", errors.New("Payload doesn't have TxInfAndSts")
	}

	txStsElement := xmldsig.GetElementByName(txInfAndStsElement, "TxSts")
	if txStsElement == nil {
		LOGGER.Infof("TxSts is missing")
		return "", errors.New("Payload doesn't have TxSts")
	}

	msgStatus := txStsElement.Text()

	var messageInstance = &message_converter.Pacs002{Raw: data}
	err = messageInstance.RequestToStruct()
	if err != nil {
		LOGGER.Errorf("Constructing to go struct failed: %v", err.Error())
		return "", err
	}

	var instructionId string
	if messageInstance.Message.Body.TxInfAndSts[0].OrgnlInstrId != nil {
		instructionId = string(*messageInstance.Message.Body.TxInfAndSts[0].OrgnlInstrId)
		if msgStatus != constant.PAYMENT_STATUS_ACTC {
			LOGGER.Infof("%v error instrId --- %v ---", msgStatus, instructionId)
		} else {
			LOGGER.Infof("%v success instrId --- %v ---", msgStatus, instructionId)
		}
	} else {
		LOGGER.Errorf("instruction id field is empty")
		return "", errors.New("instruction id empty")
	}

	return instructionId, nil
}
*/
