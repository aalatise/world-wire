package parse

import (
	"errors"
	"os"

	"github.com/IBM/world-wire/iso20022/pacs00200109"
	"github.com/IBM/world-wire/payment/constant"
	"github.com/IBM/world-wire/payment/environment"
	"github.com/IBM/world-wire/payment/utils"
)

func (handler *ResponseHandler) KafkaErrorRouter(xmlMsgType, msgId, instructionId, ofiId, rfiId string, statusCode int, generateReport bool) (string, []byte, error) {
	var report []byte
	var targetParticipant string
	BIC := os.Getenv(environment.ENV_KEY_PARTICIPANT_BIC)

	switch xmlMsgType {
	case constant.PACS008:
		targetParticipant = ofiId
	case constant.IBWF001:
		targetParticipant = rfiId
	case constant.CAMT056:
		targetParticipant = ofiId
	case constant.PACS004:
		targetParticipant = rfiId
	case constant.CAMT029:
		targetParticipant = rfiId
	case constant.IBWF002:
		targetParticipant = ofiId
	case constant.PACS009:
		targetParticipant = ofiId
	case constant.PACS002:
		targetParticipant = rfiId
	case constant.CAMT026:
		targetParticipant = rfiId
	case constant.CAMT087:
		targetParticipant = ofiId
	default:
		LOGGER.Errorf("No matching XML message type found")
		return "", nil, errors.New("No matching XML message type found")
	}

	if generateReport {
		if utils.Contains(constant.SUPPORT_CAMT_MESSAGES, xmlMsgType) {
			report = handler.CreateCamt030(BIC, msgId, instructionId, xmlMsgType, targetParticipant, statusCode)
		} else {
			originalGrpInf := &pacs00200109.OriginalGroupInformation29{
				OrgnlMsgId:   getReportMax35Text(msgId),
				OrgnlMsgNmId: getReportMax35Text(xmlMsgType),
			}
			report = handler.CreatePacs002(BIC, instructionId, targetParticipant, statusCode, originalGrpInf)
		}
	}
	return targetParticipant, report, nil
}
