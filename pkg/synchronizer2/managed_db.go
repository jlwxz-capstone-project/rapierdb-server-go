package synchronizer2

import (
	"context"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/permission_proxy"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query_executor"
)

type ManagedDb struct {
	conn db_conn.DbConnection
	// each managed db has its own context
	// so that it can be stopped independently
	ctx             context.Context
	cancel          context.CancelFunc
	queryExecutor   *query_executor.QueryExecutor
	permissionProxy *permission_proxy.PermissionProxy
	queryManager    *QueryManager
}
