package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
	"github.com/stretchr/testify/assert"
)

// 基础文档操作测试
func TestLoroDocBasic(t *testing.T) {
	t.Run("导出和导入文档快照", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.UpdateText("Hello, World!")
		snapshot := doc.ExportSnapshot()
		allUpdates := doc.ExportAllUpdates()
		doc2 := loro.NewLoroDoc()
		doc2.Import(snapshot.Bytes())
		doc3 := loro.NewLoroDoc()
		doc3.Import(allUpdates.Bytes())
		assert.Equal(t, util.Must(doc2.GetText("test").ToString()), "Hello, World!")
		assert.Equal(t, util.Must(doc3.GetText("test").ToString()), "Hello, World!")
	})

	t.Run("Fork 文档", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.UpdateText("Hello World!")
		f := doc.GetOplogFrontiers()
		text.InsertText(" Again!", text.GetLength())
		fork1 := doc.Fork()
		fork2 := doc.ForkAt(f)

		assert.Equal(t, util.Must(fork1.GetText("test").ToString()), "Hello World! Again!")
		assert.Equal(t, util.Must(fork2.GetText("test").ToString()), "Hello World!")

		// 对 fork 的修改应该不影响原文档
		fork1.GetText("test").InsertText("xxx", 0)
		assert.Equal(t, util.Must(doc.GetText("test").ToString()), "Hello World! Again!")
	})
}

// 文本操作测试
func TestLoroText(t *testing.T) {
	t.Run("更新 LoroText", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.UpdateText("Hello, World!")
		assert.Equal(t, "Hello, World!", util.Must(text.ToString()), "文本内容应该匹配")
		text.UpdateText("Hello, World! Again!")
		assert.Equal(t, "Hello, World! Again!", util.Must(text.ToString()), "文本内容应该匹配")
	})

	t.Run("文本指定位置插入", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.InsertText("World!", 0)
		text.InsertText("Hello, ", 0)
		assert.Equal(t, "Hello, World!", util.Must(text.ToString()))
	})

	t.Run("UTF-8文本操作", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.InsertText("你好😊", 0)
		assert.Equal(t, uint32(3), text.GetLength())
		assert.Equal(t, uint32(10), text.GetLengthUtf8())

		text2 := doc.GetText("test2")
		text2.InsertTextUtf8("你好", 0)
		text2.InsertTextUtf8("世界", 6)
		assert.Equal(t, "你好世界", util.Must(text2.ToString()))
	})
}

// 性能测试
func TestLoroPerformance(t *testing.T) {
	t.Run("大量文档操作性能", func(t *testing.T) {
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
		timeEnd := time.Now()
		fmt.Printf("插入100000个文档耗时: %v\n", timeEnd.Sub(timeStart))
	})

	t.Run("大量Diff操作性能", func(t *testing.T) {
		timeStart := time.Now()
		cids := make([]*loro.ContainerId, 0)
		diffEvents := make([]*loro.DiffEvent, 0)
		for i := 0; i < 100000; i++ {
			doc := loro.NewLoroDoc()
			f1 := doc.GetOplogFrontiers()
			doc.GetText("test").UpdateText("Hello, World!")
			f2 := doc.GetOplogFrontiers()
			diff := doc.Diff(f1, f2)
			events := diff.GetEvents()

			for _, event := range events {
				cid := event.ContainerId
				diffEvent := event.DiffEvent
				cids = append(cids, &cid)
				diffEvents = append(diffEvents, &diffEvent)
			}
		}
		timeEnd := time.Now()
		fmt.Printf("执行100000次Diff操作耗时: %v\n", timeEnd.Sub(timeStart))
	})
}

func TestLoroValue(t *testing.T) {
	t.Run("基本类型值操作", func(t *testing.T) {
		// 测试空值
		nullVal := loro.NewLoroValueNull()
		assert.Equal(t, loro.LORO_NULL_VALUE, nullVal.GetType())

		// 测试布尔值
		boolVal := loro.NewLoroValueBool(true)
		b, err := boolVal.GetBool()
		assert.NoError(t, err)
		assert.True(t, b)

		// 测试浮点数
		doubleVal := loro.NewLoroValueDouble(3.14)
		d, err := doubleVal.GetDouble()
		assert.NoError(t, err)
		assert.InDelta(t, 3.14, d, 0.001)

		// 测试整型
		i64Val := loro.NewLoroValueI64(123)
		i, err := i64Val.GetI64()
		assert.NoError(t, err)
		assert.Equal(t, int64(123), i)

		// 测试字符串
		strVal := loro.NewLoroValueString("hello")
		s, err := strVal.GetString()
		assert.NoError(t, err)
		assert.Equal(t, "hello", s)

		// 测试二进制
		binVal := loro.NewLoroValueBinary([]byte{0x01, 0x02})
		bin, err := binVal.GetBinary()
		assert.NoError(t, err)
		assert.Equal(t, []byte{0x01, 0x02}, bin.Bytes())
	})

	t.Run("list 和 map 操作", func(t *testing.T) {
		// 测试 list
		list := []*loro.LoroValue{
			loro.NewLoroValueBool(true),
			loro.NewLoroValueDouble(1.23),
			loro.NewLoroValueString("nested"),
		}
		listVal := loro.NewLoroValueList(list)
		l, err := listVal.GetList()
		assert.NoError(t, err)
		assert.Equal(t, 3, len(l))

		// 测试 map
		m := map[string]*loro.LoroValue{
			"key1": loro.NewLoroValueI64(123),
			"key2": loro.NewLoroValueString("value"),
		}
		mapVal := loro.NewLoroValueMap(m)
		mp, err := mapVal.GetMap()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(mp))
	})

	t.Run("JSON 转换", func(t *testing.T) {
		// 测试JSON导入
		jsonStr := `{"a":1,"b":"str","c":[true]}`
		jsonVal, err := loro.NewLoroValueFromJson(jsonStr)
		assert.NoError(t, err)

		// 测试JSON导出
		result, err := jsonVal.ToJson()
		assert.NoError(t, err)
		assert.JSONEq(t, jsonStr, result)
	})

	t.Run("错误处理", func(t *testing.T) {
		// 类型不匹配
		boolVal := loro.NewLoroValueBool(true)
		_, err := boolVal.GetString()
		assert.ErrorIs(t, err, loro.ErrLoroGetFailed)

		// 空值访问
		nullVal := loro.NewLoroValueNull()
		_, err = nullVal.GetBool()
		assert.ErrorIs(t, err, loro.ErrLoroGetFailed)
	})
}

func TestLoroInspectImport(t *testing.T) {
	doc := loro.NewLoroDoc()
	text := doc.GetText("test")
	text.UpdateText("Hello, World!")
	snapshot := doc.ExportSnapshot()
	meta := util.Must(loro.InspectImport(snapshot, true))
	fmt.Printf("meta: %v\n", meta)
}
