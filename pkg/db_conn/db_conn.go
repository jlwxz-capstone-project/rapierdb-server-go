package db_conn

import (
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

type DbConnStatus int32

const (
	DbConnStatusNotReady DbConnStatus = 0
	DbConnStatusOpening  DbConnStatus = 1
	DbConnStatusRunning  DbConnStatus = 2
	DbConnStatusClosing  DbConnStatus = 3
	DbConnStatusError    DbConnStatus = 4
	DbConnStatusClosed   DbConnStatus = 5
)

type DbConnection interface {
	Open() error
	Close() error

	// Connection Params
	// GetConnParams()

	// Database Meta
	GetDatabaseMeta() *DatabaseMeta
	UpdateSchema(newSchema *DatabaseSchema) error
	UpdatePermissionJs(newPermissionJs string) error

	// Query Related
	LoadDoc(collectionName, docID string) (*loro.LoroDoc, error)
	LoadCollection(collectionName string) (map[string]*loro.LoroDoc, error)
	InvalidateCache()

	// Transaction Related
	Commit(tr *Transaction) error

	// Transaction Events
	GetCommittedEb() *util.EventBus[*TransactionCommittedEvent]
	GetRollbackedEb() *util.EventBus[*TransactionRollbackedEvent]

	// Status Related
	GetStatus() DbConnStatus
	SubscribeStatusChange() <-chan DbConnStatus
	UnsubscribeStatusChange(ch <-chan DbConnStatus)
	WaitForStatus(targetStatus DbConnStatus) <-chan struct{}
}
