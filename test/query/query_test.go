package main

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	qfe "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/query_filter_expr"
	"github.com/stretchr/testify/assert"
)

//go:embed qfe_tests.jsonl
var qfeTests string

type QfeTestCase struct {
	Name     string
	Expr     qfe.QueryFilterExpr
	Doc      *loro.LoroDoc
	Expected *qfe.ValueExpr
}

func prepareTestCases() []QfeTestCase {
	lines := strings.Split(qfeTests, "\n")
	testCases := make([]QfeTestCase, 0, len(lines))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		var temp struct {
			Name        string `json:"name"`
			Expr        string `json:"expr"`
			DocSnapshot string `json:"docSnapshot"`
			Expected    string `json:"expected"`
		}
		err := json.Unmarshal([]byte(line), &temp)
		if err != nil {
			log.Errorf("failed to unmarshal line: %s", err)
		}

		docSnapshot, err := base64.StdEncoding.DecodeString(temp.DocSnapshot)
		if err != nil {
			log.Errorf("failed to decode doc: %s", err)
		}

		doc := loro.NewLoroDoc()
		doc.Import(docSnapshot)

		expr, err := qfe.NewQueryFilterExprFromJson([]byte(temp.Expr))
		if err != nil {
			log.Errorf("failed to parse expr: %s", err)
		}

		expected, err := qfe.NewQueryFilterExprFromJson([]byte(temp.Expected))
		if err != nil {
			log.Errorf("failed to parse expected: %s", err)
		}

		expectedValueExpr, ok := expected.(*qfe.ValueExpr)
		if !ok {
			log.Errorf("expected is not a value expr: %s", temp.Expected)
		}

		testCases = append(testCases, QfeTestCase{
			Name:     temp.Name,
			Expr:     expr,
			Doc:      doc,
			Expected: expectedValueExpr,
		})
	}

	return testCases
}

func TestQfe(t *testing.T) {
	testCases := prepareTestCases()

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			res, err := testCase.Expr.Eval(testCase.Doc)
			assert.NoError(t, err)
			assert.Equal(t, testCase.Expected.Value, res.Value)
		})
	}
}

func TestFindManyQuery(t *testing.T) {
	q := &query.FindManyQuery{
		Collection: "test",
		Filter: qfe.NewGtExpr(
			qfe.NewFieldValueExpr(qfe.NewValueExpr("age")),
			qfe.NewValueExpr(18),
		),
		Sort: []query.SortField{
			{
				Field: "name",
				Order: query.SortOrderAsc,
			},
		},
		Skip:  0,
		Limit: 10,
	}
	encoded, err := q.Encode()
	assert.NoError(t, err)
	fmt.Println(string(encoded))
	decoded, err := query.DecodeFindManyQuery(encoded)
	assert.NoError(t, err)
	assert.Equal(t, q, decoded)
	fmt.Printf("decoded: %+v\n", decoded)
}
