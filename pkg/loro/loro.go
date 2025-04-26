package loro

/*
#cgo LDFLAGS: -L./loro-c-ffi/target/release -lloro_c_ffi
#include <stdlib.h>
#include "loro-c-ffi/loro_c_ffi.h"
*/
import "C"
import (
	"reflect"
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
	ErrLoroEncodeFailed = pe.New("loro encode failed")
	ErrLoroDecodeFailed = pe.New("loro decode failed")
	ErrLoroGetNull      = pe.New("loro get null")
)

const DATA_MAP_NAME = "root"

func isNil(v any) bool {
	return (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0
}

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

func (text *LoroText) ToString() (string, error) {
	var err C.uint8_t
	ret := C.loro_text_to_string(text.ptr, &err)
	if err != 0 {
		return "", pe.New("failed to convert loro text to string")
	}
	return C.GoString(ret), nil
}

func (text *LoroText) UpdateText(content string) error {
	var err C.uint8_t
	contentPtr := C.CString(content)
	defer C.free(unsafe.Pointer(contentPtr))
	C.update_loro_text(text.ptr, contentPtr, &err)
	if err != 0 {
		return pe.New("failed to update loro text")
	}
	return nil
}

func (text *LoroText) InsertText(content string, pos uint32) error {
	var err C.uint8_t
	contentPtr := C.CString(content)
	defer C.free(unsafe.Pointer(contentPtr))
	C.insert_loro_text(text.ptr, C.uint32_t(pos), contentPtr, &err)
	if err != 0 {
		return pe.New("failed to insert loro text")
	}
	return nil
}

func (text *LoroText) InsertTextUtf8(content string, pos uint32) error {
	var err C.uint8_t
	contentPtr := C.CString(content)
	defer C.free(unsafe.Pointer(contentPtr))
	C.insert_loro_text_utf8(text.ptr, C.uint32_t(pos), contentPtr, &err)
	if err != 0 {
		return pe.New("failed to insert loro text utf8")
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

func (m *LoroMap) ToGoObject() (map[string]any, error) {
	result, err := toGoObject(m)
	if err != nil {
		return nil, err
	}
	return result.(map[string]any), nil
}

// 获取 LoroMap 的长度
func (m *LoroMap) GetLen() uint32 {
	return uint32(C.loro_map_len(m.ptr))
}

// 检查 LoroMap 是否包含指定 key
func (m *LoroMap) Contains(key string) bool {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))
	ptr := C.loro_map_get(m.ptr, cstr)
	return ptr != nil
}

// 从 LoroMap 中获取指定 key 的值。如果 key 不存在，则 ErrLoroGetNull
func (m *LoroMap) Get(key string) (LoroContainerOrValue, error) {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))
	ptr := C.loro_map_get(m.ptr, cstr)
	if ptr == nil {
		return nil, pe.Wrapf(ErrLoroGetNull, "get from map, key=%s", key)
	}
	wrapper := &LoroContainerOrValueWrapper{ptr: unsafe.Pointer(ptr)}
	defer wrapper.Destroy()
	ret, err := wrapper.Unwrap()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// MustGet 获取指定 key 的值。如果 key 不存在，会 panic
func (m *LoroMap) MustGet(key string) LoroContainerOrValue {
	v, err := m.Get(key)
	if err != nil {
		panic(err)
	}
	return v
}

// InsertValue 插入一个 LoroValue 到 LoroMap 中
func (m *LoroMap) InsertValue(key string, value LoroValue) error {
	wrapper, err := WrapLoroValue(value)
	if err != nil {
		return err
	}
	var errCode C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	C.loro_map_insert_value(m.ptr, keyPtr, wrapper.ptr, &errCode)
	if errCode != 0 {
		return pe.Errorf("insert value to map, key=%s", key)
	}
	return nil
}

// InsertValueCoerce 插入一个 any 到 LoroMap 中
//
// 如果 value 不是合法的 LoroValue，会自动尝试转换为 LoroValue
func (m *LoroMap) InsertValueCoerce(key string, value any) error {
	coerced, err := CoerceLoroValue(value)
	if err != nil {
		return err
	}
	return m.InsertValue(key, coerced)
}

// InsertContainer 插入一个 LoroContainer 到 LoroMap 中
//
// 返回插入后，连接到 LoroMap 的 LoroContainer，
// 注意返回的 LoroContainer 和传入的 LoroContainer 不同！
func (m *LoroMap) InsertContainer(key string, container LoroContainer) (LoroContainer, error) {
	wrapper, err := WrapLoroContainer(container)
	if err != nil {
		return nil, err
	}
	defer wrapper.Destroy()
	var errCode C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ptr := C.loro_map_insert_container(m.ptr, keyPtr, wrapper.ptr, &errCode)
	if errCode != 0 {
		return nil, pe.Errorf("insert container to map, key=%s: %s", key)
	}
	wrapper2 := &LoroContainerWrapper{ptr: unsafe.Pointer(ptr)}
	defer wrapper2.Destroy()
	container2, err := wrapper2.Unwrap()
	if err != nil {
		return nil, err
	}
	return container2, nil
}

// IsAttached 检查 LoroMap 是否连接到 LoroDoc
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

// PushValue 追加一个 LoroValue 到 LoroList 中
func (list *LoroList) PushValue(value LoroValue) error {
	wrapper, err := WrapLoroValue(value)
	if err != nil {
		return err
	}
	var errCode C.uint8_t
	C.loro_list_push_value(list.ptr, wrapper.ptr, &errCode)
	if errCode != 0 {
		return pe.New("push value to list")
	}
	return nil
}

// PushValueCoerce 追加一个 any 到 LoroList 中
//
// 如果 value 不是合法的 LoroValue，会自动尝试转换为 LoroValue
func (list *LoroList) PushValueCoerce(value any) error {
	coerced, err := CoerceLoroValue(value)
	if err != nil {
		return err
	}
	return list.PushValue(coerced)
}

// PushContainer 追加一个 LoroContainer 到 LoroList 中
//
// 返回追加后，连接到 LoroList 的 LoroContainer，
// 注意返回的 LoroContainer 和传入的 LoroContainer 不同！
func (list *LoroList) PushContainer(container LoroContainer) (LoroContainer, error) {
	wrapper, err := WrapLoroContainer(container)
	if err != nil {
		return nil, err
	}
	defer wrapper.Destroy()
	var errCode C.uint8_t
	ptr := C.loro_list_push_container(list.ptr, wrapper.ptr, &errCode)
	if errCode != 0 {
		return nil, pe.New("push container to list")
	}
	wrapper2 := &LoroContainerWrapper{ptr: unsafe.Pointer(ptr)}
	defer wrapper2.Destroy()
	container2, err := wrapper2.Unwrap()
	if err != nil {
		return nil, err
	}
	return container2, nil
}

// InsertValue 在指定 index 插入一个 LoroValue
func (list *LoroList) InsertValue(index uint32, value LoroValue) error {
	wrapper, err := WrapLoroValue(value)
	if err != nil {
		return err
	}
	var errCode C.uint8_t
	C.loro_list_insert_value(list.ptr, C.uint32_t(index), wrapper.ptr, &errCode)
	if errCode != 0 {
		return pe.Errorf("insert value to list, index=%d", index)
	}
	return nil
}

// InsertValueCoerce 在指定 index 插入一个 any
//
// 如果 value 不是合法的 LoroValue，会自动尝试转换为 LoroValue
func (list *LoroList) InsertValueCoerce(index uint32, value any) error {
	coerced, err := CoerceLoroValue(value)
	if err != nil {
		return err
	}
	return list.InsertValue(index, coerced)
}

// InsertContainer 在指定 index 插入一个 LoroContainer
//
// 返回插入后，连接到 LoroList 的 LoroContainer，
// 注意返回的 LoroContainer 和传入的 LoroContainer 不同！
func (list *LoroList) InsertContainer(index uint32, container LoroContainer) (LoroContainer, error) {
	wrapper, err := WrapLoroContainer(container)
	if err != nil {
		return nil, err
	}
	defer wrapper.Destroy()
	var errCode C.uint8_t
	ptr := C.loro_list_insert_container(list.ptr, C.uint32_t(index), wrapper.ptr, &errCode)
	if errCode != 0 {
		return nil, pe.Errorf("insert container to list, index=%d", index)
	}
	wrapper2 := &LoroContainerWrapper{ptr: unsafe.Pointer(ptr)}
	defer wrapper2.Destroy()
	container2, err := wrapper2.Unwrap()
	if err != nil {
		return nil, err
	}
	return container2, nil
}

// GetLen 获取 LoroList 的长度
func (list *LoroList) GetLen() uint32 {
	return uint32(C.loro_list_len(list.ptr))
}

// Get 获取指定 index 的值
func (list *LoroList) Get(index uint32) (LoroContainerOrValue, error) {
	ptr := C.loro_list_get(list.ptr, C.uint32_t(index))
	if ptr == nil {
		return nil, pe.Errorf("get from list, index=%d", index)
	}
	wrapper := &LoroContainerOrValueWrapper{ptr: unsafe.Pointer(ptr)}
	defer wrapper.Destroy()
	ret, err := wrapper.Unwrap()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// MustGet 获取指定 index 的值。如果 index 不存在，会 panic
func (list *LoroList) MustGet(index uint32) LoroContainerOrValue {
	v, err := list.Get(index)
	if err != nil {
		panic(err)
	}
	return v
}

// Delete 删除指定 index 的值
//
// pos 是起始 index，count 是删除的个数
func (list *LoroList) Delete(pos uint32, count uint32) error {
	var errCode C.uint8_t
	C.loro_list_delete(list.ptr, C.uint32_t(pos), C.uint32_t(count), &errCode)
	if errCode != 0 {
		return pe.Errorf("delete from list, pos=%d, count=%d", pos, count)
	}
	return nil
}

// Clear 清空 LoroList
func (list *LoroList) Clear() error {
	var errCode C.uint8_t
	C.loro_list_clear(list.ptr, &errCode)
	if errCode != 0 {
		return pe.New("clear list failed")
	}
	return nil
}

// IsAttached 检查 LoroList 是否连接到 LoroDoc
func (list *LoroList) IsAttached() bool {
	return C.loro_list_is_attached(list.ptr) != 0
}

func (list *LoroList) ToGoObject() (any, error) {
	return toGoObject(list)
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

func (list *LoroMovableList) GetLen() uint32 {
	return uint32(C.loro_movable_list_len(list.ptr))
}

// PushValue 追加一个 LoroValue 到 LoroMovableList 中
func (list *LoroMovableList) PushValue(value LoroValue) error {
	wrapper, err := WrapLoroValue(value)
	if err != nil {
		return err
	}
	var errCode C.uint8_t
	C.loro_movable_list_push_value(list.ptr, wrapper.ptr, &errCode)
	if errCode != 0 {
		return pe.New("push value to movable list")
	}
	return nil
}

// PushValueCoerce 追加一个 any 到 LoroMovableList 中
//
// 如果 value 不是合法的 LoroValue，会自动尝试转换为 LoroValue
func (list *LoroMovableList) PushValueCoerce(value any) error {
	coerced, err := CoerceLoroValue(value)
	if err != nil {
		return err
	}
	return list.PushValue(coerced)
}

// PushContainer 追加一个 LoroContainer 到 LoroMovableList 中
//
// 返回追加后，连接到 LoroMovableList 的 LoroContainer，
// 注意返回的 LoroContainer 和传入的 LoroContainer 不同！
func (list *LoroMovableList) PushContainer(container LoroContainer) (LoroContainer, error) {
	wrapper, err := WrapLoroContainer(container)
	if err != nil {
		return nil, err
	}
	defer wrapper.Destroy()
	var errCode C.uint8_t
	ptr := C.loro_movable_list_push_container(list.ptr, wrapper.ptr, &errCode)
	if errCode != 0 {
		return nil, pe.New("push container to movable list")
	}
	wrapper2 := &LoroContainerWrapper{ptr: unsafe.Pointer(ptr)}
	defer wrapper2.Destroy()
	container2, err := wrapper2.Unwrap()
	if err != nil {
		return nil, err
	}
	return container2, nil
}

// InsertValue 在指定 index 插入一个 LoroValue
func (list *LoroMovableList) InsertValue(index uint32, value LoroValue) error {
	wrapper, err := WrapLoroValue(value)
	if err != nil {
		return err
	}
	var errCode C.uint8_t
	C.loro_movable_list_insert_value(list.ptr, C.uint32_t(index), wrapper.ptr, &errCode)
	if errCode != 0 {
		return pe.Errorf("insert value to movable list, index=%d", index)
	}
	return nil
}

// InsertValueCoerce 在指定 index 插入一个 any
//
// 如果 value 不是合法的 LoroValue，会自动尝试转换为 LoroValue
func (list *LoroMovableList) InsertValueCoerce(index uint32, value any) error {
	coerced, err := CoerceLoroValue(value)
	if err != nil {
		return err
	}
	return list.InsertValue(index, coerced)
}

// InsertContainer 在指定 index 插入一个 LoroContainer
//
// 返回插入后，连接到 LoroMovableList 的 LoroContainer，
// 注意返回的 LoroContainer 和传入的 LoroContainer 不同！
func (list *LoroMovableList) InsertContainer(index uint32, container LoroContainer) (LoroContainer, error) {
	wrapper, err := WrapLoroContainer(container)
	if err != nil {
		return nil, err
	}
	defer wrapper.Destroy()
	var errCode C.uint8_t
	ptr := C.loro_movable_list_insert_container(list.ptr, C.uint32_t(index), wrapper.ptr, &errCode)
	if errCode != 0 {
		return nil, pe.Errorf("insert container to movable list, index=%d", index)
	}
	wrapper2 := &LoroContainerWrapper{ptr: unsafe.Pointer(ptr)}
	defer wrapper2.Destroy()
	container2, err := wrapper2.Unwrap()
	if err != nil {
		return nil, err
	}
	return container2, nil
}

// SetValue 设置指定 index 的值
func (list *LoroMovableList) SetValue(index uint32, value LoroValue) error {
	wrapper, err := WrapLoroValue(value)
	if err != nil {
		return err
	}
	var errCode C.uint8_t
	C.loro_movable_list_set_value(list.ptr, C.uint32_t(index), wrapper.ptr, &errCode)
	if errCode != 0 {
		return pe.Errorf("set value to movable list, index=%d", index)
	}
	return nil
}

// SetValueCoerce 设置指定 index 的值
//
// 如果 value 不是合法的 LoroValue，会自动尝试转换为 LoroValue
func (list *LoroMovableList) SetValueCoerce(index uint32, value any) error {
	coerced, err := CoerceLoroValue(value)
	if err != nil {
		return err
	}
	return list.SetValue(index, coerced)
}

// SetContainer 设置指定 index 的值
//
// 返回设置后，连接到 LoroMovableList 的 LoroContainer，
// 注意返回的 LoroContainer 和传入的 LoroContainer 不同！
func (list *LoroMovableList) SetContainer(index uint32, container LoroContainer) (LoroContainer, error) {
	wrapper, err := WrapLoroContainer(container)
	if err != nil {
		return nil, err
	}
	defer wrapper.Destroy()
	var errCode C.uint8_t
	ptr := C.loro_movable_list_set_container(list.ptr, C.uint32_t(index), wrapper.ptr, &errCode)
	if errCode != 0 {
		return nil, pe.Errorf("set container to movable list, index=%d", index)
	}
	wrapper2 := &LoroContainerWrapper{ptr: unsafe.Pointer(ptr)}
	defer wrapper2.Destroy()
	container2, err := wrapper2.Unwrap()
	if err != nil {
		return nil, err
	}
	return container2, nil
}

// Get 获取指定 index 的值
func (list *LoroMovableList) Get(index uint32) (LoroContainerOrValue, error) {
	ptr := C.loro_movable_list_get(list.ptr, C.uint32_t(index))
	if ptr == nil {
		return nil, pe.Errorf("get from movable list, index=%d", index)
	}
	wrapper := &LoroContainerOrValueWrapper{ptr: unsafe.Pointer(ptr)}
	defer wrapper.Destroy()
	ret, err := wrapper.Unwrap()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// MustGet 获取指定 index 的值。如果 index 不存在，会 panic
func (list *LoroMovableList) MustGet(index uint32) LoroContainerOrValue {
	v, err := list.Get(index)
	if err != nil {
		panic(err)
	}
	return v
}

// Delete 删除指定 index 的值
//
// pos 是起始 index，count 是删除的个数
func (list *LoroMovableList) Delete(pos uint32, count uint32) error {
	var errCode C.uint8_t
	C.loro_movable_list_delete(list.ptr, C.uint32_t(pos), C.uint32_t(count), &errCode)
	if errCode != 0 {
		return pe.Errorf("delete from movable list, pos=%d, count=%d", pos, count)
	}
	return nil
}

// Move 移动一个元素
//
// from 是起始 index，to 是目标 index
func (list *LoroMovableList) Move(from uint32, to uint32) error {
	var errCode C.uint8_t
	C.loro_movable_list_move(list.ptr, C.uint32_t(from), C.uint32_t(to), &errCode)
	if errCode != 0 {
		return pe.Errorf("move from movable list, from=%d, to=%d", from, to)
	}
	return nil
}

// Clear 清空 LoroMovableList
func (list *LoroMovableList) Clear() error {
	var errCode C.uint8_t
	C.loro_movable_list_clear(list.ptr, &errCode)
	if errCode != 0 {
		return pe.New("clear movable list")
	}
	return nil
}

func (list *LoroMovableList) IsAttached() bool {
	return C.loro_movable_list_is_attached(list.ptr) != 0
}

func (list *LoroMovableList) ToGoObject() (any, error) {
	return toGoObject(list)
}

// -------------- Loro Tree --------------

type LoroTree struct {
	ptr unsafe.Pointer
}

func (t *LoroTree) Destroy() {
	C.destroy_loro_tree(t.ptr)
}

func (t *LoroTree) IsAttached() bool {
	return C.loro_tree_is_attached(t.ptr) != 0
}

// ----------- Loro Container -----------

type LoroContainer interface {
	isLoroContainer()
}

func (o *LoroMap) isLoroContainer()         {}
func (o *LoroList) isLoroContainer()        {}
func (o *LoroMovableList) isLoroContainer() {}
func (o *LoroText) isLoroContainer()        {}
func (o *LoroTree) isLoroContainer()        {}

func IsLoroContainer(container LoroContainerOrValue) bool {
	switch container.(type) {
	case *LoroMap, *LoroList, *LoroMovableList, *LoroText, *LoroTree:
		return true
	default:
		return false
	}
}

// --------------- Loro Value ---------------

//	  type LoroValue =
//			| nil
//			| bool
//			| float64
//			| int64
//			| string
//			| []byte
//			| []LoroValue
//			| map[string]LoroValue
//			| ContainerId
type LoroValue any

// ------------- Loro Container Or Value -------------

// type LoroContainerOrValue = LoroContainer | LoroValue
type LoroContainerOrValue any

// ------------- Loro Container Wrapper -------------

const (
	LORO_CONTAINER_MAP          = 0
	LORO_CONTAINER_LIST         = 1
	LORO_CONTAINER_TEXT         = 2
	LORO_CONTAINER_TREE         = 3
	LORO_CONTAINER_MOVABLE_LIST = 4
	LORO_CONTAINER_COUNTER      = 5
	LORO_CONTAINER_UNKNOWN      = 6
)

type LoroContainerWrapper struct {
	ptr unsafe.Pointer
}

func (c *LoroContainerWrapper) Destroy() {
	C.destroy_loro_container(c.ptr)
}

func WrapLoroContainer(container LoroContainer) (*LoroContainerWrapper, error) {
	switch v := container.(type) {
	case *LoroText:
		ptr := C.loro_text_to_container(v.ptr)
		wrapper := &LoroContainerWrapper{ptr: ptr}
		runtime.SetFinalizer(wrapper, func(wrapper *LoroContainerWrapper) {
			wrapper.Destroy()
		})
		return wrapper, nil
	case *LoroMap:
		ptr := C.loro_map_to_container(v.ptr)
		wrapper := &LoroContainerWrapper{ptr: ptr}
		runtime.SetFinalizer(wrapper, func(wrapper *LoroContainerWrapper) {
			wrapper.Destroy()
		})
		return wrapper, nil
	case *LoroList:
		ptr := C.loro_list_to_container(v.ptr)
		wrapper := &LoroContainerWrapper{ptr: ptr}
		runtime.SetFinalizer(wrapper, func(wrapper *LoroContainerWrapper) {
			wrapper.Destroy()
		})
		return wrapper, nil
	case *LoroMovableList:
		ptr := C.loro_movable_list_to_container(v.ptr)
		wrapper := &LoroContainerWrapper{ptr: ptr}
		runtime.SetFinalizer(wrapper, func(wrapper *LoroContainerWrapper) {
			wrapper.Destroy()
		})
		return wrapper, nil
	case *LoroTree:
		ptr := C.loro_tree_to_container(v.ptr)
		wrapper := &LoroContainerWrapper{ptr: ptr}
		runtime.SetFinalizer(wrapper, func(wrapper *LoroContainerWrapper) {
			wrapper.Destroy()
		})
		return wrapper, nil
	default:
		return nil, pe.New("failed to wrap loro container, invalid container type")
	}
}

func (c *LoroContainerWrapper) Unwrap() (LoroContainer, error) {
	t := C.loro_container_get_type(c.ptr)
	switch t {
	case LORO_CONTAINER_LIST:
		ptr := C.loro_container_get_list(c.ptr)
		if ptr == nil {
			return nil, pe.Errorf("get list from loro container")
		}
		list := &LoroList{ptr: ptr}
		runtime.SetFinalizer(list, func(list *LoroList) {
			list.Destroy()
		})
		return list, nil
	case LORO_CONTAINER_MAP:
		ptr := C.loro_container_get_map(c.ptr)
		if ptr == nil {
			return nil, pe.Errorf("get map from loro container")
		}
		m := &LoroMap{ptr: ptr}
		runtime.SetFinalizer(m, func(m *LoroMap) {
			m.Destroy()
		})
		return m, nil
	case LORO_CONTAINER_TEXT:
		ptr := C.loro_container_get_text(c.ptr)
		if ptr == nil {
			return nil, pe.Errorf("get text from loro container")
		}
		text := &LoroText{ptr: ptr}
		runtime.SetFinalizer(text, func(text *LoroText) {
			text.Destroy()
		})
		return text, nil
	case LORO_CONTAINER_MOVABLE_LIST:
		ptr := C.loro_container_get_movable_list(c.ptr)
		if ptr == nil {
			return nil, pe.Errorf("get movable list from loro container")
		}
		list := &LoroMovableList{ptr: ptr}
		runtime.SetFinalizer(list, func(list *LoroMovableList) {
			list.Destroy()
		})
		return list, nil
	case LORO_CONTAINER_TREE:
		ptr := C.loro_container_get_tree(c.ptr)
		if ptr == nil {
			return nil, pe.Errorf("get tree from loro container")
		}
		tree := &LoroTree{ptr: ptr}
		runtime.SetFinalizer(tree, func(tree *LoroTree) {
			tree.Destroy()
		})
		return tree, nil
	default:
		return nil, pe.Errorf("unknown loro container type %d", t)
	}
}

// -------------- Loro Container Or Value Rust Wrapper --------------

const (
	LORO_VALUE_TYPE     = 0
	LORO_CONTAINER_TYPE = 1
)

type LoroContainerOrValueWrapper struct {
	ptr unsafe.Pointer
}

func (v *LoroContainerOrValueWrapper) Destroy() {
	C.destroy_loro_container_value(v.ptr)
}

func (v *LoroContainerOrValueWrapper) Unwrap() (LoroContainerOrValue, error) {
	t := C.loro_container_value_get_type(v.ptr)
	switch t {
	case LORO_VALUE_TYPE:
		ptr := C.loro_container_value_get_value(v.ptr)
		wrapper := &LoroValueWrapper{ptr: ptr}
		defer wrapper.Destroy()
		ret, err := wrapper.Unwrap()
		if err != nil {
			return nil, err
		}
		return ret, nil
	case LORO_CONTAINER_TYPE:
		ptr := C.loro_container_value_get_container(v.ptr)
		wrapper := &LoroContainerWrapper{ptr: ptr}
		defer wrapper.Destroy()
		ret, err := wrapper.Unwrap()
		if err != nil {
			return nil, err
		}
		return ret, nil
	default:
		return nil, pe.Errorf("impossible, type must be LORO_VALUE_TYPE or LORO_CONTAINER_TYPE")
	}
}

// --------------- Loro Value Rust Wrapper ---------------

const (
	LORO_VALUE_TYPE_NULL      = 0
	LORO_VALUE_TYPE_BOOL      = 1
	LORO_VALUE_TYPE_DOUBLE    = 2
	LORO_VALUE_TYPE_I64       = 3
	LORO_VALUE_TYPE_STRING    = 4
	LORO_VALUE_TYPE_MAP       = 5
	LORO_VALUE_TYPE_LIST      = 6
	LORO_VALUE_TYPE_BINARY    = 7
	LORO_VALUE_TYPE_CONTAINER = 9
)

// Rust 侧的 LoroValue
type LoroValueWrapper struct {
	ptr unsafe.Pointer
}

func (v *LoroValueWrapper) Destroy() {
	C.destroy_loro_value(v.ptr)
}

func CoerceLoroValue(value any) (LoroValue, error) {
	if isNil(value) {
		return nil, nil
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return util.ToInt64(v), nil
	case float32, float64:
		return util.ToFloat64(v), nil
	case string:
		return v, nil
	case []byte:
		return v, nil
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Map {
			coerced := make(map[string]LoroValue)
			for _, key := range rv.MapKeys() {
				keyStr := key.String()
				value := rv.MapIndex(key)
				coercedValue, err := CoerceLoroValue(value.Interface())
				if err != nil {
					return nil, err
				}
				coerced[keyStr] = coercedValue
			}
			return coerced, nil
		} else if rv.Kind() == reflect.Slice {
			coerced := make([]LoroValue, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				value := rv.Index(i)
				coercedValue, err := CoerceLoroValue(value.Interface())
				if err != nil {
					return nil, err
				}
				coerced[i] = coercedValue
			}
			return coerced, nil
		}
		return nil, pe.New("failed to coerce loro value, invalid value type")
	}
}

func WrapLoroValue(value LoroValue) (*LoroValueWrapper, error) {
	if isNil(value) {
		ptr := C.loro_value_new_null()
		wrapper := &LoroValueWrapper{ptr: ptr}
		runtime.SetFinalizer(wrapper, func(wrapper *LoroValueWrapper) {
			wrapper.Destroy()
		})
		return wrapper, nil
	}

	switch v := value.(type) {
	case bool:
		boolValue := 0
		if v {
			boolValue = 1
		}
		ptr := C.loro_value_new_bool(C.int(boolValue))
		wrapper := &LoroValueWrapper{ptr: ptr}
		runtime.SetFinalizer(wrapper, func(wrapper *LoroValueWrapper) {
			wrapper.Destroy()
		})
		return wrapper, nil
	case float64:
		ptr := C.loro_value_new_double(C.double(v))
		wrapper := &LoroValueWrapper{ptr: ptr}
		runtime.SetFinalizer(wrapper, func(wrapper *LoroValueWrapper) {
			wrapper.Destroy()
		})
		return wrapper, nil
	case int64:
		ptr := C.loro_value_new_i64(C.int64_t(v))
		wrapper := &LoroValueWrapper{ptr: ptr}
		runtime.SetFinalizer(wrapper, func(wrapper *LoroValueWrapper) {
			wrapper.Destroy()
		})
		return wrapper, nil
	case string:
		ptr := C.loro_value_new_string(C.CString(v))
		wrapper := &LoroValueWrapper{ptr: ptr}
		runtime.SetFinalizer(wrapper, func(wrapper *LoroValueWrapper) {
			wrapper.Destroy()
		})
		return wrapper, nil
	case []byte:
		bytesVec := NewRustBytesVec(v)
		defer bytesVec.Destroy()
		ptr := C.loro_value_new_binary(bytesVec.ptr)
		wrapper := &LoroValueWrapper{ptr: ptr}
		runtime.SetFinalizer(wrapper, func(wrapper *LoroValueWrapper) {
			wrapper.Destroy()
		})
		return wrapper, nil
	case map[string]LoroValue:
		kvVec := NewRustPtrVec()
		defer kvVec.Destroy()
		for key, value := range v {
			keyPtr := C.CString(key)
			defer C.free(unsafe.Pointer(keyPtr))
			kvVec.Push(unsafe.Pointer(keyPtr))

			wrapped, err := WrapLoroValue(value)
			if err != nil {
				return nil, err
			}
			defer wrapped.Destroy()
			kvVec.Push(wrapped.ptr)
		}
		ptr := C.loro_value_new_map(kvVec.ptr)
		wrapper := &LoroValueWrapper{ptr: ptr}
		runtime.SetFinalizer(wrapper, func(wrapper *LoroValueWrapper) {
			wrapper.Destroy()
		})
		return wrapper, nil
	case []LoroValue:
		valueVec := NewRustPtrVec()
		defer valueVec.Destroy()
		for _, value := range v {
			wrapped, err := WrapLoroValue(value)
			if err != nil {
				return nil, err
			}
			defer wrapped.Destroy()
			valueVec.Push(wrapped.ptr)
		}
		ptr := C.loro_value_new_list(valueVec.ptr)
		wrapper := &LoroValueWrapper{ptr: ptr}
		runtime.SetFinalizer(wrapper, func(wrapper *LoroValueWrapper) {
			wrapper.Destroy()
		})
		return wrapper, nil
	default:
		return nil, pe.New("failed to wrap loro value, invalid value type")
	}
}

func (v *LoroValueWrapper) Unwrap() (LoroValue, error) {
	valueType := C.loro_value_get_type(v.ptr)
	switch valueType {
	case LORO_VALUE_TYPE_NULL:
		return nil, nil
	case LORO_VALUE_TYPE_BOOL:
		var err C.uint8_t
		ret := C.loro_value_get_bool(v.ptr, &err)
		if err != 0 {
			return false, pe.Errorf("get bool from loro value")
		}
		return ret != 0, nil
	case LORO_VALUE_DOUBLE:
		var err C.uint8_t
		ret := C.loro_value_get_double(v.ptr, &err)
		if err != 0 {
			return 0, pe.Errorf("get double from loro value")
		}
		return float64(ret), nil
	case LORO_VALUE_TYPE_I64:
		var err C.uint8_t
		ret := C.loro_value_get_i64(v.ptr, &err)
		if err != 0 {
			return 0, pe.Errorf("get i64 from loro value")
		}
		return int64(ret), nil
	case LORO_VALUE_TYPE_STRING:
		var err C.uint8_t
		ret := C.loro_value_get_string(v.ptr, &err)
		if err != 0 {
			return "", pe.Errorf("get string from loro value")
		}
		return C.GoString(ret), nil
	case LORO_VALUE_TYPE_MAP:
		var err C.uint8_t
		ptr := C.loro_value_get_map(v.ptr, &err)
		if err != 0 {
			return nil, pe.Errorf("get map from loro value")
		}
		ptrVec := &RustPtrVec{ptr: ptr}
		defer ptrVec.Destroy()
		items := make(map[string]LoroValue)
		ptrVecLen := int(ptrVec.GetLen())
		ptrVecData := ptrVec.GetData()
		for i := 0; i < ptrVecLen; i += 2 {
			keyPtr := ptrVecData[i]
			valPtr := ptrVecData[i+1]
			key := C.GoString((*C.char)(keyPtr))
			valWrapper := &LoroValueWrapper{ptr: valPtr}
			defer valWrapper.Destroy()
			val, err := valWrapper.Unwrap()
			if err != nil {
				return nil, err
			}
			items[key] = val
		}
		return items, nil
	case LORO_VALUE_TYPE_LIST:
		var err C.uint8_t
		ptr := C.loro_value_get_list(v.ptr, &err)
		if err != 0 {
			return nil, pe.Errorf("get list from loro value")
		}
		ptrVec := &RustPtrVec{ptr: ptr}
		defer ptrVec.Destroy()
		items := make([]LoroValue, ptrVec.GetLen())
		for i, itemPtr := range ptrVec.GetData() {
			itemWrapper := &LoroValueWrapper{ptr: itemPtr}
			defer itemWrapper.Destroy()
			item, err := itemWrapper.Unwrap()
			if err != nil {
				return nil, err
			}
			items[i] = item
		}
		return items, nil
	case LORO_VALUE_TYPE_BINARY:
		var err C.uint8_t
		ptr := C.loro_value_get_binary(v.ptr, &err)
		if err != 0 {
			return nil, pe.Errorf("get binary from loro value")
		}
		bytesVec := &RustBytesVec{ptr: ptr}
		defer bytesVec.Destroy()
		bytes := bytesVec.Bytes()
		bytesCloned := make([]byte, len(bytes))
		copy(bytesCloned, bytes)
		return bytesCloned, nil
	case LORO_VALUE_TYPE_CONTAINER:
		var err C.uint8_t
		ptr := C.loro_value_get_container_id(v.ptr, &err)
		if err != 0 {
			return nil, pe.Errorf("get container id from loro value")
		}
		cid := &ContainerId{ptr: ptr}
		runtime.SetFinalizer(cid, func(cid *ContainerId) {
			cid.Destroy()
		})
		return cid, nil
	default:
		return nil, pe.Errorf("unknown loro value type %d", valueType)
	}
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

// func (li *ListDiffItem) GetInserted() ([]*LoroContainerOrValue, bool, error) {
// 	var err C.uint8_t
// 	var is_move C.uint8_t
// 	inserted := C.list_diff_item_get_insert(li.ptr, &is_move, &err)
// 	if err != 0 {
// 		return nil, false, pe.WithStack(fmt.Errorf("%w: get inserted from list diff item", ErrLoroGetFailed))
// 	}
// 	vec := &RustPtrVec{ptr: inserted}
// 	items := make([]*LoroContainerOrValue, vec.GetLen())
// 	for i, itemPtr := range vec.GetData() {
// 		item := &LoroContainerOrValue{ptr: itemPtr}
// 		runtime.SetFinalizer(item, func(item *LoroContainerOrValue) {
// 			item.Destroy()
// 		})
// 		items[i] = item
// 	}
// 	vec.Destroy()
// 	return items, is_move == 1, nil
// }

func (li *ListDiffItem) GetDeleteCount() (uint32, error) {
	var err C.uint8_t
	count := C.list_diff_item_get_delete_count(li.ptr, &err)
	if err != 0 {
		return 0, pe.Errorf("get delete count from list diff item")
	}
	return uint32(count), nil
}

func (li *ListDiffItem) GetRetainCount() (uint32, error) {
	var err C.uint8_t
	count := C.list_diff_item_get_retain_count(li.ptr, &err)
	if err != 0 {
		return 0, pe.Errorf("get retain count from list diff item")
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
		return nil, pe.New("inspect import blob")
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

///////////// To Go Object /////////////

func toGoObject(val LoroContainerOrValue) (any, error) {
	switch v := val.(type) {
	case *LoroMap:
		vecPtr := C.loro_map_get_items(v.ptr)
		vec := &RustPtrVec{ptr: unsafe.Pointer(vecPtr)}
		defer vec.Destroy()
		items := vec.GetData()
		result := make(map[string]any, len(items))
		vecLen := vec.GetLen()
		for i := uint32(0); i < vecLen; i += 2 {
			keyPtr := items[i]
			valPtr := items[i+1]
			key := C.GoString((*C.char)(keyPtr))
			valWrapper := &LoroContainerOrValueWrapper{ptr: valPtr}
			defer valWrapper.Destroy()
			val, err := valWrapper.Unwrap()
			if err != nil {
				return nil, err
			}
			valGo, err := toGoObject(val)
			if err != nil {
				return nil, err
			}
			result[key] = valGo
		}
		return result, nil
	case *LoroList:
		vecPtr := C.loro_list_get_items(v.ptr)
		vec := &RustPtrVec{ptr: unsafe.Pointer(vecPtr)}
		defer vec.Destroy()
		items := vec.GetData()
		result := make([]any, len(items))
		for i, ptr := range items {
			itemWrapper := &LoroContainerOrValueWrapper{ptr: unsafe.Pointer(ptr)}
			defer itemWrapper.Destroy()
			item, err := itemWrapper.Unwrap()
			if err != nil {
				return nil, err
			}
			itemGo, err := toGoObject(item)
			if err != nil {
				return nil, err
			}
			result[i] = itemGo
		}
		return result, nil
	case *LoroMovableList:
		vecPtr := C.loro_movable_list_get_items(v.ptr)
		vec := &RustPtrVec{ptr: unsafe.Pointer(vecPtr)}
		defer vec.Destroy()
		items := vec.GetData()
		result := make([]any, len(items))
		for i, ptr := range items {
			itemWrapper := &LoroContainerOrValueWrapper{ptr: unsafe.Pointer(ptr)}
			defer itemWrapper.Destroy()
			item, err := itemWrapper.Unwrap()
			if err != nil {
				return nil, err
			}
			itemGo, err := toGoObject(item)
			if err != nil {
				return nil, err
			}
			result[i] = itemGo
		}
		return result, nil
	case *LoroText:
		return v.ToString()
	case *LoroTree:
		panic("not implemented")
	case bool, int64, float64, string, []byte:
		return v, nil
	case []LoroContainerOrValue:
		result := make([]any, len(v))
		for i, item := range v {
			itemGo, err := toGoObject(item)
			if err != nil {
				return nil, err
			}
			result[i] = itemGo
		}
		return result, nil
	case map[string]LoroContainerOrValue:
		result := make(map[string]any, len(v))
		for key, item := range v {
			itemGo, err := toGoObject(item)
			if err != nil {
				return nil, err
			}
			result[key] = itemGo
		}
		return result, nil
	default:
		return v, nil
	}
}
