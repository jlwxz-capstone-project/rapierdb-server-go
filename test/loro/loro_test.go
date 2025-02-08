package main

import (
	"bytes"
	"fmt"
	"testing"
	"time"

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
	fmt.Println("len=", vec.GetLen())
	fmt.Println("capacity=", vec.GetCapacity())
}

func TestNewVecFromBytes(t *testing.T) {
	vec := loro.NewRustVecFromBytes([]byte("Hello, World!"))
	if !bytes.Equal([]byte("Hello, World!"), vec.Bytes()) {
		t.Fail()
	}
}

func TestLoroDocImport(t *testing.T) {
	doc1 := loro.NewLoroDoc()
	text := doc1.GetText("test")
	text.UpdateText("Hello, World!")
	snapshot := doc1.ExportSnapshot()

	doc2 := loro.NewLoroDoc()
	doc2.Import(snapshot.Bytes())
	text2 := doc2.GetText("test")
	if text2.ToString() != "Hello, World!" {
		t.Fail()
	}
}

func TestGC(t *testing.T) {
	timeStart := time.Now()
	docs := make([]*loro.LoroDoc, 100000)
	for i := 0; i < 100000; i++ {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.UpdateText(fmt.Sprintf("Hello, World! %d", i))
		snapshot := doc.ExportSnapshot()
		doc2 := loro.NewLoroDoc()
		doc2.Import(snapshot.Bytes())
		docs[i] = doc2
	}
	fmt.Println("Successfully inserted 100000 loro docs")
	timeEnd := time.Now()
	fmt.Println("time taken=", timeEnd.Sub(timeStart))
}

func TestCreateAndEditManyLoroDocs(t *testing.T) {
	timeStart := time.Now()
	docs := make([]*loro.LoroDoc, 100000)
	for i := 0; i < 100000; i++ {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.UpdateText(fmt.Sprintf("Hello, World! %d", i))
		_ = doc.ExportSnapshot()
		// doc2 := loro.NewLoroDoc()
		// doc2.Import(snapshot.Bytes())
		// docs[i] = doc2
	}
	_ = docs
	timeEnd := time.Now()
	fmt.Println("time taken=", timeEnd.Sub(timeStart))
}
