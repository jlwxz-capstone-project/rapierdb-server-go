package subscription

import (
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
)

type Subscription struct {
	Collection string
	Query      query.Query
}
