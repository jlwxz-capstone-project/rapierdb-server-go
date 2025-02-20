package permissions

import (
	_ "embed"
	"strings"

	"github.com/dop251/goja"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage"
)

type Permission[CTX any] struct {
	Version string                                   `json:"version"`
	Rules   map[string]CollectionPermissionRule[CTX] `json:"rules"`
}

type CollectionPermissionRule[CTX any] struct {
	CanView   func(doc storage.LoadedDoc, ctx CTX) bool
	CanCreate func(doc storage.LoadedDoc, ctx CTX) bool
	CanUpdate func(doc storage.LoadedDoc, ctx CTX) bool
	CanDelete func(doc storage.LoadedDoc, ctx CTX) bool
}

//go:embed permission_builder.js
var permissionBuilderScript string

func NewPermissionFromJs[CTX any](js string) (*Permission[CTX], error) {
	vm := goja.New()
	js = permissionBuilderScript + "\n var permission = " + strings.Trim(js, "\n ")
	vm.RunString(js)
	permissionJs := vm.Get("permission").ToObject(vm)
	version := permissionJs.Get("version").ToString().String()
	rules := permissionJs.Get("rules").ToObject(vm)
	permission := &Permission[CTX]{
		Version: version,
		Rules:   make(map[string]CollectionPermissionRule[CTX]),
	}
	for _, key := range rules.GetOwnPropertyNames() {
		// value := rules.Get(key).ToObject(vm)
		// canViewJs, _ := goja.AssertFunction(value.Get("canView"))
		// canCreateJs, _ := goja.AssertFunction(value.Get("canCreate"))
		// canUpdateJs, _ := goja.AssertFunction(value.Get("canUpdate"))
		// canDeleteJs, _ := goja.AssertFunction(value.Get("canDelete"))
		canView := func(doc storage.LoadedDoc, ctx CTX) bool {
			return true
		}
		canCreate := func(doc storage.LoadedDoc, ctx CTX) bool {
			return true
		}
		canUpdate := func(doc storage.LoadedDoc, ctx CTX) bool {
			return true
		}
		canDelete := func(doc storage.LoadedDoc, ctx CTX) bool {
			return true
		}
		permission.Rules[key] = CollectionPermissionRule[CTX]{
			CanView:   canView,
			CanCreate: canCreate,
			CanUpdate: canUpdate,
			CanDelete: canDelete,
		}
	}
	return permission, nil
}
