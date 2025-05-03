package main

import (
	"fmt"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/stretchr/testify/assert"
)

func TestLoroMap(t *testing.T) {
	m := loro.NewEmptyLoroMap()

	assert.NoError(t, m.InsertValueCoerce("key1", nil))
	assert.NoError(t, m.InsertValueCoerce("key2", true))
	assert.NoError(t, m.InsertValueCoerce("key3", 3.0))
	assert.NoError(t, m.InsertValueCoerce("key4", 1))
	assert.NoError(t, m.InsertValueCoerce("key5", "hello"))
	assert.NoError(t, m.InsertValueCoerce("key6", map[string]string{"key": "value"}))
	assert.NoError(t, m.InsertValueCoerce("key7", []string{"a", "b", "c"}))
	assert.NoError(t, m.InsertValueCoerce("key8", []byte("hello")))

	assert.True(t, m.Contains("key1"))
	assert.True(t, m.Contains("key2"))
	assert.True(t, m.Contains("key3"))
	assert.True(t, m.Contains("key4"))
	assert.True(t, m.Contains("key5"))
	assert.True(t, m.Contains("key6"))
	assert.True(t, m.Contains("key7"))
	assert.True(t, m.Contains("key8"))
	assert.Equal(t, nil, m.MustGet("key1"))
	assert.Equal(t, true, m.MustGet("key2"))
	assert.Equal(t, float64(3.0), m.MustGet("key3"))
	assert.Equal(t, int64(1), m.MustGet("key4"))
	assert.Equal(t, "hello", m.MustGet("key5"))
	assert.Equal(t, map[string]loro.LoroValue{"key": "value"}, m.MustGet("key6"))
	assert.Equal(t, []loro.LoroValue{"a", "b", "c"}, m.MustGet("key7"))
	assert.Equal(t, []byte("hello"), m.MustGet("key8"))

	assert.Equal(t, uint32(8), m.GetLen())

	m2 := loro.NewEmptyLoroMap()
	assert.NoError(t, m2.InsertValueCoerce("name", "Chris"))
	assert.NoError(t, m2.InsertValueCoerce("age", 25))
}

func TestLoroList(t *testing.T) {
	l := loro.NewEmptyLoroList()

	assert.NoError(t, l.InsertValueCoerce(0, nil))
	assert.NoError(t, l.InsertValueCoerce(1, true))
	assert.NoError(t, l.InsertValueCoerce(2, 3.0))
	assert.NoError(t, l.InsertValueCoerce(3, 1))
	assert.NoError(t, l.InsertValueCoerce(4, "hello"))
	assert.NoError(t, l.InsertValueCoerce(5, map[string]string{"key": "value"}))
	assert.NoError(t, l.InsertValueCoerce(6, []string{"a", "b", "c"}))
	assert.NoError(t, l.InsertValueCoerce(7, []byte("hello")))

	assert.Equal(t, nil, l.MustGet(0))
	assert.Equal(t, true, l.MustGet(1))
	assert.Equal(t, float64(3.0), l.MustGet(2))
	assert.Equal(t, int64(1), l.MustGet(3))
	assert.Equal(t, "hello", l.MustGet(4))
	assert.Equal(t, map[string]loro.LoroValue{"key": "value"}, l.MustGet(5))
	assert.Equal(t, []loro.LoroValue{"a", "b", "c"}, l.MustGet(6))
	assert.Equal(t, []byte("hello"), l.MustGet(7))

	assert.Equal(t, uint32(8), l.GetLen())
}

func TestVvCompare(t *testing.T) {
	doc1 := loro.NewLoroDoc()
	dataMap1 := doc1.GetMap("data")
	dataMap1.InsertValueCoerce("age", 30)

	doc2 := doc1.Fork()
	dataMap2 := doc2.GetMap("data")
	dataMap2.InsertValueCoerce("age", 31)
	vv2 := doc2.GetOplogVv()

	doc3 := doc1.Fork()
	dataMap3 := doc3.GetMap("data")
	dataMap3.InsertValueCoerce("age", 32)
	vv3 := doc3.GetOplogVv()

	fmt.Println("doc1.age", dataMap1.MustGet("age"))
	fmt.Println("doc2.age", dataMap2.MustGet("age"))
	fmt.Println("doc3.age", dataMap3.MustGet("age"))
	fmt.Println(vv2.PartialCompare(vv3))
}

func TestVersionGap(t *testing.T) {
	doc1 := loro.NewLoroDoc()
	dataMap1 := doc1.GetMap("data")
	dataMap1.InsertValueCoerce("age", 30)
	vv1 := doc1.GetOplogVv()

	doc2 := doc1.Fork()
	dataMap2 := doc2.GetMap("data")
	dataMap2.InsertValueCoerce("age", 21)
	vv21 := doc2.GetOplogVv()
	dataMap2.InsertValueCoerce("name", "John")
	updateFromVv21 := doc2.ExportUpdatesFrom(vv21).Bytes()
	updateFromVv1 := doc2.ExportUpdatesFrom(vv1).Bytes()

	doc3 := doc1.Fork()
	status := doc3.Import(updateFromVv21)
	assert.True(t, status.GetSuccess().IsEmpty())
	assert.NotNil(t, status.GetPending())

	status2 := doc3.Import(updateFromVv1)
	assert.False(t, status2.GetSuccess().IsEmpty())
	assert.Nil(t, status2.GetPending())
}
