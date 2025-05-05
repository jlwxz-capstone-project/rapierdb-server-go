package db_connector

import (
	"context"
	"net/url"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
)

type PebbleConnector struct{}

var _ DbConnector = &PebbleConnector{}

func NewPebbleConnector() *PebbleConnector {
	return &PebbleConnector{}
}

// Connect establishes a connection to a Pebble database.
//
// It will parse the dbUrl to get the db path, and the options. A valid dbUrl is like:
//
//	rapierdb://<db-path>?readonly=true&enableWal=false
func (c *PebbleConnector) ConnectWithContext(ctx context.Context, dbUrl string) (db_conn.DbConnection, error) {
	parsedUrl, err := url.Parse(dbUrl)
	if err != nil {
		return nil, err
	}

	dbPath := parsedUrl.Path
	log.Debugf("PebbleConnector.ConnectWithContext: dbPath=%s", dbPath)

	queryParams := parsedUrl.Query()
	_ = queryParams

	return db_conn.NewPebbleDbConnWithContext(
		ctx,
		&db_conn.PebbleDbConnParams{
			Path: dbPath,
		},
	)
}
