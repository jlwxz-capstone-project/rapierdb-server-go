package main

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/permissions"
)

//go:embed test_permission1.js
var testPermission1 string

//go:embed test_permission_emp.js
var testPermissionEmp string

func TestPermissionFromJs(t *testing.T) {
	t.Run("test_permission_emp", func(t *testing.T) {
		permission, err := permissions.NewPermissionFromJs(testPermissionEmp)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%+v\n", permission)
	})

	t.Run("test_permission1", func(t *testing.T) {
		permission, err := permissions.NewPermissionFromJs(testPermission1)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%+v\n", permission)
	})
}
