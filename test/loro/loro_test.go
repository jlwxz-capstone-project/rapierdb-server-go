package main

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/stretchr/testify/assert"
)

func TestLoroDoc(t *testing.T) {
	t.Run("创建和销毁文档实例", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		assert.NotNil(t, doc, "Should create valid LoroDoc instance")
	})

	t.Run("更新文本内容", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.UpdateText("Hello, World!")
		assert.Equal(t, "Hello, World!", text.ToString(), "Text content should match")
	})

	t.Run("插入文本内容", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.InsertText("Hello, World!", 0)
		assert.Equal(t, "Hello, World!", text.ToString(), "Inserted text should match")
	})

	t.Run("导出文档快照", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.UpdateText("Hello, World!")
		vec := doc.ExportSnapshot()
		assert.Greater(t, vec.GetLen(), uint32(0), "Snapshot should have non-zero length")
		assert.GreaterOrEqual(t, vec.GetCapacity(), vec.GetLen(), "Capacity should >= length")
	})

	t.Run("从字节创建向量", func(t *testing.T) {
		data := []byte("Hello, World!")
		vec := loro.NewRustVecFromBytes(data)
		assert.True(t, bytes.Equal(data, vec.Bytes()), "Bytes content should match")
		assert.Equal(t, uint32(len(data)), vec.GetLen(), "Vector length should match")
	})

	t.Run("导入文档快照", func(t *testing.T) {
		doc1 := loro.NewLoroDoc()
		text := doc1.GetText("test")
		text.UpdateText("Hello, World!")
		snapshot := doc1.ExportSnapshot()

		doc2 := loro.NewLoroDoc()
		doc2.Import(snapshot.Bytes())
		text2 := doc2.GetText("test")
		assert.Equal(t, "Hello, World!", text2.ToString(), "Imported text should match")

		vv1 := doc1.GetStateVv()
		vv2 := doc2.GetStateVv()
		assert.True(t, bytes.Equal(vv1.Encode().Bytes(), vv2.Encode().Bytes()), "Version vectors should match")
	})

	t.Run("测试插入和编辑大量文档", func(t *testing.T) {
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
	})

	t.Run("文本基本操作", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.UpdateText("Hello, World!")
		assert.Equal(t, "Hello, World!", text.ToString(), "文本内容应该匹配")
	})

	t.Run("文本指定位置插入", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.InsertText("World!", 0)
		text.InsertText("Hello, ", 0)
		assert.Equal(t, "Hello, World!", text.ToString(), "插入的文本应该正确拼接")
	})

	t.Run("UTF-8文本长度计算", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.InsertText("你好😊", 0)
		// Unicode码点数应该是3（2个中文 + 1个emoji）
		// UTF-8编码字节数应该是2*3 + 4 = 10字节
		assert.Equal(t, uint32(3), text.GetLength(), "Unicode码点数应该正确")
		assert.Equal(t, uint32(10), text.GetLengthUtf8(), "UTF-8字节数应该正确")
	})

	t.Run("UTF-8位置插入", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.InsertTextUtf8("你好", 0)
		text.InsertTextUtf8("世界", 6) // 因为"你好"占6个字节
		assert.Equal(t, "你好世界", text.ToString(), "UTF-8位置插入应该正确")
	})

	t.Run("获取版本向量和边界", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		oplogVv := doc.GetOplogVv()
		stateVv := doc.GetStateVv()
		oplogFrontiers := doc.GetOplogFrontiers()
		stateFrontiers := doc.GetStateFrontiers()

		assert.NotNil(t, oplogVv, "操作日志版本向量应该存在")
		assert.NotNil(t, stateVv, "状态版本向量应该存在")
		assert.NotNil(t, oplogFrontiers, "操作日志边界应该存在")
		assert.NotNil(t, stateFrontiers, "状态边界应该存在")
	})

	t.Run("版本向量和边界的编码解码", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.UpdateText("Hello, World!")

		oplogVv := doc.GetOplogVv()
		stateFrontiers := doc.GetStateFrontiers()

		oplogVvBytes := oplogVv.Encode()
		stateFrontiersBytes := stateFrontiers.Encode()

		oplogVvDecoded := loro.NewVvFromBytes(oplogVvBytes)
		stateFrontiersDecoded := loro.NewFrontiersFromBytes(stateFrontiersBytes)

		assert.True(t, bytes.Equal(oplogVvBytes.Bytes(), oplogVvDecoded.Encode().Bytes()),
			"编码/解码后的版本向量应该匹配")
		assert.True(t, bytes.Equal(stateFrontiersBytes.Bytes(), stateFrontiersDecoded.Encode().Bytes()),
			"编码/解码后的边界应该匹配")
	})

	t.Run("导出更新", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.UpdateText("Hello, World!")
		allUpdates := doc.ExportAllUpdates()
		assert.Greater(t, allUpdates.GetLen(), uint32(0), "应该有可导出的更新")
	})

	t.Run("版本向量和边界的相互转换", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		text := doc.GetText("test")
		text.UpdateText("Hello, World!")
		vv := doc.GetOplogVv()
		frontiers := doc.GetOplogFrontiers()
		vvFromFrontiers := doc.FrontiersToVv(frontiers)
		frontiersFromVv := doc.VvToFrontiers(vv)

		assert.True(t, bytes.Equal(vv.Encode().Bytes(), vvFromFrontiers.Encode().Bytes()),
			"从边界转换的版本向量应该与原版本向量匹配")
		assert.True(t, bytes.Equal(frontiers.Encode().Bytes(), frontiersFromVv.Encode().Bytes()),
			"从版本向量转换的边界应该与原边界匹配")
	})

	t.Run("创建操作ID", func(t *testing.T) {
		opId := loro.NewOpId(1, 2)
		assert.NotNil(t, opId, "应该成功创建操作ID")
	})

	t.Run("边界工具函数", func(t *testing.T) {
		frontiers := loro.NewEmptyFrontiers()
		opId := loro.NewOpId(1, 2)

		assert.False(t, frontiers.Contains(opId), "新边界应该为空")

		frontiers.Push(opId)
		assert.True(t, frontiers.Contains(opId), "应该包含已推入的操作ID")

		frontiers.Remove(opId)
		assert.False(t, frontiers.Contains(opId), "应该成功移除操作ID")
	})

	t.Run("文档分叉", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		doc.GetText("test").UpdateText("Hello, World!")
		frontiers := doc.GetOplogFrontiers()
		fork1 := doc.Fork()
		fork2 := doc.ForkAt(frontiers)

		assert.NotNil(t, fork1, "直接分叉的文档应该有效")
		assert.NotNil(t, fork2, "在指定边界分叉的文档应该有效")
	})
}
