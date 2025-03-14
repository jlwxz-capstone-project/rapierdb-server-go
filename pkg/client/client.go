package client

import (
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

type TestClient struct {
	Docs    map[string]map[string]*loro.LoroDoc
	Queries map[string]ReactiveQuery
}

func NewTestClient() *TestClient {
	return &TestClient{
		Docs:    make(map[string]map[string]*loro.LoroDoc),
		Queries: make(map[string]ReactiveQuery),
	}
}
