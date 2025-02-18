package main

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/permissions"
	"github.com/stretchr/testify/assert"
)

//go:embed test_permission1.js
var testPermission1 string

func TestPermissionFromJs(t *testing.T) {
	permission, err := permissions.NewPermissionFromJs[any](testPermission1)
	assert.NoError(t, err)
	fmt.Println(permission)
}
