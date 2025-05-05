package permission_proxy

import "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"

type PermissionProxy struct {
	conn       db_conn.DbConnection
	permission *Permissions
}

func NewPermissionProxy(dbConn db_conn.DbConnection) (*PermissionProxy, error) {
	permissionJs := dbConn.GetDatabaseMeta().GetPermissionJs()
	permission, err := NewPermissionFromJs(permissionJs)
	if err != nil {
		return nil, err
	}

	return &PermissionProxy{
		conn:       dbConn,
		permission: permission,
	}, nil
}

func (p *PermissionProxy) CanView(params CanViewParams) bool {
	return p.permission.CanView(params)
}

func (p *PermissionProxy) CanCreate(params CanCreateParams) bool {
	return p.permission.CanCreate(params)
}

func (p *PermissionProxy) CanUpdate(params CanUpdateParams) bool {
	return p.permission.CanUpdate(params)
}

func (p *PermissionProxy) CanDelete(params CanDeleteParams) bool {
	return p.permission.CanDelete(params)
}
