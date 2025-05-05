package query

import (
	"encoding/json"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	pe "github.com/pkg/errors"
)

const (
	FIND_MANY_QUERY_TYPE uint64 = 1
	FIND_ONE_QUERY_TYPE  uint64 = 2
)

type Query interface {
	log.DebugPrintable
	isQuery()
	Encode() ([]byte, error)
}

type DocWithId struct {
	DocId string
	Doc   *loro.LoroDoc
}

func DecodeQuery(data []byte) (Query, error) {
	var temp struct {
		Type uint64 `json:"type"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, err
	}
	switch temp.Type {
	case FIND_MANY_QUERY_TYPE:
		return DecodeFindManyQuery(data)
	case FIND_ONE_QUERY_TYPE:
		return DecodeFindOneQuery(data)
	default:
		return nil, pe.Errorf("unknown query type: %d", temp.Type)
	}
}
