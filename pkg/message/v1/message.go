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
	Type() uint64
	Encode() ([]byte, error)
}

// 消息类型
const (
	MSG_TYPE_POST_TRANSACTION_V1    uint64 = 1
	MSG_TYPE_SUBSCRIPTION_UPDATE_V1 uint64 = 2
	MSG_TYPE_POST_DOC_V1            uint64 = 3
	MSG_TYPE_ACK_TRANSACTION_V1     uint64 = 4
	MSG_TYPE_TRANSACTION_FAILED_V1  uint64 = 5
	MSG_TYPE_VERSION_QUERY_V1       uint64 = 6
	MSG_TYPE_VERSION_QUERY_RESP_V1  uint64 = 7
	MSG_TYPE_SUBSCRIPTION_RESET_V1  uint64 = 8
)

func DecodeMessage(b *bytes.Buffer) (Message, error) {
	msgType, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}

	switch msgType {
	case MSG_TYPE_SUBSCRIPTION_UPDATE_V1:
		return decodeSubscriptionUpdateMessageV1Body(b)
	case MSG_TYPE_ACK_TRANSACTION_V1:
		return decodeAckTransactionMessageV1Body(b)
	case MSG_TYPE_TRANSACTION_FAILED_V1:
		return decodeTransactionFailedMessageV1Body(b)
	case MSG_TYPE_POST_TRANSACTION_V1:
		return decodePostTransactionMessageV1Body(b)
	case MSG_TYPE_POST_DOC_V1:
		return decodeSyncMessageV1Body(b)
	case MSG_TYPE_VERSION_QUERY_V1:
		return decodeVersionQueryMessageV1Body(b)
	case MSG_TYPE_VERSION_QUERY_RESP_V1:
		return decodeVersionQueryRespMessageV1Body(b)
	case MSG_TYPE_SUBSCRIPTION_RESET_V1:
		return decodeSubscriptionResetMessageV1Body(b)
	default:
		return nil, errors.New("未知的消息类型")
	}
}
