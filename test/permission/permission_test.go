package main

import (
	_ "embed"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/permissions"
)

//go:embed test_permission1.js
var testPermission1 string

func TestPermissionFromJs(t *testing.T) {
	permissions.NewPermissionFromJs(testPermission1)
}
