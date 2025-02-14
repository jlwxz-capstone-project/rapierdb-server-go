package main

import (
	"fmt"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/stretchr/testify/assert"
)

func Must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

// 基础文档操作测试
// func TestLoroDocBasic(t *testing.T) {
// 	t.Run("创建和销毁文档实例", func(t *testing.T) {
// 		doc := loro.NewLoroDoc()
// 		assert.NotNil(t, doc, "Should create valid LoroDoc instance")
// 	})

// 	t.Run("导出文档快照", func(t *testing.T) {
// 		doc := loro.NewLoroDoc()
// 		text := doc.GetText("test")
// 		text.UpdateText("Hello, World!")
// 		vec := doc.ExportSnapshot()
// 		assert.Greater(t, vec.GetLen(), uint32(0), "Snapshot should have non-zero length")
// 		assert.GreaterOrEqual(t, vec.GetCapacity(), vec.GetLen(), "Capacity should >= length")
// 	})

// 	t.Run("导入文档快照", func(t *testing.T) {
// 		doc1 := loro.NewLoroDoc()
// 		text := doc1.GetText("test")
// 		text.UpdateText("Hello, World!")
// 		snapshot := doc1.ExportSnapshot()

// 		doc2 := loro.NewLoroDoc()
// 		doc2.Import(snapshot.Bytes())
// 		text2 := doc2.GetText("test")
// 		assert.Equal(t, "Hello, World!", text2.ToString(), "Imported text should match")

// 		vv1 := doc1.GetStateVv()
// 		vv2 := doc2.GetStateVv()
// 		assert.True(t, bytes.Equal(vv1.Encode().Bytes(), vv2.Encode().Bytes()), "Version vectors should match")
// 	})

// 	t.Run("文档分叉", func(t *testing.T) {
// 		doc := loro.NewLoroDoc()
// 		doc.GetText("test").UpdateText("Hello, World!")
// 		frontiers := doc.GetOplogFrontiers()
// 		fork1 := doc.Fork()
// 		fork2 := doc.ForkAt(frontiers)

// 		assert.NotNil(t, fork1, "直接分叉的文档应该有效")
// 		assert.NotNil(t, fork2, "在指定边界分叉的文档应该有效")
// 	})
// }

// // 文本操作测试
// func TestLoroText(t *testing.T) {
// 	t.Run("基本文本操作", func(t *testing.T) {
// 		doc := loro.NewLoroDoc()
// 		text := doc.GetText("test")
// 		text.UpdateText("Hello, World!")
// 		assert.Equal(t, "Hello, World!", text.ToString(), "文本内容应该匹配")
// 	})

// 	t.Run("文本指定位置插入", func(t *testing.T) {
// 		doc := loro.NewLoroDoc()
// 		text := doc.GetText("test")
// 		text.InsertText("World!", 0)
// 		text.InsertText("Hello, ", 0)
// 		assert.Equal(t, "Hello, World!", text.ToString(), "插入的文本应该正确拼接")
// 	})

// 	t.Run("UTF-8文本操作", func(t *testing.T) {
// 		doc := loro.NewLoroDoc()
// 		text := doc.GetText("test")
// 		text.InsertText("你好😊", 0)
// 		assert.Equal(t, uint32(3), text.GetLength(), "Unicode码点数应该正确")
// 		assert.Equal(t, uint32(10), text.GetLengthUtf8(), "UTF-8字节数应该正确")

// 		text2 := doc.GetText("test2")
// 		text2.InsertTextUtf8("你好", 0)
// 		text2.InsertTextUtf8("世界", 6)
// 		assert.Equal(t, "你好世界", text2.ToString(), "UTF-8位置插入应该正确")
// 	})
// }

// // 版本控制测试
// func TestLoroVersionControl(t *testing.T) {
// 	t.Run("版本向量和边界基础操作", func(t *testing.T) {
// 		doc := loro.NewLoroDoc()
// 		oplogVv := doc.GetOplogVv()
// 		stateVv := doc.GetStateVv()
// 		oplogFrontiers := doc.GetOplogFrontiers()
// 		stateFrontiers := doc.GetStateFrontiers()

// 		assert.NotNil(t, oplogVv, "操作日志版本向量应该存在")
// 		assert.NotNil(t, stateVv, "状态版本向量应该存在")
// 		assert.NotNil(t, oplogFrontiers, "操作日志边界应该存在")
// 		assert.NotNil(t, stateFrontiers, "状态边界应该存在")
// 	})

// 	t.Run("版本向量和边界编解码", func(t *testing.T) {
// 		doc := loro.NewLoroDoc()
// 		text := doc.GetText("test")
// 		text.UpdateText("Hello, World!")

// 		oplogVv := doc.GetOplogVv()
// 		stateFrontiers := doc.GetStateFrontiers()

// 		oplogVvBytes := oplogVv.Encode()
// 		stateFrontiersBytes := stateFrontiers.Encode()

// 		oplogVvDecoded := loro.NewVvFromBytes(oplogVvBytes)
// 		stateFrontiersDecoded := loro.NewFrontiersFromBytes(stateFrontiersBytes)

// 		assert.True(t, bytes.Equal(oplogVvBytes.Bytes(), oplogVvDecoded.Encode().Bytes()),
// 			"编码/解码后的版本向量应该匹配")
// 		assert.True(t, bytes.Equal(stateFrontiersBytes.Bytes(), stateFrontiersDecoded.Encode().Bytes()),
// 			"编码/解码后的边界应该匹配")
// 	})

// 	t.Run("边界工具函数", func(t *testing.T) {
// 		frontiers := loro.NewEmptyFrontiers()
// 		opId := loro.NewOpId(1, 2)

// 		assert.False(t, frontiers.Contains(opId), "新边界应该为空")
// 		frontiers.Push(opId)
// 		assert.True(t, frontiers.Contains(opId), "应该包含已推入的操作ID")
// 		frontiers.Remove(opId)
// 		assert.False(t, frontiers.Contains(opId), "应该成功移除操作ID")
// 	})

// 	t.Run("增量更新导出", func(t *testing.T) {
// 		doc := loro.NewLoroDoc()
// 		text := doc.GetText("test")
// 		text.UpdateText("v1")
// 		vv1 := doc.GetStateVv()

// 		text.UpdateText("v2")
// 		updates := doc.ExportUpdatesFrom(vv1)
// 		assert.Greater(t, updates.GetLen(), uint32(0), "应该导出增量更新")

// 		doc2 := loro.NewLoroDoc()
// 		doc2.Import(updates.Bytes())
// 		assert.Equal(t, "v2", doc2.GetText("test").ToString())
// 	})

// 	t.Run("版本边界转换", func(t *testing.T) {
// 		doc := loro.NewLoroDoc()
// 		text := doc.GetText("test")
// 		text.UpdateText("test")

// 		frontiers := doc.GetOplogFrontiers()
// 		vv := doc.FrontiersToVv(frontiers)
// 		newFrontiers := doc.VvToFrontiers(vv)

// 		assert.True(t, bytes.Equal(frontiers.Encode().Bytes(), newFrontiers.Encode().Bytes()),
// 			"版本向量和边界转换应该可逆")
// 	})
// }

// // 数据结构测试
// func TestLoroDataStructures(t *testing.T) {
// 	t.Run("RustVec操作", func(t *testing.T) {
// 		data := []byte("Hello, World!")
// 		vec := loro.NewRustBytesVec(data)
// 		assert.True(t, bytes.Equal(data, vec.Bytes()), "Bytes content should match")
// 		assert.Equal(t, uint32(len(data)), vec.GetLen(), "Vector length should match")
// 	})

// 	t.Run("RustPtrVec操作", func(t *testing.T) {
// 		vec := loro.NewRustPtrVec()
// 		doc1 := loro.NewLoroDoc()
// 		doc2 := loro.NewLoroDoc()
// 		vec.Push(doc1.Ptr)
// 		vec.Push(doc2.Ptr)
// 		assert.Equal(t, uint32(2), vec.GetLen(), "PtrVec length should be 2")
// 	})

// 	t.Run("LoroList操作", func(t *testing.T) {
// 		list := loro.NewEmptyLoroList()
// 		list.PushNull()
// 		list.PushBool(true)
// 		list.PushDouble(1.23)
// 		assert.Equal(t, nil, list.GetNull(0))
// 		b, err := list.GetBool(1)
// 		assert.NoError(t, err)
// 		assert.Equal(t, true, b)
// 		d, err := list.GetDouble(2)
// 		assert.NoError(t, err)
// 		assert.Equal(t, 1.23, d)
// 	})

// 	t.Run("LoroMap操作", func(t *testing.T) {
// 		doc := loro.NewLoroDoc()
// 		m := doc.GetMap("test")

// 		// 基本类型测试
// 		m.InsertBool("bool", true)
// 		m.InsertDouble("double", 1.23)
// 		m.InsertI64("i64", 123)
// 		m.InsertString("string", "hello")
// 		assert.Equal(t, true, m.GetBool("bool"))
// 		assert.Equal(t, 1.23, m.GetDouble("double"))
// 		assert.Equal(t, int64(123), m.GetI64("i64"))
// 		assert.Equal(t, "hello", m.GetString("string"))

// 		// 文本类型测试
// 		text := doc.GetText("text")
// 		text.UpdateText("Hello World")
// 		text2 := m.InsertText("text", text)
// 		text3 := m.GetText("text")
// 		assert.Equal(t, "Hello World", text3.ToString())
// 		text3.UpdateText("Hello World, Again")
// 		assert.Equal(t, "Hello World", text.ToString())
// 		assert.Equal(t, "Hello World, Again", text2.ToString())
// 		assert.Equal(t, "Hello World, Again", text3.ToString())
// 	})

// 	t.Run("LoroMovableList 操作", func(t *testing.T) {
// 		list := loro.NewEmptyLoroMovableList()

// 		// 测试基本类型插入
// 		list.PushNull()
// 		list.PushBool(true)
// 		list.PushDouble(3.14)
// 		list.PushI64(123)
// 		list.PushString("hello")

// 		// 测试容器类型插入
// 		subList := list.PushList(loro.NewEmptyLoroList())
// 		subList.PushString("sublist")

// 		// 验证结果
// 		assert.Equal(t, uint32(5), list.GetLen())
// 		s, err := list.GetString(4)
// 		assert.NoError(t, err)
// 		assert.Equal(t, "hello", s)

// 		// 测试移动操作
// 		list.PushMovableList(loro.NewEmptyLoroMovableList())
// 	})

// 	t.Run("List错误处理", func(t *testing.T) {
// 		list := loro.NewEmptyLoroList()

// 		// 测试越界访问
// 		_, err := list.GetBool(0)
// 		assert.ErrorIs(t, err, loro.ErrFailedToGetBool)

// 		// 测试类型不匹配
// 		list.PushBool(true)
// 		_, err = list.GetString(0)
// 		assert.ErrorIs(t, err, loro.ErrFailedToGetString)
// 	})
// }

// // 性能测试
// func TestLoroPerformance(t *testing.T) {
// 	t.Run("大量文档操作性能", func(t *testing.T) {
// 		timeStart := time.Now()
// 		docs := make([]*loro.LoroDoc, 100000)
// 		for i := 0; i < 100000; i++ {
// 			doc := loro.NewLoroDoc()
// 			text := doc.GetText("test")
// 			text.UpdateText(fmt.Sprintf("Hello, World! %d", i))
// 			snapshot := doc.ExportSnapshot()
// 			doc2 := loro.NewLoroDoc()
// 			doc2.Import(snapshot.Bytes())
// 			docs[i] = doc2
// 		}
// 		timeEnd := time.Now()
// 		fmt.Printf("插入100000个文档耗时: %v\n", timeEnd.Sub(timeStart))
// 	})

// 	t.Run("大量Diff操作性能", func(t *testing.T) {
// 		timeStart := time.Now()
// 		cids := make([]*loro.ContainerId, 0)
// 		diffEvents := make([]*loro.DiffEvent, 0)
// 		for i := 0; i < 100000; i++ {
// 			doc := loro.NewLoroDoc()
// 			f1 := doc.GetOplogFrontiers()
// 			doc.GetText("test").UpdateText("Hello, World!")
// 			f2 := doc.GetOplogFrontiers()
// 			diff := doc.Diff(f1, f2)
// 			events := diff.GetEvents()

// 			for _, event := range events {
// 				cid := event.ContainerId
// 				diffEvent := event.DiffEvent
// 				cids = append(cids, &cid)
// 				diffEvents = append(diffEvents, &diffEvent)
// 			}
// 		}
// 		timeEnd := time.Now()
// 		fmt.Printf("执行100000次Diff操作耗时: %v\n", timeEnd.Sub(timeStart))
// 	})
// }

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
		assert.ErrorIs(t, err, loro.ErrGetLoroValue)

		// 空值访问
		nullVal := loro.NewLoroValueNull()
		_, err = nullVal.GetBool()
		assert.ErrorIs(t, err, loro.ErrGetLoroValue)
	})
}

func TestLoroListDiff(t *testing.T) {
	doc := loro.NewLoroDoc()
	f1 := doc.GetOplogFrontiers()
	doc.GetText("text").UpdateText("Hello, World!")
	f2 := doc.GetOplogFrontiers()
	diff := doc.Diff(f1, f2)
	events := diff.GetEvents()
	for _, event := range events {
		fmt.Println(event.DiffEvent.GetType())
	}
}
