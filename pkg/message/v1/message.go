package message

import (
	"bytes"
	"errors"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

type Message interface {
	log.DebugPrintable
	isMessage()
	Type() uint8
	Encode() ([]byte, error)
}

// 消息类型
const (
	MSG_TYPE_POST_TRANSACTION_V1    uint8 = 1
	MSG_TYPE_SUBSCRIPTION_UPDATE_V1 uint8 = 2
	// MSG_TYPE_POST_DOC_V1            uint8 = 3
	MSG_TYPE_ACK_TRANSACTION_V1    uint8 = 4
	MSG_TYPE_TRANSACTION_FAILED_V1 uint8 = 5
	MSG_TYPE_VERSION_QUERY_V1      uint8 = 6
	MSG_TYPE_VERSION_QUERY_RESP_V1 uint8 = 7
	MSG_TYPE_SUBSCRIPTION_RESET_V1 uint8 = 8
	MSG_TYPE_SYNC_V1               uint8 = 9
	MSG_TYPE_VERSION_GAP_V1        uint8 = 10
)

func DecodeMessage(b *bytes.Buffer) (Message, error) {
	msgType, err := util.ReadUint8(b)
	if err != nil {
		return nil, err
	}

	switch msgType {
	case MSG_TYPE_POST_TRANSACTION_V1:
		return decodePostTransactionMessageV1(b)
	case MSG_TYPE_SUBSCRIPTION_UPDATE_V1:
		return decodeSubscriptionUpdateMessageV1(b)
	case MSG_TYPE_ACK_TRANSACTION_V1:
		return decodeAckTransactionMessageV1(b)
	case MSG_TYPE_TRANSACTION_FAILED_V1:
		return decodeTransactionFailedMessageV1(b)
	case MSG_TYPE_VERSION_QUERY_V1:
		return decodeVersionQueryMessageV1(b)
	case MSG_TYPE_VERSION_QUERY_RESP_V1:
		return decodeVersionQueryRespMessageV1(b)
	case MSG_TYPE_SUBSCRIPTION_RESET_V1:
		return decodeSubscriptionResetMessageV1(b)
	case MSG_TYPE_SYNC_V1:
		return decodeSyncMessageV1(b)
	case MSG_TYPE_VERSION_GAP_V1:
		return decodeVersionGapMessageV1(b)
	default:
		return nil, errors.New("未知的消息类型")
	}
}
