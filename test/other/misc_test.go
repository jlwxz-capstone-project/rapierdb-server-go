package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"unsafe"

	pe "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	err := pe.WithStack(errors.New("test error"))
	fmt.Printf("%+v\n", err)
}

type DebugPrintable interface {
	DebugPrint() string
}

type StructA struct {
	field1 int
	field2 int
	DebugPrintable
}

func (s *StructA) DebugPrint() string {
	return fmt.Sprintf("StructA{field1: %d, field2: %d}", s.field1, s.field2)
}

var _ DebugPrintable = &StructA{}

func TestNil(t *testing.T) {
	var v1 *int
	assert.True(t, v1 == nil)

	var v2 any
	assert.True(t, v2 == nil)

	v2 = v1
	assert.True(t, v2 == nil)
}

func TestJson(t *testing.T) {
	val := map[string]any{
		"field1": "value1",
		"field2": 2,
	}

	json, err := json.Marshal(val)
	assert.NoError(t, err)
	fmt.Println(string(json))
}

func IsNil(v any) bool {
	return (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0
}

func TestIsNil(t *testing.T) {
	var v1 chan int = nil
	assert.True(t, IsNil(v1))

	var v2 *int = nil
	assert.True(t, IsNil(v2))

	var v5 any = nil
	assert.True(t, IsNil(v5))

	assert.True(t, IsNil(nil))
}
