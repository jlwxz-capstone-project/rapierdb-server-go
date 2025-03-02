package main

import (
	"errors"
	"fmt"
	"testing"

	pe "github.com/pkg/errors"
)

func TestError(t *testing.T) {
	err := pe.WithStack(errors.New("test error"))
	fmt.Printf("%+v\n", err)
}
