package main

import (
	"fmt"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

func TestCreateDestroyLoroDoc(t *testing.T) {
	_ = loro.NewLoroDoc()
}

func TestUpdateLoroText(t *testing.T) {
	doc := loro.NewLoroDoc()
	text := doc.GetText("test")
	text.UpdateText("Hello, World!")
	if text.ToString() != "Hello, World!" {
		t.Fail()
	}
}

func TestExportSnapshot(t *testing.T) {
	doc := loro.NewLoroDoc()
	text := doc.GetText("test")
	text.UpdateText("Hello, World!")
	vec := doc.ExportSnapshot()
	fmt.Println("size=", vec.GetSize())
	fmt.Println("capacity=", vec.GetCapacity())
}
