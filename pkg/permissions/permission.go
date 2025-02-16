package permissions

type Permission struct {
	Version string                              `json:"version"`
	Rules   map[string]CollectionPermissionRule `json:"rules"`
}

type CollectionPermissionRule struct {
	CanView   func(doc interface{}) bool
	CanCreate func(doc interface{}) bool
	CanUpdate func(doc interface{}) bool
	CanDelete func(doc interface{}) bool
}
