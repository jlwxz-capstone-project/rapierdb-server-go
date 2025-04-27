package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/message/v1"
	"github.com/stretchr/testify/assert"
)

//go:embed message_encoding_decoding.jsonl
var msgLines string

type testCase struct {
	Name    string `json:"name"`
	Encoded string `json:"encoded"`
}

func TestMessageEncodingDecoding(t *testing.T) {
	lines := strings.Split(msgLines, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var tc testCase
		err := json.Unmarshal([]byte(line), &tc)
		assert.NoError(t, err)

		t.Run(tc.Name, func(t *testing.T) {
			decoded, err := base64.StdEncoding.DecodeString(tc.Encoded)
			assert.NoError(t, err)
			buf := bytes.NewBuffer(decoded)
			_, err = message.DecodeMessage(buf)
			assert.NoError(t, err)
		})
	}
}
