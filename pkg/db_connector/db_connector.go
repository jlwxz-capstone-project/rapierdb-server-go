package db_connector

import (
	"context"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"
)

type DbConnector interface {
	ConnectWithContext(ctx context.Context, url string) (db_conn.DbConnection, error)
}
