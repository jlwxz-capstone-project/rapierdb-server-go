package main

import (
	"errors"
	"fmt"
	"testing"

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
