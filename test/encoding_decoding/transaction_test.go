package main

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/stretchr/testify/assert"
)

//go:embed encoding_decoding_transaction
var transactionLines string

func TestTransactionDecoding(t *testing.T) {
	lines := strings.Split(transactionLines, "\n")
	for i, line := range lines {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			if line == "" {
				t.Skip()
			}
			decoded, err := base64.StdEncoding.DecodeString(line)
			assert.NoError(t, err)
			_, err = storage_engine.DecodeTransaction(decoded)
			assert.NoError(t, err)
		})
	}
}
