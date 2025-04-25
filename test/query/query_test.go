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
	qfe "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/query_filter_expr"
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
		fmt.Println("--------------" + testCase.Name + "-----------------")
		fmt.Println(testCase.Expr.DebugPrint())
		res, err := testCase.Expr.Eval(testCase.Doc)
		if err != nil {
			log.Errorf("failed to eval expr: %s", err)
		}
		fmt.Printf("expected: %v, actual: %v\n", testCase.Expected, res)
	}
}
