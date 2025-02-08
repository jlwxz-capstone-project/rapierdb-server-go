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

func TestInsertLoroText(t *testing.T) {
	doc := loro.NewLoroDoc()
	text := doc.GetText("test")
	text.InsertText("Hello, World!", 0)
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

func TestInsertAndEditManyLoroDocs(t *testing.T) {
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

func TestGetVvAndFrontiers(t *testing.T) {
	doc := loro.NewLoroDoc()
	oplogVv := doc.GetOplogVv()
	stateVv := doc.GetStateVv()
	oplogFrontiers := doc.GetOplogFrontiers()
	stateFrontiers := doc.GetStateFrontiers()
	fmt.Println("oplog vv=", oplogVv)
	fmt.Println("state vv=", stateVv)
	fmt.Println("oplog frontiers=", oplogFrontiers)
	fmt.Println("state frontiers=", stateFrontiers)
}

func TestEncodeAndDecodeVvAndFrontiers(t *testing.T) {
	doc := loro.NewLoroDoc()
	text := doc.GetText("test")
	text.UpdateText("Hello, World!")
	oplogVv := doc.GetOplogVv()
	stateVv := doc.GetOplogFrontiers()
	oplogVvBytes := oplogVv.Encode()
	stateVvBytes := stateVv.Encode()
	oplogVvDecoded := loro.NewVvFromBytes(oplogVvBytes)
	stateVvDecoded := loro.NewVvFromBytes(stateVvBytes)
	oplogVvDecodedBytes := oplogVvDecoded.Encode()
	stateVvDecodedBytes := stateVvDecoded.Encode()
	if !bytes.Equal(oplogVvBytes.Bytes(), oplogVvDecodedBytes.Bytes()) {
		t.Fail()
	}
	if !bytes.Equal(stateVvBytes.Bytes(), stateVvDecodedBytes.Bytes()) {
		t.Fail()
	}
}

func TestExportUpdates(t *testing.T) {
	doc := loro.NewLoroDoc()
	text := doc.GetText("test")
	text.UpdateText("Hello, World!")
	allUpdates := doc.ExportAllUpdates()
	fmt.Println("all updates=", allUpdates.GetLen())
}

func TestVvFrontiersConversion(t *testing.T) {
	doc := loro.NewLoroDoc()
	text := doc.GetText("test")
	text.UpdateText("Hello, World!")
	vv := doc.GetOplogVv()
	frontiers := doc.GetOplogFrontiers()
	vvFromFrontiers := doc.FrontiersToVv(frontiers)
	frontiersFromVv := doc.VvToFrontiers(vv)
	if !bytes.Equal(vv.Encode().Bytes(), vvFromFrontiers.Encode().Bytes()) {
		t.Fail()
	}
	if !bytes.Equal(frontiers.Encode().Bytes(), frontiersFromVv.Encode().Bytes()) {
		t.Fail()
	}
}

func TestCreateOpId(t *testing.T) {
	opId := loro.NewOpId(1, 2)
	fmt.Println("opId=", opId)
}

func TestFrontiersUtil(t *testing.T) {
	frontiers := loro.NewEmptyFrontiers()
	opId := loro.NewOpId(1, 2)
	contains := frontiers.Contains(opId)
	if contains {
		t.Fail()
	}
	frontiers.Push(opId)
	contains = frontiers.Contains(opId)
	if !contains {
		t.Fail()
	}
	frontiers.Remove(opId)
	contains = frontiers.Contains(opId)
	if contains {
		t.Fail()
	}
}
