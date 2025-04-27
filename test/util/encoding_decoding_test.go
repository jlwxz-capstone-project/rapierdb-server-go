package util

import (
	"bytes"
	"math"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestWriteAndReadVarUint(t *testing.T) {
	testCases := []struct {
		name    string
		input   uint64
		wantErr bool
	}{
		{
			name:    "small number",
			input:   1,
			wantErr: false,
		},
		{
			name:    "medium number",
			input:   127,
			wantErr: false,
		},
		{
			name:    "large number",
			input:   128,
			wantErr: false,
		},
		{
			name:    "very large number",
			input:   uint64(1) << 63,
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}

			// Test encoding
			err := util.WriteVarUint(buf, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Test decoding
			decoded, err := util.ReadVarUint(buf)
			assert.NoError(t, err)
			assert.Equal(t, tc.input, decoded)
		})
	}
}

func TestWriteAndReadVarString(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty string",
			input:   "",
			wantErr: false,
		},
		{
			name:    "short string",
			input:   "hello",
			wantErr: false,
		},
		{
			name:    "long string",
			input:   "this is a very long string for testing purposes",
			wantErr: false,
		},
		{
			name:    "string with special characters",
			input:   "Hello, 世界！",
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}

			// Test encoding
			err := util.WriteVarString(buf, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Test decoding
			decoded, err := util.ReadVarString(buf)
			assert.NoError(t, err)
			assert.Equal(t, tc.input, decoded)
		})
	}
}

func TestWriteAndReadUint32(t *testing.T) {
	testCases := []struct {
		name    string
		input   uint32
		wantErr bool
	}{
		{
			name:    "zero",
			input:   0,
			wantErr: false,
		},
		{
			name:    "small number",
			input:   42,
			wantErr: false,
		},
		{
			name:    "max uint32",
			input:   4294967295,
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}

			// Test encoding
			err := util.WriteUint32(buf, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Test decoding
			decoded, err := util.ReadUint32(buf)
			assert.NoError(t, err)
			assert.Equal(t, tc.input, decoded)
		})
	}
}

func TestWriteAndReadInt8(t *testing.T) {
	testCases := []struct {
		name    string
		input   int8
		wantErr bool
	}{
		{
			name:    "zero",
			input:   0,
			wantErr: false,
		},
		{
			name:    "positive",
			input:   127,
			wantErr: false,
		},
		{
			name:    "negative",
			input:   -128,
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := util.WriteInt8(buf, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			decoded, err := util.ReadInt8(buf)
			assert.NoError(t, err)
			assert.Equal(t, tc.input, decoded)
		})
	}
}

func TestWriteAndReadInt64(t *testing.T) {
	testCases := []struct {
		name    string
		input   int64
		wantErr bool
	}{
		{
			name:    "zero",
			input:   0,
			wantErr: false,
		},
		{
			name:    "positive",
			input:   9223372036854775807, // math.MaxInt64
			wantErr: false,
		},
		{
			name:    "negative",
			input:   -9223372036854775808, // math.MinInt64
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := util.WriteInt64(buf, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			decoded, err := util.ReadInt64(buf)
			assert.NoError(t, err)
			assert.Equal(t, tc.input, decoded)
		})
	}
}

func TestWriteAndReadFloat64(t *testing.T) {
	testCases := []struct {
		name    string
		input   float64
		wantErr bool
	}{
		{
			name:    "zero",
			input:   0,
			wantErr: false,
		},
		{
			name:    "positive",
			input:   math.MaxFloat64,
			wantErr: false,
		},
		{
			name:    "negative",
			input:   -math.MaxFloat64,
			wantErr: false,
		},
		{
			name:    "small",
			input:   math.SmallestNonzeroFloat64,
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := util.WriteFloat64(buf, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			decoded, err := util.ReadFloat64(buf)
			assert.NoError(t, err)
			assert.Equal(t, tc.input, decoded)
		})
	}
}

func TestWriteAndReadBytes(t *testing.T) {
	testCases := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "empty",
			input:   []byte{},
			wantErr: false,
		},
		{
			name:    "small",
			input:   []byte{1, 2, 3, 4, 5},
			wantErr: false,
		},
		{
			name:    "large",
			input:   bytes.Repeat([]byte{1}, 1000),
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := util.WriteVarByteArray(buf, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			decoded, err := util.ReadVarByteArray(buf)
			assert.NoError(t, err)
			assert.Equal(t, tc.input, decoded)
		})
	}
}

func TestWriteAndReadVarInt(t *testing.T) {
	testCases := []struct {
		name    string
		input   int64
		wantErr bool
	}{
		{
			name:    "zero",
			input:   0,
			wantErr: false,
		},
		{
			name:    "small positive",
			input:   127,
			wantErr: false,
		},
		{
			name:    "small negative",
			input:   -127,
			wantErr: false,
		},
		{
			name:    "large positive",
			input:   9223372036854775807, // math.MaxInt64
			wantErr: false,
		},
		{
			name:    "large negative",
			input:   -9223372036854775808, // math.MinInt64
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := util.WriteVarInt(buf, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			decoded, err := util.ReadVarInt(buf)
			assert.NoError(t, err)
			assert.Equal(t, tc.input, decoded)
		})
	}
}
