package loro

/*
#cgo LDFLAGS: -L./loro-c-ffi/target/release -lloro_c_ffi
#include <stdlib.h>
#include "loro-c-ffi/loro_c_ffi.h"
*/
import "C"
import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"runtime"
	"unsafe"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
	pe "github.com/pkg/errors"
)

const (
	LORO_VALUE_NULL      = 0
	LORO_VALUE_BOOL      = 1
	LORO_VALUE_DOUBLE    = 2
	LORO_VALUE_I64       = 3
	LORO_VALUE_BINARY    = 4
	LORO_VALUE_STRING    = 5
	LORO_VALUE_LIST      = 6
	LORO_VALUE_MAP       = 7
	LORO_VALUE_CONTAINER = 8
)

var (
	ErrLoroGetFailed       = errors.New("loro get failed")
	ErrLoroInsertFailed    = pe.New("loro insert failed")
	ErrLoroEncodeFailed    = errors.New("loro encode failed")
	ErrLoroDecodeFailed    = errors.New("loro decode failed")
	ErrInspectImportFailed = errors.New("inspect import failed")
)

const DATA_MAP_NAME = "root"

// ----------- Rust Bytes Vec -----------

type RustBytesVec struct {
	ptr unsafe.Pointer
}

func NewRustBytesVec(data []byte) *RustBytesVec {
	dataPtr := unsafe.Pointer(&data[0])
	dataLen := len(data)
	var newDataPtr *C.uint8_t
	ptr := C.new_vec_from_bytes(dataPtr, C.uint32_t(dataLen), C.uint32_t(dataLen), &newDataPtr)

	ret := &RustBytesVec{
		ptr: ptr,
	}
	runtime.SetFinalizer(ret, func(vec *RustBytesVec) {
		vec.Destroy()
	})
	return ret
}

func (vec *RustBytesVec) Destroy() {
	C.destroy_bytes_vec(vec.ptr)
}

// 获取字节向量的长度，即字节数
func (vec *RustBytesVec) GetLen() uint32 {
	return uint32(C.get_vec_len(vec.ptr))
}

// 获取字节向量的容量，即字节数组的最大长度
func (vec *RustBytesVec) GetCapacity() uint32 {
	return uint32(C.get_vec_cap(vec.ptr))
}

// 转换为 Go 字节切片，注意：返回的切片是原始字节数组的视图，
// 可以修改原始字节数组的内容，但不要修改切片的长度！
func (vec *RustBytesVec) Bytes() []byte {
	len := vec.GetLen()
	dataPtr := C.get_vec_data(vec.ptr)
	return unsafe.Slice((*byte)(dataPtr), len)
}

// ----------- Rust Ptr Vec ----------

// RustPtrVec 内部封装的是 Rust 中 Vec<*mut u8>
// 它是一个用于存储指针的向量，可以存储任意类型的指针，
// RustPtrVec 常作为 Go 和 Rust 之间复杂数据的中间表示。
// 比如：HashMap<String, LoroValue> <=> RustPtrVec <=> map[string]*LoroValue
type RustPtrVec struct {
	ptr unsafe.Pointer
}

func NewRustPtrVec() *RustPtrVec {
	ptr := C.new_ptr_vec()
	vec := &RustPtrVec{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(vec, func(vec *RustPtrVec) {
		vec.Destroy()
	})
	return vec
}

func (vec *RustPtrVec) Destroy() {
	C.destroy_ptr_vec(vec.ptr)
}

func (vec *RustPtrVec) Get(index uint32) unsafe.Pointer {
	return C.ptr_vec_get(vec.ptr, C.uint32_t(index))
}

func (vec *RustPtrVec) Push(value unsafe.Pointer) {
	C.ptr_vec_push(vec.ptr, value)
}

func (vec *RustPtrVec) GetLen() uint32 {
	return uint32(C.get_ptr_vec_len(vec.ptr))
}

func (vec *RustPtrVec) GetCapacity() uint32 {
	return uint32(C.get_ptr_vec_cap(vec.ptr))
}

func (vec *RustPtrVec) GetData() []unsafe.Pointer {
	len := vec.GetLen()
	dataPtr := C.get_ptr_vec_data(vec.ptr)
	return unsafe.Slice((*unsafe.Pointer)(dataPtr), len)
}

// ----------- Loro Doc -----------

type LoroDoc struct {
	Ptr unsafe.Pointer
}

func (doc *LoroDoc) Destroy() {
	C.destroy_loro_doc(doc.Ptr)
}

func NewLoroDoc() *LoroDoc {
	ret := C.create_loro_doc()
	loroDoc := &LoroDoc{
		Ptr: unsafe.Pointer(ret),
	}
	runtime.SetFinalizer(loroDoc, func(doc *LoroDoc) {
		doc.Destroy()
	})
	return loroDoc
}

func (doc *LoroDoc) GetDataMap() *LoroMap {
	return doc.GetMap(DATA_MAP_NAME)
}

// 获取指定 ID 的文本容器。如果指定 ID 的文本容器不存在，
// 则创建一个并 attach 到当前文档
func (doc *LoroDoc) GetText(id string) *LoroText {
	idPtr := C.CString(id)
	defer C.free(unsafe.Pointer(idPtr))
	ret := C.get_text(doc.Ptr, idPtr)
	loroText := &LoroText{
		ptr: unsafe.Pointer(ret),
	}
	runtime.SetFinalizer(loroText, func(text *LoroText) {
		text.Destroy()
	})
	return loroText
}

// 获取指定 ID 的列表容器。如果指定 ID 的列表容器不存在，
// 则创建一个并 attach 到当前文档
func (doc *LoroDoc) GetList(id string) *LoroList {
	idPtr := C.CString(id)
	defer C.free(unsafe.Pointer(idPtr))
	ret := C.get_list(doc.Ptr, idPtr)
	loroList := &LoroList{
		ptr: unsafe.Pointer(ret),
	}
	runtime.SetFinalizer(loroList, func(list *LoroList) {
		list.Destroy()
	})
	return loroList
}

// 获取指定 ID 的可移动列表容器。如果指定 ID 的可移动列表容器不存在，
// 则创建一个并 attach 到当前文档
func (doc *LoroDoc) GetMovableList(id string) *LoroMovableList {
	idPtr := C.CString(id)
	defer C.free(unsafe.Pointer(idPtr))
	ret := C.get_movable_list(doc.Ptr, idPtr)
	loroMovableList := &LoroMovableList{
		ptr: unsafe.Pointer(ret),
	}
	runtime.SetFinalizer(loroMovableList, func(movableList *LoroMovableList) {
		movableList.Destroy()
	})
	return loroMovableList
}

// 获取指定 ID 的映射容器。如果指定 ID 的映射容器不存在，
// 则创建一个并 attach 到当前文档
func (doc *LoroDoc) GetMap(id string) *LoroMap {
	idPtr := C.CString(id)
	defer C.free(unsafe.Pointer(idPtr))
	ret := C.get_map(doc.Ptr, idPtr)
	loroMap := &LoroMap{
		ptr: unsafe.Pointer(ret),
	}
	runtime.SetFinalizer(loroMap, func(m *LoroMap) {
		m.Destroy()
	})
	return loroMap
}

// 以快照的形式导出。快照包含文档的完整历史
func (doc *LoroDoc) ExportSnapshot() *RustBytesVec {
	ptr := C.export_loro_doc_snapshot(doc.Ptr)
	bytesVec := &RustBytesVec{
		ptr: ptr,
	}
	runtime.SetFinalizer(bytesVec, func(vec *RustBytesVec) {
		vec.Destroy()
	})
	return bytesVec
}

// 导出文档的所有更新。导入所有更新和导入快照都可以恢复文档历史，但快照更节省空间。
// 导出更新的优势是可以通过 ExportUpdatesFrom 和 ExportUpdatesTill 等函数选择性地
// 导出部分更新（例如从特定版本开始），无需导出全部历史，这对增量同步场景很有帮助。
func (doc *LoroDoc) ExportAllUpdates() *RustBytesVec {
	ptr := C.export_loro_doc_all_updates(doc.Ptr)
	bytesVec := &RustBytesVec{
		ptr: ptr,
	}
	runtime.SetFinalizer(bytesVec, func(vec *RustBytesVec) {
		vec.Destroy()
	})
	return bytesVec
}

// 导出从指定版本开始，到最新版本的更新。
func (doc *LoroDoc) ExportUpdatesFrom(from *VersionVector) *RustBytesVec {
	ptr := C.export_loro_doc_updates_from(doc.Ptr, from.ptr)
	bytesVec := &RustBytesVec{
		ptr: ptr,
	}
	runtime.SetFinalizer(bytesVec, func(vec *RustBytesVec) {
		vec.Destroy()
	})
	return bytesVec
}

// 导出从初始状态开始，到指定版本的更新。
func (doc *LoroDoc) ExportUpdatesTill(till *VersionVector) *RustBytesVec {
	ptr := C.export_loro_doc_updates_till(doc.Ptr, till.ptr)
	bytesVec := &RustBytesVec{
		ptr: ptr,
	}
	runtime.SetFinalizer(bytesVec, func(vec *RustBytesVec) {
		vec.Destroy()
	})
	return bytesVec
}

// 导入快照 / 更新（都是一堆字节）
func (doc *LoroDoc) Import(data []byte) {
	snapshot := NewRustBytesVec(data)
	C.loro_doc_import(doc.Ptr, snapshot.ptr)
}

func (doc *LoroDoc) GetOplogVv() *VersionVector {
	ptr := C.get_oplog_vv(doc.Ptr)
	vv := &VersionVector{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(vv, func(vv *VersionVector) {
		vv.Destroy()
	})
	return vv
}

func (doc *LoroDoc) GetStateVv() *VersionVector {
	ptr := C.get_state_vv(doc.Ptr)
	vv := &VersionVector{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(vv, func(vv *VersionVector) {
		vv.Destroy()
	})
	return vv
}

func (doc *LoroDoc) GetOplogFrontiers() *Frontiers {
	ptr := C.get_oplog_frontiers(doc.Ptr)
	frontier := &Frontiers{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(frontier, func(f *Frontiers) {
		f.Destroy()
	})
	return frontier
}

func (doc *LoroDoc) GetStateFrontiers() *Frontiers {
	ptr := C.get_state_frontiers(doc.Ptr)
	frontier := &Frontiers{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(frontier, func(f *Frontiers) {
		f.Destroy()
	})
	return frontier
}

func (doc *LoroDoc) FrontiersToVv(frontiers *Frontiers) *VersionVector {
	ptr := C.frontiers_to_vv(doc.Ptr, frontiers.ptr)
	vv := &VersionVector{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(vv, func(vv *VersionVector) {
		vv.Destroy()
	})
	return vv
}

func (doc *LoroDoc) VvToFrontiers(vv *VersionVector) *Frontiers {
	ptr := C.vv_to_frontiers(doc.Ptr, vv.ptr)
	frontiers := &Frontiers{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(frontiers, func(f *Frontiers) {
		f.Destroy()
	})
	return frontiers
}

func (doc *LoroDoc) Fork() *LoroDoc {
	ptr := C.fork_doc(doc.Ptr)
	loroDoc := &LoroDoc{
		Ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(loroDoc, func(doc *LoroDoc) {
		doc.Destroy()
	})
	return loroDoc
}

func (doc *LoroDoc) ForkAt(frontiers *Frontiers) *LoroDoc {
	ptr := C.fork_doc_at(doc.Ptr, frontiers.ptr)
	loroDoc := &LoroDoc{
		Ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(loroDoc, func(doc *LoroDoc) {
		doc.Destroy()
	})
	return loroDoc
}

func (doc *LoroDoc) Diff(v1, v2 *Frontiers) *DiffBatch {
	ptr := C.diff_loro_doc(doc.Ptr, v1.ptr, v2.ptr)
	diffBatch := &DiffBatch{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(diffBatch, func(d *DiffBatch) {
		d.Destroy()
	})
	return diffBatch
}

// deprecated
// func (doc *LoroDoc) GetByPath(path string) *LoroContainerOrValue {
// 	pathPtr := C.CString(path)
// 	defer C.free(unsafe.Pointer(pathPtr))
// 	ptr := C.loro_doc_get_by_path(doc.Ptr, pathPtr)
// 	if ptr == nil {
// 		return nil
// 	}
// 	containerOrValue := &LoroContainerOrValue{
// 		ptr: unsafe.Pointer(ptr),
// 	}
// 	runtime.SetFinalizer(containerOrValue, func(c *LoroContainerOrValue) {
// 		c.Destroy()
// 	})
// 	return containerOrValue
// }

// ----------- Version Vector -----------

type VersionVector struct {
	ptr unsafe.Pointer
}

func NewEmptyVv() *VersionVector {
	ptr := C.vv_new_empty()
	vv := &VersionVector{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(vv, func(vv *VersionVector) {
		vv.Destroy()
	})
	return vv
}

func NewVvFromBytes(data *RustBytesVec) *VersionVector {
	ptr := C.decode_vv(data.ptr)
	vv := &VersionVector{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(vv, func(vv *VersionVector) {
		vv.Destroy()
	})
	return vv
}

func (vv *VersionVector) Destroy() {
	C.destroy_vv(vv.ptr)
}

func (vv *VersionVector) Encode() *RustBytesVec {
	ptr := C.encode_vv(vv.ptr)
	bytesVec := &RustBytesVec{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(bytesVec, func(vec *RustBytesVec) {
		vec.Destroy()
	})
	return bytesVec
}

// -------------- Frontier -----------

type Frontiers struct {
	ptr unsafe.Pointer
}

func NewEmptyFrontiers() *Frontiers {
	ptr := C.frontiers_new_empty()
	frontiers := &Frontiers{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(frontiers, func(f *Frontiers) {
		f.Destroy()
	})
	return frontiers
}

func NewFrontiersFromBytes(data *RustBytesVec) *Frontiers {
	ptr := C.decode_frontiers(data.ptr)
	frontiers := &Frontiers{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(frontiers, func(f *Frontiers) {
		f.Destroy()
	})
	return frontiers
}

func (f *Frontiers) Destroy() {
	C.destroy_frontiers(f.ptr)
}

func (f *Frontiers) Encode() *RustBytesVec {
	ptr := C.encode_frontiers(f.ptr)
	bytesVec := &RustBytesVec{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(bytesVec, func(vec *RustBytesVec) {
		vec.Destroy()
	})
	return bytesVec
}

func (f *Frontiers) Contains(id *OpId) bool {
	idPtr := unsafe.Pointer(&id.cLayoutId)
	ret := C.frontiers_contains(f.ptr, idPtr)
	return ret != 0
}

func (f *Frontiers) Push(id *OpId) {
	idPtr := unsafe.Pointer(&id.cLayoutId)
	C.frontiers_push(f.ptr, idPtr)
}

func (f *Frontiers) Remove(id *OpId) {
	idPtr := unsafe.Pointer(&id.cLayoutId)
	C.frontiers_remove(f.ptr, idPtr)
}

// ----------- Loro Text -----------

type LoroText struct {
	ptr unsafe.Pointer
}

var (
	ErrFailedToConvertLoroText    = errors.New("failed to convert loro text to string")
	ErrFailedToUpdateLoroText     = errors.New("failed to update loro text")
	ErrFailedToInsertLoroText     = errors.New("failed to insert loro text")
	ErrFailedToInsertLoroTextUtf8 = errors.New("failed to insert loro text utf8")
)

func NewLoroText() *LoroText {
	ptr := C.new_loro_text()
	text := &LoroText{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(text, func(text *LoroText) {
		text.Destroy()
	})
	return text
}

func (text *LoroText) Destroy() {
	// fmt.Println("destroying loro text")
	C.destroy_loro_text(text.ptr)
}

func (text *LoroText) ToContainer() *LoroContainer {
	ptr := C.loro_text_to_container(text.ptr)
	container := &LoroContainer{ptr: ptr}
	runtime.SetFinalizer(container, func(container *LoroContainer) {
		container.Destroy()
	})
	return container
}

func (text *LoroText) ToString() (string, error) {
	var err C.uint8_t
	ret := C.loro_text_to_string(text.ptr, &err)
	if err != 0 {
		return "", pe.WithStack(ErrFailedToConvertLoroText)
	}
	return C.GoString(ret), nil
}

func (text *LoroText) UpdateText(content string) error {
	var err C.uint8_t
	contentPtr := C.CString(content)
	defer C.free(unsafe.Pointer(contentPtr))
	C.update_loro_text(text.ptr, contentPtr, &err)
	if err != 0 {
		return pe.WithStack(ErrFailedToUpdateLoroText)
	}
	return nil
}

func (text *LoroText) InsertText(content string, pos uint32) error {
	var err C.uint8_t
	contentPtr := C.CString(content)
	defer C.free(unsafe.Pointer(contentPtr))
	C.insert_loro_text(text.ptr, C.uint32_t(pos), contentPtr, &err)
	if err != 0 {
		return pe.WithStack(ErrFailedToInsertLoroText)
	}
	return nil
}

func (text *LoroText) InsertTextUtf8(content string, pos uint32) error {
	var err C.uint8_t
	contentPtr := C.CString(content)
	defer C.free(unsafe.Pointer(contentPtr))
	C.insert_loro_text_utf8(text.ptr, C.uint32_t(pos), contentPtr, &err)
	if err != 0 {
		return pe.WithStack(ErrFailedToInsertLoroTextUtf8)
	}
	return nil
}

func (text *LoroText) GetLength() uint32 {
	return uint32(C.loro_text_length(text.ptr))
}

func (text *LoroText) GetLengthUtf8() uint32 {
	return uint32(C.loro_text_length_utf8(text.ptr))
}

func (text *LoroText) IsAttached() bool {
	return C.loro_text_is_attached(text.ptr) != 0
}

// ----------- Loro Map -----------

type LoroMap struct {
	ptr unsafe.Pointer
}

func (m *LoroMap) Destroy() {
	C.destroy_loro_map(m.ptr)
}

func NewEmptyLoroMap() *LoroMap {
	ptr := C.loro_map_new_empty()
	m := &LoroMap{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(m, func(m *LoroMap) {
		m.Destroy()
	})
	return m
}

func (m *LoroMap) ToContainer() *LoroContainer {
	ptr := C.loro_map_to_container(m.ptr)
	container := &LoroContainer{ptr: ptr}
	runtime.SetFinalizer(container, func(container *LoroContainer) {
		container.Destroy()
	})
	return container
}

func (m *LoroMap) ToGoObject() (map[string]any, error) {
	vecPtr := C.loro_map_get_items(m.ptr)
	vec := &RustPtrVec{ptr: unsafe.Pointer(vecPtr)}
	defer vec.Destroy()
	items := vec.GetData()
	result := make(map[string]any, len(items))
	vecLen := vec.GetLen()
	for i := uint32(0); i < vecLen; i += 2 {
		keyPtr := items[i]
		valPtr := items[i+1]
		key := C.GoString((*C.char)(keyPtr))
		val := &LoroContainerOrValue{ptr: valPtr}
		defer val.Destroy()
		valGo, err := val.ToGoObject()
		if valGo == nil {
			return nil, err
		}
		result[key] = valGo
	}
	return result, nil
}

func (m *LoroMap) GetLen() uint32 {
	return uint32(C.loro_map_len(m.ptr))
}

func (m *LoroMap) Get(key string) *LoroContainerOrValue {
	ptr := C.loro_map_get(m.ptr, C.CString(key))
	if ptr == nil {
		return nil
	}
	ret := &LoroContainerOrValue{ptr: unsafe.Pointer(ptr)}
	runtime.SetFinalizer(ret, func(ret *LoroContainerOrValue) {
		ret.Destroy()
	})
	return ret
}

func (m *LoroMap) GetNull(key string) error {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	C.loro_map_get_null(m.ptr, keyPtr, &err)
	if err != 0 {
		return pe.WithStack(fmt.Errorf("%w: get null from map, key=%s", ErrLoroGetFailed, key))
	}
	return nil
}

func (m *LoroMap) GetBool(key string) (bool, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_get_bool(m.ptr, keyPtr, &err)
	if err != 0 {
		return false, pe.WithStack(fmt.Errorf("%w: get bool from map, key=%s", ErrLoroGetFailed, key))
	}
	return ret != 0, nil
}

func (m *LoroMap) GetDouble(key string) (float64, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_get_double(m.ptr, keyPtr, &err)
	if err != 0 {
		return 0, pe.WithStack(fmt.Errorf("%w: get double from map, key=%s", ErrLoroGetFailed, key))
	}
	return float64(ret), nil
}

func (m *LoroMap) GetI64(key string) (int64, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_get_i64(m.ptr, keyPtr, &err)
	if err != 0 {
		return 0, pe.WithStack(fmt.Errorf("%w: get i64 from map, key=%s", ErrLoroGetFailed, key))
	}
	return int64(ret), nil
}

func (m *LoroMap) GetString(key string) (string, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_get_string(m.ptr, keyPtr, &err)
	if err != 0 {
		return "", pe.WithStack(fmt.Errorf("%w: get string from map, key=%s", ErrLoroGetFailed, key))
	}
	return C.GoString(ret), nil
}

func (m *LoroMap) GetText(key string) (*LoroText, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_get_text(m.ptr, keyPtr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: get text from map, key=%s", ErrLoroGetFailed, key))
	}
	text := &LoroText{ptr: unsafe.Pointer(ret)}
	runtime.SetFinalizer(text, func(text *LoroText) {
		text.Destroy()
	})
	return text, nil
}

func (m *LoroMap) GetList(key string) (*LoroList, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_get_list(m.ptr, keyPtr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: get list from map, key=%s", ErrLoroGetFailed, key))
	}
	list := &LoroList{ptr: unsafe.Pointer(ret)}
	runtime.SetFinalizer(list, func(list *LoroList) {
		list.Destroy()
	})
	return list, nil
}

func (m *LoroMap) GetMovableList(key string) (*LoroMovableList, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_get_movable_list(m.ptr, keyPtr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: get movable list from map, key=%s", ErrLoroGetFailed, key))
	}
	list := &LoroMovableList{ptr: unsafe.Pointer(ret)}
	runtime.SetFinalizer(list, func(list *LoroMovableList) {
		list.Destroy()
	})
	return list, nil
}

func (m *LoroMap) GetMap(key string) (*LoroMap, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_get_map(m.ptr, keyPtr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: get map from map, key=%s", ErrLoroGetFailed, key))
	}
	newMap := &LoroMap{ptr: unsafe.Pointer(ret)}
	runtime.SetFinalizer(newMap, func(m *LoroMap) {
		m.Destroy()
	})
	return newMap, nil
}

func (m *LoroMap) InsertValue(key string, value *LoroValue) error {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	C.loro_map_insert_value(m.ptr, keyPtr, value.ptr, &err)
	if err != 0 {
		return pe.WithStack(fmt.Errorf("%w: insert value to map, key=%s", ErrLoroInsertFailed, key))
	}
	return nil
}

func (m *LoroMap) InsertNull(key string) error {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	C.loro_map_insert_null(m.ptr, keyPtr, &err)
	if err != 0 {
		return pe.WithStack(fmt.Errorf("%w: insert null to map, key=%s", ErrLoroInsertFailed, key))
	}
	return nil
}

func (m *LoroMap) InsertBool(key string, value bool) error {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	boolValue := 0
	if value {
		boolValue = 1
	}
	C.loro_map_insert_bool(m.ptr, keyPtr, C.int(boolValue), &err)
	if err != 0 {
		return pe.WithStack(fmt.Errorf("%w: insert bool to map, key=%s", ErrLoroInsertFailed, key))
	}
	return nil
}

func (m *LoroMap) InsertDouble(key string, value float64) error {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	C.loro_map_insert_double(m.ptr, keyPtr, C.double(value), &err)
	if err != 0 {
		return pe.WithStack(fmt.Errorf("%w: insert double to map, key=%s", ErrLoroInsertFailed, key))
	}
	return nil
}

func (m *LoroMap) InsertI64(key string, value int64) error {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	C.loro_map_insert_i64(m.ptr, keyPtr, C.int64_t(value), &err)
	if err != 0 {
		return pe.WithStack(fmt.Errorf("%w: insert i64 to map, key=%s", ErrLoroInsertFailed, key))
	}
	return nil
}

func (m *LoroMap) InsertString(key string, value string) error {
	var err C.uint8_t
	keyPtr := C.CString(key)
	valuePtr := C.CString(value)
	defer C.free(unsafe.Pointer(keyPtr))
	defer C.free(unsafe.Pointer(valuePtr))
	C.loro_map_insert_string(m.ptr, keyPtr, valuePtr, &err)
	if err != 0 {
		return pe.WithStack(fmt.Errorf("%w: insert string to map, key=%s", ErrLoroInsertFailed, key))
	}
	return nil
}

func (m *LoroMap) InsertText(key string, text *LoroText) (*LoroText, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_insert_text(m.ptr, keyPtr, text.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: insert text to map, key=%s", ErrLoroInsertFailed, key))
	}
	newText := &LoroText{ptr: ret}
	runtime.SetFinalizer(newText, func(text *LoroText) {
		text.Destroy()
	})
	return newText, nil
}

func (m *LoroMap) InsertList(key string, list *LoroList) (*LoroList, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_insert_list(m.ptr, keyPtr, list.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: insert list to map, key=%s", ErrLoroInsertFailed, key))
	}
	newList := &LoroList{ptr: ret}
	runtime.SetFinalizer(newList, func(list *LoroList) {
		list.Destroy()
	})
	return newList, nil
}

func (m *LoroMap) InsertMovableList(key string, list *LoroMovableList) (*LoroMovableList, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_insert_movable_list(m.ptr, keyPtr, list.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: insert movable list to map, key=%s", ErrLoroInsertFailed, key))
	}
	newList := &LoroMovableList{ptr: ret}
	runtime.SetFinalizer(newList, func(list *LoroMovableList) {
		list.Destroy()
	})
	return newList, nil
}

func (m *LoroMap) InsertMap(key string, mapValue *LoroMap) (*LoroMap, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_insert_map(m.ptr, keyPtr, mapValue.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: insert map to map, key=%s", ErrLoroInsertFailed, key))
	}
	newMap := &LoroMap{ptr: ret}
	runtime.SetFinalizer(newMap, func(m *LoroMap) {
		m.Destroy()
	})
	return newMap, nil
}

func (m *LoroMap) IsAttached() bool {
	return C.loro_map_is_attached(m.ptr) != 0
}

// ----------- Loro List -----------

type LoroList struct {
	ptr unsafe.Pointer
}

func NewEmptyLoroList() *LoroList {
	ptr := C.loro_list_new_empty()
	list := &LoroList{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(list, func(list *LoroList) {
		list.Destroy()
	})
	return list
}

func (list *LoroList) Destroy() {
	// fmt.Println("destroying loro list")
	C.destroy_loro_list(list.ptr)
}

func (list *LoroList) ToContainer() *LoroContainer {
	ptr := C.loro_list_to_container(list.ptr)
	container := &LoroContainer{ptr: ptr}
	runtime.SetFinalizer(container, func(container *LoroContainer) {
		container.Destroy()
	})
	return container
}

func (list *LoroList) PushNull() error {
	var err C.uint8_t
	C.loro_list_push_null(list.ptr, &err)
	if err != 0 {
		return pe.WithStack(fmt.Errorf("%w: push null to list", ErrLoroInsertFailed))
	}
	return nil
}

func (list *LoroList) PushBool(value bool) (bool, error) {
	var err C.uint8_t
	boolValue := 0
	if value {
		boolValue = 1
	}
	C.loro_list_push_bool(list.ptr, C.int32_t(boolValue), &err)
	if err != 0 {
		return false, pe.WithStack(fmt.Errorf("%w: push bool to list", ErrLoroInsertFailed))
	}
	return value, nil
}

func (list *LoroList) PushDouble(value float64) (float64, error) {
	var err C.uint8_t
	C.loro_list_push_double(list.ptr, C.double(value), &err)
	if err != 0 {
		return math.NaN(), pe.WithStack(fmt.Errorf("%w: push double to list", ErrLoroInsertFailed))
	}
	return value, nil
}

func (list *LoroList) PushI64(value int64) (int64, error) {
	var err C.uint8_t
	C.loro_list_push_i64(list.ptr, C.int64_t(value), &err)
	if err != 0 {
		return -1, pe.WithStack(fmt.Errorf("%w: push i64 to list", ErrLoroInsertFailed))
	}
	return value, nil
}

func (list *LoroList) PushString(value string) (string, error) {
	var err C.uint8_t
	C.loro_list_push_string(list.ptr, C.CString(value), &err)
	if err != 0 {
		return "", pe.WithStack(fmt.Errorf("%w: push string to list", ErrLoroInsertFailed))
	}
	return value, nil
}

func (list *LoroList) PushText(textValue *LoroText) (*LoroText, error) {
	var err C.uint8_t
	ptr := C.loro_list_push_text(list.ptr, textValue.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: push text to list", ErrLoroInsertFailed))
	}
	newText := &LoroText{ptr: ptr}
	runtime.SetFinalizer(newText, func(text *LoroText) {
		text.Destroy()
	})
	return newText, nil
}

func (list *LoroList) PushList(listValue *LoroList) (*LoroList, error) {
	var err C.uint8_t
	ptr := C.loro_list_push_list(list.ptr, listValue.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: push list to list", ErrLoroInsertFailed))
	}
	newList := &LoroList{ptr: ptr}
	runtime.SetFinalizer(newList, func(list *LoroList) {
		list.Destroy()
	})
	return newList, nil
}

func (list *LoroList) PushMovableList(movableList *LoroMovableList) (*LoroMovableList, error) {
	var err C.uint8_t
	ptr := C.loro_list_push_movable_list(list.ptr, movableList.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: push movable list to list", ErrLoroInsertFailed))
	}
	newMovableList := &LoroMovableList{ptr: ptr}
	runtime.SetFinalizer(newMovableList, func(movableList *LoroMovableList) {
		movableList.Destroy()
	})
	return newMovableList, nil
}

func (list *LoroList) PushMap(mapValue *LoroMap) (*LoroMap, error) {
	var err C.uint8_t
	ptr := C.loro_list_push_map(list.ptr, mapValue.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: push map to list", ErrLoroInsertFailed))
	}
	newMap := &LoroMap{ptr: ptr}
	runtime.SetFinalizer(newMap, func(mapValue *LoroMap) {
		mapValue.Destroy()
	})
	return newMap, nil
}

func (list *LoroList) Push(value any) (any, error) {
	switch v := value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		intValue := util.ToInt64(v)
		list.PushI64(intValue)
	case float64, float32:
		doubleValue := util.ToFloat64(v)
		list.PushDouble(doubleValue)
	case bool:
		list.PushBool(v)
	case string:
		list.PushString(v)
	case *LoroText:
		list.PushText(v)
	case *LoroList:
		list.PushList(v)
	case *LoroMovableList:
		list.PushMovableList(v)
	case *LoroMap:
		list.PushMap(v)
	case nil:
		list.PushNull()
	}
	return nil, pe.WithStack(fmt.Errorf("unsupported value type: %T", value))
}

func (list *LoroList) InsertNull(index uint32) error {
	var err C.uint8_t
	C.loro_list_insert_null(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return pe.WithStack(fmt.Errorf("%w: insert null to list", ErrLoroInsertFailed))
	}
	return nil
}

func (list *LoroList) InsertBool(index uint32, value bool) error {
	var err C.uint8_t
	boolValue := 0
	if value {
		boolValue = 1
	}
	C.loro_list_insert_bool(list.ptr, C.uint32_t(index), C.int32_t(boolValue), &err)
	if err != 0 {
		return pe.WithStack(fmt.Errorf("%w: insert bool to list", ErrLoroInsertFailed))
	}
	return nil
}

func (list *LoroList) InsertDouble(index uint32, value float64) error {
	var err C.uint8_t
	C.loro_list_insert_double(list.ptr, C.uint32_t(index), C.double(value), &err)
	if err != 0 {
		return pe.WithStack(fmt.Errorf("%w: insert double to list", ErrLoroInsertFailed))
	}
	return nil
}

func (list *LoroList) InsertI64(index uint32, value int64) error {
	var err C.uint8_t
	C.loro_list_insert_i64(list.ptr, C.uint32_t(index), C.int64_t(value), &err)
	if err != 0 {
		return pe.WithStack(fmt.Errorf("%w: insert i64 to list", ErrLoroInsertFailed))
	}
	return nil
}

func (list *LoroList) InsertString(index uint32, value string) error {
	var err C.uint8_t
	C.loro_list_insert_string(list.ptr, C.uint32_t(index), C.CString(value), &err)
	if err != 0 {
		return pe.WithStack(fmt.Errorf("%w: insert string to list", ErrLoroInsertFailed))
	}
	return nil
}

func (list *LoroList) InsertText(index uint32, value *LoroText) (*LoroText, error) {
	var err C.uint8_t
	ptr := C.loro_list_insert_text(list.ptr, C.uint32_t(index), value.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: insert text to list", ErrLoroInsertFailed))
	}
	newText := &LoroText{ptr: unsafe.Pointer(ptr)}
	runtime.SetFinalizer(newText, func(text *LoroText) {
		text.Destroy()
	})
	return newText, nil
}

func (list *LoroList) InsertList(index uint32, value *LoroList) (*LoroList, error) {
	var err C.uint8_t
	ptr := C.loro_list_insert_list(list.ptr, C.uint32_t(index), value.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: insert list to list", ErrLoroInsertFailed))
	}
	newList := &LoroList{ptr: unsafe.Pointer(ptr)}
	runtime.SetFinalizer(newList, func(list *LoroList) {
		list.Destroy()
	})
	return newList, nil
}

func (list *LoroList) InsertMovableList(index uint32, value *LoroMovableList) (*LoroMovableList, error) {
	var err C.uint8_t
	ptr := C.loro_list_insert_movable_list(list.ptr, C.uint32_t(index), value.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: insert movable list to list", ErrLoroInsertFailed))
	}
	newMovableList := &LoroMovableList{ptr: unsafe.Pointer(ptr)}
	runtime.SetFinalizer(newMovableList, func(movableList *LoroMovableList) {
		movableList.Destroy()
	})
	return newMovableList, nil
}

func (list *LoroList) InsertMap(index uint32, value *LoroMap) (*LoroMap, error) {
	var err C.uint8_t
	ptr := C.loro_list_insert_map(list.ptr, C.uint32_t(index), value.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: insert map to list", ErrLoroInsertFailed))
	}
	newMap := &LoroMap{ptr: unsafe.Pointer(ptr)}
	runtime.SetFinalizer(newMap, func(mapValue *LoroMap) {
		mapValue.Destroy()
	})
	return newMap, nil
}

func (list *LoroList) GetLen() uint32 {
	return uint32(C.loro_list_len(list.ptr))
}

func (list *LoroList) Get(index uint32) *LoroContainerOrValue {
	ptr := C.loro_list_get(list.ptr, C.uint32_t(index))
	if ptr == nil {
		return nil
	}
	ret := &LoroContainerOrValue{ptr: unsafe.Pointer(ptr)}
	runtime.SetFinalizer(ret, func(ret *LoroContainerOrValue) {
		ret.Destroy()
	})
	return ret
}

func (list *LoroList) GetNull(index uint32) error {
	var err C.uint8_t
	C.loro_list_get_null(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return pe.WithStack(fmt.Errorf("%w: get null from list", ErrLoroGetFailed))
	}
	return nil
}

func (list *LoroList) GetBool(index uint32) (bool, error) {
	var err C.uint8_t
	ret := C.loro_list_get_bool(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return false, pe.WithStack(fmt.Errorf("%w: get bool from list", ErrLoroGetFailed))
	}
	return ret != 0, nil
}

func (list *LoroList) GetDouble(index uint32) (float64, error) {
	var err C.uint8_t
	ret := C.loro_list_get_double(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return 0, pe.WithStack(fmt.Errorf("%w: get double from list", ErrLoroGetFailed))
	}
	return float64(ret), nil
}

func (list *LoroList) GetI64(index uint32) (int64, error) {
	var err C.uint8_t
	ret := C.loro_list_get_i64(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return 0, pe.WithStack(fmt.Errorf("%w: get i64 from list", ErrLoroGetFailed))
	}
	return int64(ret), nil
}

func (list *LoroList) GetString(index uint32) (string, error) {
	var err C.uint8_t
	ret := C.loro_list_get_string(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return "", pe.WithStack(fmt.Errorf("%w: get string from list", ErrLoroGetFailed))
	}
	return C.GoString(ret), nil
}

func (list *LoroList) GetText(index uint32) (*LoroText, error) {
	var err C.uint8_t
	ret := C.loro_list_get_text(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: get text from list", ErrLoroGetFailed))
	}
	text := &LoroText{ptr: unsafe.Pointer(ret)}
	runtime.SetFinalizer(text, func(text *LoroText) {
		text.Destroy()
	})
	return text, nil
}

func (list *LoroList) GetList(index uint32) (*LoroList, error) {
	var err C.uint8_t
	ret := C.loro_list_get_list(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: get list from list", ErrLoroGetFailed))
	}
	newList := &LoroList{ptr: unsafe.Pointer(ret)}
	runtime.SetFinalizer(newList, func(list *LoroList) {
		list.Destroy()
	})
	return newList, nil
}

func (list *LoroList) GetMovableList(index uint32) (*LoroMovableList, error) {
	var err C.uint8_t
	ret := C.loro_list_get_movable_list(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: get movable list from list", ErrLoroGetFailed))
	}
	newMovableList := &LoroMovableList{ptr: unsafe.Pointer(ret)}
	runtime.SetFinalizer(newMovableList, func(movableList *LoroMovableList) {
		movableList.Destroy()
	})
	return newMovableList, nil
}

func (list *LoroList) GetMap(index uint32) (*LoroMap, error) {
	var err C.uint8_t
	ret := C.loro_list_get_map(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: get map from list", ErrLoroGetFailed))
	}
	newMap := &LoroMap{ptr: unsafe.Pointer(ret)}
	runtime.SetFinalizer(newMap, func(mapValue *LoroMap) {
		mapValue.Destroy()
	})
	return newMap, nil
}

func (list *LoroList) IsAttached() bool {
	return C.loro_list_is_attached(list.ptr) != 0
}

func (list *LoroList) ToGoObject() (any, error) {
	vecPtr := C.loro_list_get_items(list.ptr)
	vec := &RustPtrVec{ptr: unsafe.Pointer(vecPtr)}
	defer vec.Destroy()
	items := vec.GetData()
	result := make([]any, len(items))
	for i, ptr := range items {
		item := &LoroContainerOrValue{ptr: unsafe.Pointer(ptr)}
		defer item.Destroy()
		itemGo, err := item.ToGoObject()
		if err != nil {
			return nil, err
		}
		result[i] = itemGo
	}
	return result, nil
}

// ----------- Loro Movable List -----------
type LoroMovableList struct {
	ptr unsafe.Pointer
}

func (movableList *LoroMovableList) Destroy() {
	// fmt.Println("destroying loro movable list")
	C.destroy_loro_movable_list(movableList.ptr)
}

func NewEmptyLoroMovableList() *LoroMovableList {
	ptr := C.loro_movable_list_new_empty()
	list := &LoroMovableList{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(list, func(list *LoroMovableList) {
		list.Destroy()
	})
	return list
}

func (list *LoroMovableList) ToContainer() *LoroContainer {
	ptr := C.loro_movable_list_to_container(list.ptr)
	container := &LoroContainer{ptr: ptr}
	runtime.SetFinalizer(container, func(container *LoroContainer) {
		container.Destroy()
	})
	return container
}

func (list *LoroMovableList) GetLen() uint32 {
	return uint32(C.loro_movable_list_len(list.ptr))
}

func (list *LoroMovableList) PushNull() (any, error) {
	var err C.uint8_t
	C.loro_movable_list_push_null(list.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: push null to movable list", ErrLoroInsertFailed))
	}
	return nil, nil
}

func (list *LoroMovableList) PushBool(value bool) (bool, error) {
	var err C.uint8_t
	boolValue := 0
	if value {
		boolValue = 1
	}
	C.loro_movable_list_push_bool(list.ptr, C.int32_t(boolValue), &err)
	if err != 0 {
		return false, pe.WithStack(fmt.Errorf("%w: push bool to movable list", ErrLoroInsertFailed))
	}
	return value, nil
}

func (list *LoroMovableList) PushDouble(value float64) (float64, error) {
	var err C.uint8_t
	C.loro_movable_list_push_double(list.ptr, C.double(value), &err)
	if err != 0 {
		return math.NaN(), pe.WithStack(fmt.Errorf("%w: push double to movable list", ErrLoroInsertFailed))
	}
	return value, nil
}

func (list *LoroMovableList) PushI64(value int64) (int64, error) {
	var err C.uint8_t
	C.loro_movable_list_push_i64(list.ptr, C.int64_t(value), &err)
	if err != 0 {
		return -1, pe.WithStack(fmt.Errorf("%w: push i64 to movable list", ErrLoroInsertFailed))
	}
	return value, nil
}

func (list *LoroMovableList) PushString(value string) (string, error) {
	var err C.uint8_t
	C.loro_movable_list_push_string(list.ptr, C.CString(value), &err)
	if err != 0 {
		return "", pe.WithStack(fmt.Errorf("%w: push string to movable list", ErrLoroInsertFailed))
	}
	return value, nil
}

func (list *LoroMovableList) PushText(textValue *LoroText) (*LoroText, error) {
	var err C.uint8_t
	ptr := C.loro_movable_list_push_text(list.ptr, textValue.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: push text to movable list", ErrLoroInsertFailed))
	}
	newText := &LoroText{ptr: ptr}
	runtime.SetFinalizer(newText, func(text *LoroText) {
		text.Destroy()
	})
	return newText, nil
}

func (list *LoroMovableList) PushList(listValue *LoroList) (*LoroList, error) {
	var err C.uint8_t
	ptr := C.loro_movable_list_push_list(list.ptr, listValue.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: push list to movable list", ErrLoroInsertFailed))
	}
	newList := &LoroList{ptr: ptr}
	runtime.SetFinalizer(newList, func(list *LoroList) {
		list.Destroy()
	})
	return newList, nil
}

func (list *LoroMovableList) PushMovableList(movableList *LoroMovableList) (*LoroMovableList, error) {
	var err C.uint8_t
	ptr := C.loro_movable_list_push_movable_list(list.ptr, movableList.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: push movable list to movable list", ErrLoroInsertFailed))
	}
	newMovableList := &LoroMovableList{ptr: ptr}
	runtime.SetFinalizer(newMovableList, func(movableList *LoroMovableList) {
		movableList.Destroy()
	})
	return newMovableList, nil
}

func (list *LoroMovableList) PushMap(mapValue *LoroMap) (*LoroMap, error) {
	var err C.uint8_t
	ptr := C.loro_movable_list_push_map(list.ptr, mapValue.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: push map to movable list", ErrLoroInsertFailed))
	}
	newMap := &LoroMap{ptr: ptr}
	runtime.SetFinalizer(newMap, func(mapValue *LoroMap) {
		mapValue.Destroy()
	})
	return newMap, nil
}

func (list *LoroMovableList) Push(value any) (any, error) {
	switch v := value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		intValue := util.ToInt64(v)
		list.PushI64(intValue)
	case float64, float32:
		doubleValue := util.ToFloat64(v)
		list.PushDouble(doubleValue)
	case bool:
		list.PushBool(v)
	case string:
		list.PushString(v)
	case *LoroText:
		list.PushText(v)
	case *LoroList:
		list.PushList(v)
	case *LoroMovableList:
		list.PushMovableList(v)
	case *LoroMap:
		list.PushMap(v)
	case nil:
		list.PushNull()
	}
	return nil, pe.WithStack(fmt.Errorf("unsupported value type: %T", value))
}

func (list *LoroMovableList) Get(index uint32) *LoroContainerOrValue {
	ptr := C.loro_movable_list_get(list.ptr, C.uint32_t(index))
	if ptr == nil {
		return nil
	}
	ret := &LoroContainerOrValue{ptr: unsafe.Pointer(ptr)}
	runtime.SetFinalizer(ret, func(ret *LoroContainerOrValue) {
		ret.Destroy()
	})
	return ret
}

func (list *LoroMovableList) GetNull(index uint32) error {
	var err C.uint8_t
	C.loro_movable_list_get_null(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return pe.WithStack(fmt.Errorf("%w: get null from movable list", ErrLoroGetFailed))
	}
	return nil
}

func (list *LoroMovableList) GetBool(index uint32) (bool, error) {
	var err C.uint8_t
	ret := C.loro_movable_list_get_bool(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return false, pe.WithStack(fmt.Errorf("%w: get bool from movable list", ErrLoroGetFailed))
	}
	return ret != 0, nil
}

func (list *LoroMovableList) GetDouble(index uint32) (float64, error) {
	var err C.uint8_t
	ret := C.loro_movable_list_get_double(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return 0, pe.WithStack(fmt.Errorf("%w: get double from movable list", ErrLoroGetFailed))
	}
	return float64(ret), nil
}

func (list *LoroMovableList) GetI64(index uint32) (int64, error) {
	var err C.uint8_t
	ret := C.loro_movable_list_get_i64(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return 0, pe.WithStack(fmt.Errorf("%w: get i64 from movable list", ErrLoroGetFailed))
	}
	return int64(ret), nil
}

func (list *LoroMovableList) GetString(index uint32) (string, error) {
	var err C.uint8_t
	ret := C.loro_movable_list_get_string(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return "", pe.WithStack(fmt.Errorf("%w: get string from movable list", ErrLoroGetFailed))
	}
	return C.GoString(ret), nil
}

func (list *LoroMovableList) GetText(index uint32) (*LoroText, error) {
	var err C.uint8_t
	ret := C.loro_movable_list_get_text(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: get text from movable list", ErrLoroGetFailed))
	}
	text := &LoroText{ptr: unsafe.Pointer(ret)}
	runtime.SetFinalizer(text, func(text *LoroText) {
		text.Destroy()
	})
	return text, nil
}

func (list *LoroMovableList) GetList(index uint32) (*LoroList, error) {
	var err C.uint8_t
	ret := C.loro_movable_list_get_list(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: get list from movable list", ErrLoroGetFailed))
	}
	newList := &LoroList{ptr: unsafe.Pointer(ret)}
	runtime.SetFinalizer(newList, func(list *LoroList) {
		list.Destroy()
	})
	return newList, nil
}

func (list *LoroMovableList) GetMovableList(index uint32) (*LoroMovableList, error) {
	var err C.uint8_t
	ret := C.loro_movable_list_get_movable_list(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: get movable list from movable list", ErrLoroGetFailed))
	}
	newMovableList := &LoroMovableList{ptr: unsafe.Pointer(ret)}
	runtime.SetFinalizer(newMovableList, func(movableList *LoroMovableList) {
		movableList.Destroy()
	})
	return newMovableList, nil
}

func (list *LoroMovableList) GetMap(index uint32) (*LoroMap, error) {
	var err C.uint8_t
	ret := C.loro_movable_list_get_map(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: get map from movable list", ErrLoroGetFailed))
	}
	newMap := &LoroMap{ptr: unsafe.Pointer(ret)}
	runtime.SetFinalizer(newMap, func(mapValue *LoroMap) {
		mapValue.Destroy()
	})
	return newMap, nil
}

func (list *LoroMovableList) IsAttached() bool {
	return C.loro_movable_list_is_attached(list.ptr) != 0
}

func (list *LoroMovableList) ToGoObject() (any, error) {
	vecPtr := C.loro_movable_list_get_items(list.ptr)
	vec := &RustPtrVec{ptr: unsafe.Pointer(vecPtr)}
	defer vec.Destroy()
	items := vec.GetData()
	result := make([]any, len(items))
	for i, ptr := range items {
		item := &LoroContainerOrValue{ptr: unsafe.Pointer(ptr)}
		defer item.Destroy()
		itemGo, err := item.ToGoObject()
		if err != nil {
			return nil, err
		}
		result[i] = itemGo
	}
	return result, nil
}

// -------------- Loro Tree --------------

type LoroTree struct {
	ptr unsafe.Pointer
}

func (t *LoroTree) Destroy() {
	C.destroy_loro_tree(t.ptr)
}

func (t *LoroTree) ToContainer() *LoroContainer {
	ptr := C.loro_tree_to_container(t.ptr)
	container := &LoroContainer{ptr: ptr}
	runtime.SetFinalizer(container, func(container *LoroContainer) {
		container.Destroy()
	})
	return container
}

func (t *LoroTree) IsAttached() bool {
	return C.loro_tree_is_attached(t.ptr) != 0
}

// ----------- OpId -----------

type OpId struct {
	cLayoutId C.CLayoutID
}

func NewOpId(peer uint64, counter int32) *OpId {
	opId := &OpId{
		cLayoutId: C.CLayoutID{
			peer:    C.uint64_t(peer),
			counter: C.uint32_t(counter),
		},
	}
	return opId
}

func (id *OpId) GetPeer() uint64 {
	return uint64(id.cLayoutId.peer)
}

func (id *OpId) GetCounter() int32 {
	return int32(id.cLayoutId.counter)
}

// ----------- Diff Batch -----------

type DiffBatch struct {
	ptr unsafe.Pointer
}

type CidEventPair struct {
	ContainerId ContainerId
	DiffEvent   DiffEvent
}

func (d *DiffBatch) Destroy() {
	C.destroy_diff_batch(d.ptr)
}

func (d *DiffBatch) GetEvents() []CidEventPair {
	var cidsPtr unsafe.Pointer
	var eventsPtr unsafe.Pointer
	C.diff_batch_events(d.ptr, &cidsPtr, &eventsPtr)

	cidVec := &RustPtrVec{ptr: cidsPtr}
	eventVec := &RustPtrVec{ptr: eventsPtr}
	cids := cidVec.GetData()
	events := eventVec.GetData()

	pairs := make([]CidEventPair, len(cids))
	for i := 0; i < len(cids); i++ {
		cidPtr := cids[i]
		eventPtr := events[i]
		cid := &ContainerId{ptr: cidPtr}
		diffEvent := &DiffEvent{ptr: eventPtr}
		pair := CidEventPair{
			ContainerId: *cid,
			DiffEvent:   *diffEvent,
		}
		runtime.SetFinalizer(cid, func(cid *ContainerId) {
			cid.Destroy()
		})
		runtime.SetFinalizer(diffEvent, func(event *DiffEvent) {
			event.Destroy()
		})
		pairs[i] = pair
	}
	// 在从 vec 中拿到数据后，vec 就不需要了，手动销毁
	cidVec.Destroy()
	eventVec.Destroy()
	return pairs
}

// ------------- ContainerId and DiffEvent ------------

type CidAndDiffEvent struct {
	cidPtr   ContainerId
	eventPtr DiffEvent
}

// ----------- Container Id -----------

type ContainerId struct {
	ptr unsafe.Pointer
}

const (
	CONTAINER_TYPE_MAP          = 0
	CONTAINER_TYPE_LIST         = 1
	CONTAINER_TYPE_TEXT         = 2
	CONTAINER_TYPE_TREE         = 3
	CONTAINER_TYPE_MOVABLE_LIST = 4
	CONTAINER_TYPE_COUNTER      = 5
	CONTAINER_TYPE_UNKNOWN      = 6
)

func (cid *ContainerId) Destroy() {
	C.destroy_container_id(cid.ptr)
}

func (cid *ContainerId) IsRoot() bool {
	return C.is_container_id_root(cid.ptr) != 0
}

func (cid *ContainerId) IsNormal() bool {
	return C.is_container_id_normal(cid.ptr) != 0
}

func (cid *ContainerId) GetRootName() string {
	ret := C.container_id_root_name(cid.ptr)
	return C.GoString(ret)
}

func (cid *ContainerId) GetNormalPeer() uint64 {
	return uint64(C.container_id_normal_peer(cid.ptr))
}

func (cid *ContainerId) GetNormalCounter() int32 {
	return int32(C.container_id_normal_counter(cid.ptr))
}

func (cid *ContainerId) GetContainerType() int32 {
	return int32(C.container_id_container_type(cid.ptr))
}

// ------------- DiffEvent ------------

type DiffEvent struct {
	ptr unsafe.Pointer
}

const (
	DIFF_EVENT_TYPE_LIST    = 0
	DIFF_EVENT_TYPE_TEXT    = 1
	DIFF_EVENT_TYPE_MAP     = 2
	DIFF_EVENT_TYPE_TREE    = 3
	DIFF_EVENT_TYPE_UNKNOWN = 4
)

func (de *DiffEvent) GetType() int32 {
	return int32(C.diff_event_get_type(de.ptr))
}

func (de *DiffEvent) Destroy() {
	C.destroy_diff_event(de.ptr)
}

func (de *DiffEvent) GetListDiff() []*ListDiffItem {
	ptr := C.diff_event_get_list_diff(de.ptr)
	if ptr == nil {
		return nil
	}
	itemsVec := &RustPtrVec{ptr: ptr}
	items := make([]*ListDiffItem, itemsVec.GetLen())
	for i, itemPtr := range itemsVec.GetData() {
		item := &ListDiffItem{ptr: itemPtr}
		runtime.SetFinalizer(item, func(item *ListDiffItem) {
			item.Destroy()
		})
		items[i] = item
	}
	itemsVec.Destroy()
	return items
}

func (de *DiffEvent) GetTextDiff() []*TextDelta {
	ptr := C.diff_event_get_text_delta(de.ptr)
	if ptr == nil {
		return nil
	}
	itemsVec := &RustPtrVec{ptr: ptr}
	items := make([]*TextDelta, itemsVec.GetLen())
	for i, itemPtr := range itemsVec.GetData() {
		item := &TextDelta{ptr: itemPtr}
		runtime.SetFinalizer(item, func(item *TextDelta) {
			item.Destroy()
		})
		items[i] = item
	}
	itemsVec.Destroy()
	return items
}

func (de *DiffEvent) GetMapDiff() *MapDelta {
	ptr := C.diff_event_get_map_delta(de.ptr)
	if ptr == nil {
		return nil
	}
	mapDelta := &MapDelta{ptr: ptr}
	runtime.SetFinalizer(mapDelta, func(md *MapDelta) {
		md.Destroy()
	})
	return mapDelta
}

func (de *DiffEvent) GetTreeDiff() *TreeDiff {
	ptr := C.diff_event_get_tree_diff(de.ptr)
	if ptr == nil {
		return nil
	}
	treeDiff := &TreeDiff{ptr: ptr}
	runtime.SetFinalizer(treeDiff, func(td *TreeDiff) {
		td.Destroy()
	})
	return treeDiff
}

// ------------ List Diff Item -----------

const (
	LIST_DIFF_ITEM_INSERT ListDiffItemType = 0
	LIST_DIFF_ITEM_DELETE ListDiffItemType = 1
	LIST_DIFF_ITEM_RETAIN ListDiffItemType = 2
)

type ListDiffItemType int32

type ListDiffItem struct {
	ptr unsafe.Pointer
}

func (li *ListDiffItem) Destroy() {
	C.destroy_list_diff_item(li.ptr)
}

func (li *ListDiffItem) GetType() ListDiffItemType {
	return ListDiffItemType(C.list_diff_item_get_type(li.ptr))
}

func (li *ListDiffItem) GetInserted() ([]*LoroContainerOrValue, bool, error) {
	var err C.uint8_t
	var is_move C.uint8_t
	inserted := C.list_diff_item_get_insert(li.ptr, &is_move, &err)
	if err != 0 {
		return nil, false, pe.WithStack(fmt.Errorf("%w: get inserted from list diff item", ErrLoroGetFailed))
	}
	vec := &RustPtrVec{ptr: inserted}
	items := make([]*LoroContainerOrValue, vec.GetLen())
	for i, itemPtr := range vec.GetData() {
		item := &LoroContainerOrValue{ptr: itemPtr}
		runtime.SetFinalizer(item, func(item *LoroContainerOrValue) {
			item.Destroy()
		})
		items[i] = item
	}
	vec.Destroy()
	return items, is_move == 1, nil
}

func (li *ListDiffItem) GetDeleteCount() (uint32, error) {
	var err C.uint8_t
	count := C.list_diff_item_get_delete_count(li.ptr, &err)
	if err != 0 {
		return 0, pe.WithStack(fmt.Errorf("%w: get delete count from list diff item", ErrLoroGetFailed))
	}
	return uint32(count), nil
}

func (li *ListDiffItem) GetRetainCount() (uint32, error) {
	var err C.uint8_t
	count := C.list_diff_item_get_retain_count(li.ptr, &err)
	if err != 0 {
		return 0, pe.WithStack(fmt.Errorf("%w: get retain count from list diff item", ErrLoroGetFailed))
	}
	return uint32(count), nil
}

// ------------  Text Delta -----------

type TextDelta struct {
	ptr unsafe.Pointer
}

func (td *TextDelta) Destroy() {
	C.destroy_text_delta(td.ptr)
}

// ------------ Map Delta -----------

type MapDelta struct {
	ptr unsafe.Pointer
}

func (md *MapDelta) Destroy() {
	C.destroy_map_delta(md.ptr)
}

// ------------ Tree Diff -----------

type TreeDiff struct {
	ptr unsafe.Pointer
}

func (td *TreeDiff) Destroy() {
	C.destroy_tree_diff(td.ptr)
}

// ------------ LoroValue -----------

type LoroValueType int32

const (
	LORO_NULL_VALUE      LoroValueType = 0
	LORO_BOOL_VALUE      LoroValueType = 1
	LORO_DOUBLE_VALUE    LoroValueType = 2
	LORO_I64_VALUE       LoroValueType = 3
	LORO_STRING_VALUE    LoroValueType = 4
	LORO_MAP_VALUE       LoroValueType = 5
	LORO_LIST_VALUE      LoroValueType = 6
	LORO_BINARY_VALUE    LoroValueType = 7
	LORO_CONTAINER_VALUE LoroValueType = 8
	LORO_CONTAINER_ID    LoroValueType = 9
)

type LoroValue struct {
	ptr unsafe.Pointer
}

func (lv *LoroValue) Unwrap() (any, error) {
	t := lv.GetType()
	switch t {
	case LORO_NULL_VALUE:
		return nil, nil
	case LORO_BOOL_VALUE:
		return lv.GetBool()
	case LORO_DOUBLE_VALUE:
		return lv.GetDouble()
	case LORO_I64_VALUE:
		return lv.GetI64()
	case LORO_STRING_VALUE:
		return lv.GetString()
	case LORO_BINARY_VALUE:
		b, err := lv.GetBinary()
		if err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	case LORO_MAP_VALUE:
		return lv.GetMap()
	case LORO_LIST_VALUE:
		return lv.GetList()
	}
	return nil, pe.WithStack(fmt.Errorf("unknown loro value type: %d", t))
}

func (lv *LoroValue) IsComparable() bool {
	t := lv.GetType()
	return t != LORO_MAP_VALUE && t != LORO_LIST_VALUE
}

func (lv *LoroValue) Compare(lv2 *LoroValue) (int, error) {
	if !lv.IsComparable() || !lv2.IsComparable() {
		return 0, pe.WithStack(fmt.Errorf("comparable type required for comparison"))
	}

	t1 := lv.GetType()
	t2 := lv2.GetType()

	if t1 == LORO_NULL_VALUE {
		if t2 == LORO_NULL_VALUE {
			return 0, nil
		} else {
			return -1, nil
		}
	}

	if t2 == LORO_NULL_VALUE {
		return 1, nil
	}

	if t1 == LORO_BOOL_VALUE && t2 == LORO_BOOL_VALUE {
		val1, err := lv.GetBool()
		if err != nil {
			return 0, pe.WithStack(fmt.Errorf("get bool value: %w", err))
		}
		val2, err := lv2.GetBool()
		if err != nil {
			return 0, pe.WithStack(fmt.Errorf("get bool value: %w", err))
		}
		if val1 && !val2 {
			return -1, nil
		} else if !val1 && val2 {
			return 1, nil
		} else {
			return 0, nil
		}
	}

	if t1 == LORO_DOUBLE_VALUE && t2 == LORO_DOUBLE_VALUE {
		val1, err := lv.GetDouble()
		if err != nil {
			return 0, pe.WithStack(fmt.Errorf("get double value: %w", err))
		}
		val2, err := lv2.GetDouble()
		if err != nil {
			return 0, pe.WithStack(fmt.Errorf("get double value: %w", err))
		}
		if val1 < val2 {
			return -1, nil
		} else if val1 > val2 {
			return 1, nil
		} else {
			return 0, nil
		}
	}

	if t1 == LORO_I64_VALUE && t2 == LORO_I64_VALUE {
		val1, err := lv.GetI64()
		if err != nil {
			return 0, pe.WithStack(fmt.Errorf("get i64 value: %w", err))
		}
		val2, err := lv2.GetI64()
		if err != nil {
			return 0, pe.WithStack(fmt.Errorf("get i64 value: %w", err))
		}
		if val1 < val2 {
			return -1, nil
		} else if val1 > val2 {
			return 1, nil
		} else {
			return 0, nil
		}
	}

	if t1 == LORO_STRING_VALUE && t2 == LORO_STRING_VALUE {
		val1, err := lv.GetString()
		if err != nil {
			return 0, pe.WithStack(fmt.Errorf("get string value: %w", err))
		}
		val2, err := lv2.GetString()
		if err != nil {
			return 0, pe.WithStack(fmt.Errorf("get string value: %w", err))
		}
		if val1 < val2 {
			return -1, nil
		} else if val1 > val2 {
			return 1, nil
		} else {
			return 0, nil
		}
	}

	if t1 == LORO_BINARY_VALUE && t2 == LORO_BINARY_VALUE {
		val1, err := lv.GetBinary()
		if err != nil {
			return 0, pe.WithStack(fmt.Errorf("get binary value: %w", err))
		}
		val2, err := lv2.GetBinary()
		if err != nil {
			return 0, pe.WithStack(fmt.Errorf("get binary value: %w", err))
		}
		val1Bytes := val1.Bytes()
		val2Bytes := val2.Bytes()
		cmp := bytes.Compare(val1Bytes, val2Bytes)
		return cmp, nil
	}

	return 0, pe.WithStack(fmt.Errorf("unknown loro value type: %d", t1))
}

func (lv *LoroValue) MustCompare(lv2 *LoroValue) int {
	cmp, err := lv.Compare(lv2)
	if err != nil {
		panic(err)
	}
	return cmp
}

func NewLoroValueFromJson(json string) (*LoroValue, error) {
	ptr := C.loro_value_from_json(C.CString(json))
	if ptr == nil {
		return nil, pe.WithStack(fmt.Errorf("%w: from json \"%s\"", ErrLoroDecodeFailed, json))
	}
	lv := &LoroValue{ptr: ptr}
	runtime.SetFinalizer(lv, func(lv *LoroValue) {
		lv.Destroy()
	})
	return lv, nil
}

func NewLoroValue(value any) (*LoroValue, error) {
	switch v := value.(type) {
	case *LoroValue:
		return v, nil
	case nil:
		return NewLoroValueNull(), nil
	case bool:
		return NewLoroValueBool(v), nil
	case float32, float64:
		return NewLoroValueDouble(util.ToFloat64(v)), nil
	case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
		return NewLoroValueI64(util.ToInt64(v)), nil
	case string:
		return NewLoroValueString(v), nil
	case []byte:
		return NewLoroValueBinary(v), nil
	case []any:
		l := make([]*LoroValue, len(v))
		for i, v := range v {
			lv, err := NewLoroValue(v)
			if err != nil {
				return nil, err
			}
			l[i] = lv
		}
		return NewLoroValueList(l), nil
	case map[string]any:
		m := make(map[string]*LoroValue)
		for k, v := range v {
			lv, err := NewLoroValue(v)
			if err != nil {
				return nil, err
			}
			m[k] = lv
		}
		return NewLoroValueMap(m), nil
	default:
		return nil, pe.WithStack(fmt.Errorf("%w: convert %T to loro value", ErrLoroEncodeFailed, value))
	}
}

func NewLoroValueNull() *LoroValue {
	ptr := C.loro_value_new_null()
	lv := &LoroValue{ptr: ptr}
	runtime.SetFinalizer(lv, func(lv *LoroValue) {
		lv.Destroy()
	})
	return lv
}

func NewLoroValueBool(value bool) *LoroValue {
	boolValue := 0
	if value {
		boolValue = 1
	}
	ptr := C.loro_value_new_bool(C.int(boolValue))
	lv := &LoroValue{ptr: ptr}
	runtime.SetFinalizer(lv, func(lv *LoroValue) {
		lv.Destroy()
	})
	return lv
}

func NewLoroValueDouble(value float64) *LoroValue {
	ptr := C.loro_value_new_double(C.double(value))
	lv := &LoroValue{ptr: ptr}
	runtime.SetFinalizer(lv, func(lv *LoroValue) {
		lv.Destroy()
	})
	return lv
}

func NewLoroValueI64(value int64) *LoroValue {
	ptr := C.loro_value_new_i64(C.int64_t(value))
	lv := &LoroValue{ptr: ptr}
	runtime.SetFinalizer(lv, func(lv *LoroValue) {
		lv.Destroy()
	})
	return lv
}

func NewLoroValueString(value string) *LoroValue {
	ptr := C.loro_value_new_string(C.CString(value))
	lv := &LoroValue{ptr: ptr}
	runtime.SetFinalizer(lv, func(lv *LoroValue) {
		lv.Destroy()
	})
	return lv
}

func NewLoroValueBinary(value []byte) *LoroValue {
	bytesVec := NewRustBytesVec(value)
	defer bytesVec.Destroy()
	ptr := C.loro_value_new_binary(bytesVec.ptr)
	lv := &LoroValue{ptr: ptr}
	runtime.SetFinalizer(lv, func(lv *LoroValue) {
		lv.Destroy()
	})
	return lv
}

func NewLoroValueList(value []*LoroValue) *LoroValue {
	ptrVec := NewRustPtrVec()
	defer ptrVec.Destroy()
	for _, v := range value {
		ptrVec.Push(v.ptr)
	}
	ptr := C.loro_value_new_list(ptrVec.ptr)
	lv := &LoroValue{ptr: ptr}
	runtime.SetFinalizer(lv, func(lv *LoroValue) {
		lv.Destroy()
	})
	return lv
}

func NewLoroValueListDeep(value []*LoroValue) *LoroValue {
	ptrVec := NewRustPtrVec()
	defer ptrVec.Destroy()
	for _, v := range value {
		ptrVec.Push(v.ptr)
	}
	ptr := C.loro_value_new_list(ptrVec.ptr)
	lv := &LoroValue{ptr: ptr}
	runtime.SetFinalizer(lv, func(lv *LoroValue) {
		lv.Destroy()
	})
	return lv
}

func NewLoroValueMap(value map[string]*LoroValue) *LoroValue {
	ptrVec := NewRustPtrVec()
	defer ptrVec.Destroy()
	for k, v := range value {
		kPtr := C.CString(k)
		ptrVec.Push(unsafe.Pointer(kPtr))
		ptrVec.Push(v.ptr)
	}
	ptr := C.loro_value_new_map(ptrVec.ptr)
	lv := &LoroValue{ptr: ptr}
	runtime.SetFinalizer(lv, func(lv *LoroValue) {
		lv.Destroy()
	})
	return lv
}

func (lv *LoroValue) Destroy() {
	C.destroy_loro_value(lv.ptr)
}

func (lv *LoroValue) GetType() LoroValueType {
	t := C.loro_value_get_type(lv.ptr)
	return LoroValueType(t)
}

func (lv *LoroValue) Get() (any, error) {
	t := lv.GetType()
	switch t {
	case LORO_NULL_VALUE:
		return nil, nil
	case LORO_BOOL_VALUE:
		bv, err := lv.GetBool()
		if err != nil {
			return nil, err
		}
		return bv, nil
	case LORO_DOUBLE_VALUE:
		dv, err := lv.GetDouble()
		if err != nil {
			return nil, err
		}
		return dv, nil
	case LORO_I64_VALUE:
		iv, err := lv.GetI64()
		if err != nil {
			return nil, err
		}
		return iv, nil
	case LORO_STRING_VALUE:
		sv, err := lv.GetString()
		if err != nil {
			return nil, err
		}
		return sv, nil
	case LORO_MAP_VALUE:
		mv, err := lv.GetMap()
		if err != nil {
			return nil, err
		}
		return mv, nil
	case LORO_LIST_VALUE:
		lv, err := lv.GetList()
		if err != nil {
			return nil, err
		}
		return lv, nil
	case LORO_BINARY_VALUE:
		bv, err := lv.GetBinary()
		if err != nil {
			return nil, err
		}
		return bv.Bytes(), nil
	default:
		return nil, pe.WithStack(fmt.Errorf("unknown loro value type: %d", t))
	}
}

func (lv *LoroValue) GetBool() (bool, error) {
	var err C.uint8_t
	ret := C.loro_value_get_bool(lv.ptr, &err)
	if err != 0 {
		return false, pe.WithStack(fmt.Errorf("%w: get bool from loro value", ErrLoroGetFailed))
	}
	return ret != 0, nil
}

func (lv *LoroValue) GetDouble() (float64, error) {
	var err C.uint8_t
	ret := C.loro_value_get_double(lv.ptr, &err)
	if err != 0 {
		return 0, pe.WithStack(fmt.Errorf("%w: get double from loro value", ErrLoroGetFailed))
	}
	return float64(ret), nil
}

func (lv *LoroValue) GetI64() (int64, error) {
	var err C.uint8_t
	ret := C.loro_value_get_i64(lv.ptr, &err)
	if err != 0 {
		return 0, pe.WithStack(fmt.Errorf("%w: get i64 from loro value", ErrLoroGetFailed))
	}
	return int64(ret), nil
}

func (lv *LoroValue) GetString() (string, error) {
	var err C.uint8_t
	ret := C.loro_value_get_string(lv.ptr, &err)
	if err != 0 {
		return "", pe.WithStack(fmt.Errorf("%w: get string from loro value", ErrLoroGetFailed))
	}
	return C.GoString(ret), nil
}

func (lv *LoroValue) GetMap() (map[string]*LoroValue, error) {
	var err C.uint8_t
	ptr := C.loro_value_get_map(lv.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: get map from loro value", ErrLoroGetFailed))
	}
	ptrVec := &RustPtrVec{ptr: ptr}
	items := make(map[string]*LoroValue)
	ptrVecLen := int(ptrVec.GetLen())
	ptrVecData := ptrVec.GetData()
	for i := 0; i < ptrVecLen; i += 2 {
		keyPtr := ptrVecData[i]
		valPtr := ptrVecData[i+1]
		key := C.GoString((*C.char)(keyPtr))
		val := &LoroValue{ptr: valPtr}
		runtime.SetFinalizer(val, func(val *LoroValue) {
			val.Destroy()
		})
		items[key] = val
	}
	ptrVec.Destroy()
	return items, nil
}

func (lv *LoroValue) GetList() ([]*LoroValue, error) {
	var err C.uint8_t
	ptr := C.loro_value_get_list(lv.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: get list from loro value", ErrLoroGetFailed))
	}
	ptrVec := &RustPtrVec{ptr: ptr}
	items := make([]*LoroValue, ptrVec.GetLen())
	for i, itemPtr := range ptrVec.GetData() {
		item := &LoroValue{ptr: itemPtr}
		runtime.SetFinalizer(item, func(item *LoroValue) {
			item.Destroy()
		})
		items[i] = item
	}
	ptrVec.Destroy()
	return items, nil
}

func (lv *LoroValue) GetBinary() (*RustBytesVec, error) {
	var err C.uint8_t
	ptr := C.loro_value_get_binary(lv.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: get binary from loro value", ErrLoroGetFailed))
	}
	bytesVec := &RustBytesVec{ptr: ptr}
	runtime.SetFinalizer(bytesVec, func(bytesVec *RustBytesVec) {
		bytesVec.Destroy()
	})
	return bytesVec, nil
}

func (lv *LoroValue) GetContainerId() (*ContainerId, error) {
	var err C.uint8_t
	ptr := C.loro_value_get_container_id(lv.ptr, &err)
	if err != 0 {
		return nil, pe.WithStack(fmt.Errorf("%w: get container id from loro value", ErrLoroGetFailed))
	}
	cid := &ContainerId{ptr: ptr}
	runtime.SetFinalizer(cid, func(cid *ContainerId) {
		cid.Destroy()
	})
	return cid, nil
}

func (lv *LoroValue) ToJson() (string, error) {
	ptr := C.loro_value_to_json(lv.ptr)
	if ptr == nil {
		return "", pe.WithStack(fmt.Errorf("%w: dump json from loro value", ErrLoroEncodeFailed))
	}
	return C.GoString(ptr), nil
}

// ToGoObject 将 loro 值转换为 go 对象
// 支持的类型:
//
//   - LORO_NULL_VALUE => nil
//   - LORO_BOOL_VALUE => bool
//   - LORO_I64_VALUE => int64
//   - LORO_DOUBLE_VALUE => float64
//   - LORO_STRING_VALUE => string
//   - LORO_BINARY_VALUE => []byte
//   - LORO_MAP_VALUE => map[string]any
//   - LORO_LIST_VALUE => []any
func (lv *LoroValue) ToGoObject() (any, error) {
	t := lv.GetType()
	switch t {
	case LORO_NULL_VALUE:
		return nil, nil
	case LORO_BOOL_VALUE:
		return lv.GetBool()
	case LORO_I64_VALUE:
		return lv.GetI64()
	case LORO_DOUBLE_VALUE:
		return lv.GetDouble()
	case LORO_STRING_VALUE:
		return lv.GetString()
	case LORO_BINARY_VALUE:
		b, err := lv.GetBinary()
		if err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	case LORO_MAP_VALUE:
		m, err := lv.GetMap()
		if err != nil {
			return nil, err
		}
		m2 := make(map[string]any)
		for k, v := range m {
			goV, err := v.ToGoObject()
			if err != nil {
				return nil, err
			}
			m2[k] = goV
		}
		return m2, nil
	case LORO_LIST_VALUE:
		l, err := lv.GetList()
		if err != nil {
			return nil, err
		}
		l2 := make([]any, len(l))
		for i, v := range l {
			goV, err := v.ToGoObject()
			if err != nil {
				return nil, err
			}
			l2[i] = goV
		}
		return l2, nil
	}
	return nil, pe.WithStack(fmt.Errorf("unknown loro value type: %d", t))
}

// -------------- Loro Container --------------

const (
	LORO_CONTAINER_MAP          = 0
	LORO_CONTAINER_LIST         = 1
	LORO_CONTAINER_TEXT         = 2
	LORO_CONTAINER_TREE         = 3
	LORO_CONTAINER_MOVABLE_LIST = 4
	LORO_CONTAINER_COUNTER      = 5
	LORO_CONTAINER_UNKNOWN      = 6
)

type LoroContainerType int32

type LoroContainer struct {
	ptr unsafe.Pointer
}

func (c *LoroContainer) Destroy() {
	C.destroy_loro_container(c.ptr)
}

func (c *LoroContainer) GetType() LoroContainerType {
	t := C.loro_container_get_type(c.ptr)
	return LoroContainerType(t)
}

func (c *LoroContainer) GetList() (*LoroList, error) {
	ptr := C.loro_container_get_list(c.ptr)
	if ptr == nil {
		return nil, pe.WithStack(fmt.Errorf("%w: get list from loro container", ErrLoroGetFailed))
	}
	list := &LoroList{ptr: ptr}
	runtime.SetFinalizer(list, func(list *LoroList) {
		list.Destroy()
	})
	return list, nil
}

func (c *LoroContainer) GetMap() (*LoroMap, error) {
	ptr := C.loro_container_get_map(c.ptr)
	if ptr == nil {
		return nil, pe.WithStack(fmt.Errorf("%w: get map from loro container", ErrLoroGetFailed))
	}
	m := &LoroMap{ptr: ptr}
	runtime.SetFinalizer(m, func(m *LoroMap) {
		m.Destroy()
	})
	return m, nil
}

func (c *LoroContainer) GetText() (*LoroText, error) {
	ptr := C.loro_container_get_text(c.ptr)
	if ptr == nil {
		return nil, pe.WithStack(fmt.Errorf("%w: get text from loro container", ErrLoroGetFailed))
	}
	text := &LoroText{ptr: ptr}
	runtime.SetFinalizer(text, func(text *LoroText) {
		text.Destroy()
	})
	return text, nil
}

func (c *LoroContainer) GetMovableList() (*LoroMovableList, error) {
	ptr := C.loro_container_get_movable_list(c.ptr)
	if ptr == nil {
		return nil, pe.WithStack(fmt.Errorf("%w: get movable list from loro container", ErrLoroGetFailed))
	}
	list := &LoroMovableList{ptr: ptr}
	runtime.SetFinalizer(list, func(list *LoroMovableList) {
		list.Destroy()
	})
	return list, nil
}

func (c *LoroContainer) GetTree() (*LoroTree, error) {
	ptr := C.loro_container_get_tree(c.ptr)
	if ptr == nil {
		return nil, pe.WithStack(fmt.Errorf("%w: get tree from loro container", ErrLoroGetFailed))
	}
	tree := &LoroTree{ptr: ptr}
	runtime.SetFinalizer(tree, func(tree *LoroTree) {
		tree.Destroy()
	})
	return tree, nil
}

func (c *LoroContainer) ToGoObject() (any, error) {
	t := c.GetType()
	switch t {
	case LORO_CONTAINER_LIST:
		l, err := c.GetList()
		if err != nil {
			return nil, err
		}
		return l.ToGoObject()
	case LORO_CONTAINER_MOVABLE_LIST:
		l, err := c.GetMovableList()
		if err != nil {
			return nil, err
		}
		return l.ToGoObject()
	case LORO_CONTAINER_COUNTER:
		return nil, pe.WithStack(fmt.Errorf("counter container is not supported"))
	case LORO_CONTAINER_UNKNOWN:
		return nil, pe.WithStack(fmt.Errorf("unknown container type"))
	case LORO_CONTAINER_MAP:
		m, err := c.GetMap()
		if err != nil {
			return nil, err
		}
		return m.ToGoObject()
	case LORO_CONTAINER_TEXT:
		text, err := c.GetText()
		if err != nil {
			return nil, err
		}
		return text.ToString()
	case LORO_CONTAINER_TREE:
		return nil, pe.WithStack(fmt.Errorf("tree container is not supported"))
	}
	return nil, pe.WithStack(fmt.Errorf("unknown container type"))
}

func (c *LoroContainer) Unwrap() (any, error) {
	t := c.GetType()
	switch t {
	case LORO_CONTAINER_MAP:
		m, err := c.GetMap()
		if err != nil {
			return nil, err
		}
		return m, nil
	case LORO_CONTAINER_LIST:
		l, err := c.GetList()
		if err != nil {
			return nil, err
		}
		return l, nil
	case LORO_CONTAINER_MOVABLE_LIST:
		l, err := c.GetMovableList()
		if err != nil {
			return nil, err
		}
		return l, nil
	case LORO_CONTAINER_TEXT:
		text, err := c.GetText()
		if err != nil {
			return nil, err
		}
		return text, nil
	case LORO_CONTAINER_TREE:
		tree, err := c.GetTree()
		if err != nil {
			return nil, err
		}
		return tree, nil
	}
	return nil, pe.WithStack(fmt.Errorf("unknown container type: %d", t))
}

// -------------- Loro Container Value --------------

const (
	LORO_VALUE_TYPE     = 0
	LORO_CONTAINER_TYPE = 1
)

type LoroContainerValueType int32

type LoroContainerOrValue struct {
	ptr unsafe.Pointer
}

func (lv *LoroContainerOrValue) Unwrap() (any, error) {
	t := lv.GetType()
	switch t {
	case LORO_VALUE_TYPE:
		value, err := lv.GetValue()
		if err != nil {
			return nil, err
		}
		return value.Unwrap()
	case LORO_CONTAINER_TYPE:
		container, err := lv.GetContainer()
		if err != nil {
			return nil, err
		}
		return container.Unwrap()
	}
	return nil, pe.WithStack(fmt.Errorf("unknown loro container or value type: %d", t))
}

func (lv *LoroContainerOrValue) Destroy() {
	C.destroy_loro_container_value(lv.ptr)
}

// 0 - value, 1 - container
func (lv *LoroContainerOrValue) GetType() int {
	t := C.loro_container_value_get_type(lv.ptr)
	return int(t)
}

func (lv *LoroContainerOrValue) IsValue() bool {
	return lv.GetType() == LORO_VALUE_TYPE
}

func (lv *LoroContainerOrValue) IsContainer() bool {
	return lv.GetType() == LORO_CONTAINER_TYPE
}

func (lv *LoroContainerOrValue) GetContainer() (*LoroContainer, error) {
	ptr := C.loro_container_value_get_container(lv.ptr)
	if ptr == nil {
		return nil, pe.WithStack(fmt.Errorf("%w: get container from loro container or value", ErrLoroGetFailed))
	}
	container := &LoroContainer{ptr: ptr}
	runtime.SetFinalizer(container, func(container *LoroContainer) {
		container.Destroy()
	})
	return container, nil
}

func (lv *LoroContainerOrValue) GetValue() (*LoroValue, error) {
	ptr := C.loro_container_value_get_value(lv.ptr)
	if ptr == nil {
		return nil, pe.WithStack(fmt.Errorf("%w: get value from loro container or value", ErrLoroGetFailed))
	}
	value := &LoroValue{ptr: ptr}
	runtime.SetFinalizer(value, func(value *LoroValue) {
		value.Destroy()
	})
	return value, nil
}

func (lv *LoroContainerOrValue) ToGoObject() (any, error) {
	t := lv.GetType()
	switch t {
	case LORO_VALUE_TYPE:
		value, err := lv.GetValue()
		if err != nil {
			return nil, err
		}
		return value.ToGoObject()
	case LORO_CONTAINER_TYPE:
		container, err := lv.GetContainer()
		if err != nil {
			return nil, err
		}
		return container.ToGoObject()
	}
	return nil, pe.WithStack(fmt.Errorf("unknown loro container or value type: %d", t))
}

// ----------- Import Blob Meta --------------

const (
	ENCODE_BLOB_SNAPSHOT          = 0
	ENCODE_BLOB_OUTDATED_SNAPSHOT = 1
	ENCODE_BLOB_SHALLOW_SNAPSHOT  = 2
	ENCODE_BLOB_OUTDATED_RLE      = 3
	ENCODE_BLOB_UPDATES           = 4
)

type EncodeBlobMode int32

type ImportBlobMeta struct {
	Mode           EncodeBlobMode
	PartialStartVV *VersionVector
	PartialEndVV   *VersionVector
	StartFrontiers *Frontiers
	StartTimestamp int64
	EndTimestamp   int64
	ChangeNumber   uint32
}

func InspectImport[T *RustBytesVec | []byte](importBlob T, checkChecksum bool) (*ImportBlobMeta, error) {
	var err C.uint8_t
	var psvvPtr unsafe.Pointer
	var pevvPtr unsafe.Pointer
	var sfPtr unsafe.Pointer
	var mode C.uint8_t
	var st C.int64_t
	var et C.int64_t
	var cn C.uint32_t

	var vec *RustBytesVec
	if bytes, ok := any(importBlob).([]byte); ok {
		vec = NewRustBytesVec(bytes)
		defer vec.Destroy()
	} else {
		vec = (any)(importBlob).(*RustBytesVec)
	}

	checkChecksumI := 0
	if checkChecksum {
		checkChecksumI = 1
	}

	C.loro_doc_decode_import_blob_meta(
		vec.ptr,
		C.int32_t(checkChecksumI),
		&err,
		unsafe.Pointer(&psvvPtr),
		unsafe.Pointer(&pevvPtr),
		unsafe.Pointer(&sfPtr),
		&mode,
		&st,
		&et,
		&cn,
	)
	if err != 0 {
		return nil, ErrInspectImportFailed
	}

	psvv := &VersionVector{ptr: psvvPtr}
	runtime.SetFinalizer(psvv, func(psvv *VersionVector) {
		psvv.Destroy()
	})

	pevv := &VersionVector{ptr: pevvPtr}
	runtime.SetFinalizer(pevv, func(pevv *VersionVector) {
		pevv.Destroy()
	})

	sf := &Frontiers{ptr: sfPtr}
	runtime.SetFinalizer(sf, func(sf *Frontiers) {
		sf.Destroy()
	})

	return &ImportBlobMeta{
		Mode:           EncodeBlobMode(mode),
		PartialStartVV: psvv,
		PartialEndVV:   pevv,
		StartFrontiers: sf,
		StartTimestamp: int64(st),
		EndTimestamp:   int64(et),
		ChangeNumber:   uint32(cn),
	}, nil
}
