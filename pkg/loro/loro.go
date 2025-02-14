package loro

/*
#cgo LDFLAGS: -L./loro-c-ffi/target/release -lloro_c_ffi
#include <stdlib.h>
#include "loro-c-ffi/loro_c_ffi.h"
*/
import "C"
import (
	"errors"
	"math"
	"runtime"
	"unsafe"
)

const LORO_VALUE_NULL = 0
const LORO_VALUE_BOOL = 1
const LORO_VALUE_DOUBLE = 2
const LORO_VALUE_I64 = 3
const LORO_VALUE_BINARY = 4
const LORO_VALUE_STRING = 5
const LORO_VALUE_LIST = 6
const LORO_VALUE_MAP = 7
const LORO_VALUE_CONTAINER = 8

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
	// fmt.Println("destroying rust bytes vec")
	C.destroy_bytes_vec(vec.ptr)
}

func (vec *RustBytesVec) GetLen() uint32 {
	return uint32(C.get_vec_len(vec.ptr))
}

func (vec *RustBytesVec) GetCapacity() uint32 {
	return uint32(C.get_vec_cap(vec.ptr))
}

func (vec *RustBytesVec) Bytes() []byte {
	len := vec.GetLen()
	dataPtr := C.get_vec_data(vec.ptr)
	return unsafe.Slice((*byte)(dataPtr), len)
}

// ----------- Rust Ptr Vec ----------

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
	// fmt.Println("destroying rust bytes vec")
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
	// fmt.Println("destroying loro doc")
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
		return "", ErrFailedToConvertLoroText
	}
	return C.GoString(ret), nil
}

func (text *LoroText) UpdateText(content string) error {
	var err C.uint8_t
	contentPtr := C.CString(content)
	defer C.free(unsafe.Pointer(contentPtr))
	C.update_loro_text(text.ptr, contentPtr, &err)
	if err != 0 {
		return ErrFailedToUpdateLoroText
	}
	return nil
}

func (text *LoroText) InsertText(content string, pos uint32) error {
	var err C.uint8_t
	contentPtr := C.CString(content)
	defer C.free(unsafe.Pointer(contentPtr))
	C.insert_loro_text(text.ptr, C.uint32_t(pos), contentPtr, &err)
	if err != 0 {
		return ErrFailedToInsertLoroText
	}
	return nil
}

func (text *LoroText) InsertTextUtf8(content string, pos uint32) error {
	var err C.uint8_t
	contentPtr := C.CString(content)
	defer C.free(unsafe.Pointer(contentPtr))
	C.insert_loro_text_utf8(text.ptr, C.uint32_t(pos), contentPtr, &err)
	if err != 0 {
		return ErrFailedToInsertLoroTextUtf8
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

var (
	ErrFailedToGetMapNull           = errors.New("failed to get null from map")
	ErrFailedToGetMapBool           = errors.New("failed to get bool from map")
	ErrFailedToGetMapDouble         = errors.New("failed to get double from map")
	ErrFailedToGetMapI64            = errors.New("failed to get i64 from map")
	ErrFailedToGetMapString         = errors.New("failed to get string from map")
	ErrFailedToGetMapText           = errors.New("failed to get text from map")
	ErrFailedToGetMapList           = errors.New("failed to get list from map")
	ErrFailedToGetMapMovableList    = errors.New("failed to get movable list from map")
	ErrFailedToGetMapMap            = errors.New("failed to get map from map")
	ErrFailedToInsertMapNull        = errors.New("failed to insert null to map")
	ErrFailedToInsertMapBool        = errors.New("failed to insert bool to map")
	ErrFailedToInsertMapDouble      = errors.New("failed to insert double to map")
	ErrFailedToInsertMapI64         = errors.New("failed to insert i64 to map")
	ErrFailedToInsertMapString      = errors.New("failed to insert string to map")
	ErrFailedToInsertMapText        = errors.New("failed to insert text to map")
	ErrFailedToInsertMapList        = errors.New("failed to insert list to map")
	ErrFailedToInsertMapMovableList = errors.New("failed to insert movable list to map")
	ErrFailedToInsertMapMap         = errors.New("failed to insert map to map")
)

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
		return ErrFailedToGetMapNull
	}
	return nil
}

func (m *LoroMap) GetBool(key string) (bool, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_get_bool(m.ptr, keyPtr, &err)
	if err != 0 {
		return false, ErrFailedToGetMapBool
	}
	return ret != 0, nil
}

func (m *LoroMap) GetDouble(key string) (float64, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_get_double(m.ptr, keyPtr, &err)
	if err != 0 {
		return 0, ErrFailedToGetMapDouble
	}
	return float64(ret), nil
}

func (m *LoroMap) GetI64(key string) (int64, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_get_i64(m.ptr, keyPtr, &err)
	if err != 0 {
		return 0, ErrFailedToGetMapI64
	}
	return int64(ret), nil
}

func (m *LoroMap) GetString(key string) (string, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_get_string(m.ptr, keyPtr, &err)
	if err != 0 {
		return "", ErrFailedToGetMapString
	}
	return C.GoString(ret), nil
}

func (m *LoroMap) GetText(key string) (*LoroText, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_get_text(m.ptr, keyPtr, &err)
	if err != 0 {
		return nil, ErrFailedToGetMapText
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
		return nil, ErrFailedToGetMapList
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
		return nil, ErrFailedToGetMapMovableList
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
		return nil, ErrFailedToGetMapMap
	}
	newMap := &LoroMap{ptr: unsafe.Pointer(ret)}
	runtime.SetFinalizer(newMap, func(m *LoroMap) {
		m.Destroy()
	})
	return newMap, nil
}

func (m *LoroMap) InsertNull(key string) error {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	C.loro_map_insert_null(m.ptr, keyPtr, &err)
	if err != 0 {
		return ErrFailedToInsertMapNull
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
		return ErrFailedToInsertMapBool
	}
	return nil
}

func (m *LoroMap) InsertDouble(key string, value float64) error {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	C.loro_map_insert_double(m.ptr, keyPtr, C.double(value), &err)
	if err != 0 {
		return ErrFailedToInsertMapDouble
	}
	return nil
}

func (m *LoroMap) InsertI64(key string, value int64) error {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	C.loro_map_insert_i64(m.ptr, keyPtr, C.int64_t(value), &err)
	if err != 0 {
		return ErrFailedToInsertMapI64
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
		return ErrFailedToInsertMapString
	}
	return nil
}

func (m *LoroMap) InsertText(key string, text *LoroText) (*LoroText, error) {
	var err C.uint8_t
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))
	ret := C.loro_map_insert_text(m.ptr, keyPtr, text.ptr, &err)
	if err != 0 {
		return nil, ErrFailedToInsertMapText
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
		return nil, ErrFailedToInsertMapList
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
		return nil, ErrFailedToInsertMapMovableList
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
		return nil, ErrFailedToInsertMapMap
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

var ErrFailedToPushNull = errors.New("failed to push null")
var ErrFailedToPushBool = errors.New("failed to push bool")
var ErrFailedToPushDouble = errors.New("failed to push double")
var ErrFailedToPushI64 = errors.New("failed to push i64")
var ErrFailedToPushString = errors.New("failed to push string")
var ErrFailedToPushText = errors.New("failed to push text")
var ErrFailedToPushList = errors.New("failed to push list")
var ErrFailedToPushMovableList = errors.New("failed to push movable list")
var ErrFailedToPushMap = errors.New("failed to push map")
var ErrFailedToGetNull = errors.New("failed to get null")
var ErrFailedToGetBool = errors.New("failed to get bool")
var ErrFailedToGetDouble = errors.New("failed to get double")
var ErrFailedToGetI64 = errors.New("failed to get i64")
var ErrFailedToGetString = errors.New("failed to get string")
var ErrFailedToGetText = errors.New("failed to get text")
var ErrFailedToList = errors.New("failed to get list")
var ErrFailedToGetMovableList = errors.New("failed to get movable list")
var ErrFailedToGetMap = errors.New("failed to get map")

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

func (list *LoroList) PushNull() (interface{}, error) {
	var err C.uint8_t
	C.loro_list_push_null(list.ptr, &err)
	if err != 0 {
		return nil, ErrFailedToPushNull
	}
	return nil, nil
}

func (list *LoroList) PushBool(value bool) (bool, error) {
	var err C.uint8_t
	boolValue := 0
	if value {
		boolValue = 1
	}
	C.loro_list_push_bool(list.ptr, C.int32_t(boolValue), &err)
	if err != 0 {
		return false, ErrFailedToPushBool
	}
	return value, nil
}

func (list *LoroList) PushDouble(value float64) (float64, error) {
	var err C.uint8_t
	C.loro_list_push_double(list.ptr, C.double(value), &err)
	if err != 0 {
		return math.NaN(), ErrFailedToPushDouble
	}
	return value, nil
}

func (list *LoroList) PushI64(value int64) (int64, error) {
	var err C.uint8_t
	C.loro_list_push_i64(list.ptr, C.int64_t(value), &err)
	if err != 0 {
		return -1, ErrFailedToPushI64
	}
	return value, nil
}

func (list *LoroList) PushString(value string) (string, error) {
	var err C.uint8_t
	C.loro_list_push_string(list.ptr, C.CString(value), &err)
	if err != 0 {
		return "", ErrFailedToPushString
	}
	return value, nil
}

func (list *LoroList) PushText(textValue *LoroText) (*LoroText, error) {
	var err C.uint8_t
	ptr := C.loro_list_push_text(list.ptr, textValue.ptr, &err)
	if err != 0 {
		return nil, ErrFailedToPushText
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
		return nil, ErrFailedToPushList
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
		return nil, ErrFailedToPushMovableList
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
		return nil, ErrFailedToPushMap
	}
	newMap := &LoroMap{ptr: ptr}
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
		return ErrFailedToGetNull
	}
	return nil
}

func (list *LoroList) GetBool(index uint32) (bool, error) {
	var err C.uint8_t
	ret := C.loro_list_get_bool(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return false, ErrFailedToGetBool
	}
	return ret != 0, nil
}

func (list *LoroList) GetDouble(index uint32) (float64, error) {
	var err C.uint8_t
	ret := C.loro_list_get_double(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return 0, ErrFailedToGetDouble
	}
	return float64(ret), nil
}

func (list *LoroList) GetI64(index uint32) (int64, error) {
	var err C.uint8_t
	ret := C.loro_list_get_i64(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return 0, ErrFailedToGetI64
	}
	return int64(ret), nil
}

func (list *LoroList) GetString(index uint32) (string, error) {
	var err C.uint8_t
	ret := C.loro_list_get_string(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return "", ErrFailedToGetString
	}
	return C.GoString(ret), nil
}

func (list *LoroList) GetText(index uint32) (*LoroText, error) {
	var err C.uint8_t
	ret := C.loro_list_get_text(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return nil, ErrFailedToGetText
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
		return nil, ErrFailedToList
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
		return nil, ErrFailedToGetMovableList
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
		return nil, ErrFailedToGetMap
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

// ----------- Loro Movable List -----------
var (
	ErrFailedToPushMovableListNull        = errors.New("failed to push null to movable list")
	ErrFailedToPushMovableListBool        = errors.New("failed to push bool to movable list")
	ErrFailedToPushMovableListDouble      = errors.New("failed to push double to movable list")
	ErrFailedToPushMovableListI64         = errors.New("failed to push i64 to movable list")
	ErrFailedToPushMovableListString      = errors.New("failed to push string to movable list")
	ErrFailedToPushMovableListText        = errors.New("failed to push text to movable list")
	ErrFailedToPushMovableListList        = errors.New("failed to push list to movable list")
	ErrFailedToPushMovableListMovableList = errors.New("failed to push movable list to movable list")
	ErrFailedToPushMovableListMap         = errors.New("failed to push map to movable list")
	ErrFailedToGetMovableListNull         = errors.New("failed to get null from movable list")
	ErrFailedToGetMovableListBool         = errors.New("failed to get bool from movable list")
	ErrFailedToGetMovableListDouble       = errors.New("failed to get double from movable list")
	ErrFailedToGetMovableListI64          = errors.New("failed to get i64 from movable list")
	ErrFailedToGetMovableListString       = errors.New("failed to get string from movable list")
	ErrFailedToGetMovableListText         = errors.New("failed to get text from movable list")
	ErrFailedToGetMovableListList         = errors.New("failed to get list from movable list")
	ErrFailedToGetMovableListMovableList  = errors.New("failed to get movable list from movable list")
	ErrFailedToGetMovableListMap          = errors.New("failed to get map from movable list")
)

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

func (list *LoroMovableList) PushNull() (interface{}, error) {
	var err C.uint8_t
	C.loro_movable_list_push_null(list.ptr, &err)
	if err != 0 {
		return nil, ErrFailedToPushMovableListNull
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
		return false, ErrFailedToPushMovableListBool
	}
	return value, nil
}

func (list *LoroMovableList) PushDouble(value float64) (float64, error) {
	var err C.uint8_t
	C.loro_movable_list_push_double(list.ptr, C.double(value), &err)
	if err != 0 {
		return math.NaN(), ErrFailedToPushMovableListDouble
	}
	return value, nil
}

func (list *LoroMovableList) PushI64(value int64) (int64, error) {
	var err C.uint8_t
	C.loro_movable_list_push_i64(list.ptr, C.int64_t(value), &err)
	if err != 0 {
		return -1, ErrFailedToPushMovableListI64
	}
	return value, nil
}

func (list *LoroMovableList) PushString(value string) (string, error) {
	var err C.uint8_t
	C.loro_movable_list_push_string(list.ptr, C.CString(value), &err)
	if err != 0 {
		return "", ErrFailedToPushMovableListString
	}
	return value, nil
}

func (list *LoroMovableList) PushText(textValue *LoroText) (*LoroText, error) {
	var err C.uint8_t
	ptr := C.loro_movable_list_push_text(list.ptr, textValue.ptr, &err)
	if err != 0 {
		return nil, ErrFailedToPushMovableListText
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
		return nil, ErrFailedToPushMovableListList
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
		return nil, ErrFailedToPushMovableListMovableList
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
		return nil, ErrFailedToPushMovableListMap
	}
	newMap := &LoroMap{ptr: ptr}
	runtime.SetFinalizer(newMap, func(mapValue *LoroMap) {
		mapValue.Destroy()
	})
	return newMap, nil
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
		return ErrFailedToGetMovableListNull
	}
	return nil
}

func (list *LoroMovableList) GetBool(index uint32) (bool, error) {
	var err C.uint8_t
	ret := C.loro_movable_list_get_bool(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return false, ErrFailedToGetMovableListBool
	}
	return ret != 0, nil
}

func (list *LoroMovableList) GetDouble(index uint32) (float64, error) {
	var err C.uint8_t
	ret := C.loro_movable_list_get_double(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return 0, ErrFailedToGetMovableListDouble
	}
	return float64(ret), nil
}

func (list *LoroMovableList) GetI64(index uint32) (int64, error) {
	var err C.uint8_t
	ret := C.loro_movable_list_get_i64(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return 0, ErrFailedToGetMovableListI64
	}
	return int64(ret), nil
}

func (list *LoroMovableList) GetString(index uint32) (string, error) {
	var err C.uint8_t
	ret := C.loro_movable_list_get_string(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return "", ErrFailedToGetMovableListString
	}
	return C.GoString(ret), nil
}

func (list *LoroMovableList) GetText(index uint32) (*LoroText, error) {
	var err C.uint8_t
	ret := C.loro_movable_list_get_text(list.ptr, C.uint32_t(index), &err)
	if err != 0 {
		return nil, ErrFailedToGetMovableListText
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
		return nil, ErrFailedToGetMovableListList
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
		return nil, ErrFailedToGetMovableListMovableList
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
		return nil, ErrFailedToGetMovableListMap
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

var (
	ErrFailedGetInsertedThings = errors.New("failed to get inserted things")
	ErrFailedGetDeleteCount    = errors.New("failed to get delete count")
	ErrFailedGetRetainCount    = errors.New("failed to get retain count")
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
		return nil, false, ErrFailedGetInsertedThings
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
		return 0, ErrFailedGetDeleteCount
	}
	return uint32(count), nil
}

func (li *ListDiffItem) GetRetainCount() (uint32, error) {
	var err C.uint8_t
	count := C.list_diff_item_get_retain_count(li.ptr, &err)
	if err != 0 {
		return 0, ErrFailedGetRetainCount
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

type LoroValueOrContainerType int32

const (
	LORO_NULL_VALUE      LoroValueOrContainerType = 0
	LORO_BOOL_VALUE      LoroValueOrContainerType = 1
	LORO_DOUBLE_VALUE    LoroValueOrContainerType = 2
	LORO_I64_VALUE       LoroValueOrContainerType = 3
	LORO_STRING_VALUE    LoroValueOrContainerType = 4
	LORO_MAP_VALUE       LoroValueOrContainerType = 5
	LORO_LIST_VALUE      LoroValueOrContainerType = 6
	LORO_BINARY_VALUE    LoroValueOrContainerType = 7
	LORO_CONTAINER_VALUE LoroValueOrContainerType = 8
	LORO_CONTAINER_ID    LoroValueOrContainerType = 9
)

var (
	ErrGetLoroValue           = errors.New("failed to get loro value")
	ErrFailedConvertLoroValue = errors.New("failed to convert loro value")
	ErrDumpJson               = errors.New("failed to get loro value json")
)

type LoroValue struct {
	ptr unsafe.Pointer
}

func NewLoroValueFromJson(json string) (*LoroValue, error) {
	ptr := C.loro_value_from_json(C.CString(json))
	if ptr == nil {
		return nil, ErrGetLoroValue
	}
	lv := &LoroValue{ptr: ptr}
	runtime.SetFinalizer(lv, func(lv *LoroValue) {
		lv.Destroy()
	})
	return lv, nil
}

func NewLoroValue(value interface{}) (*LoroValue, error) {
	switch v := value.(type) {
	case *LoroValue:
		return v, nil
	case nil:
		return NewLoroValueNull(), nil
	case bool:
		return NewLoroValueBool(v), nil
	case float32, float64:
		return NewLoroValueDouble(ToFloat64(v)), nil
	case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
		return NewLoroValueI64(ToInt64(v)), nil
	case string:
		return NewLoroValueString(v), nil
	case []byte:
		return NewLoroValueBinary(v), nil
	case []interface{}:
		l := make([]*LoroValue, len(v))
		for i, v := range v {
			lv, err := NewLoroValue(v)
			if err != nil {
				return nil, err
			}
			l[i] = lv
		}
		return NewLoroValueList(l), nil
	case map[string]interface{}:
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
		return nil, ErrFailedConvertLoroValue
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

func (lv *LoroValue) GetType() LoroValueOrContainerType {
	t := C.loro_value_get_type(lv.ptr)
	return LoroValueOrContainerType(t)
}

func (lv *LoroValue) GetBool() (bool, error) {
	var err C.uint8_t
	ret := C.loro_value_get_bool(lv.ptr, &err)
	if err != 0 {
		return false, ErrGetLoroValue
	}
	return ret != 0, nil
}

func (lv *LoroValue) GetDouble() (float64, error) {
	var err C.uint8_t
	ret := C.loro_value_get_double(lv.ptr, &err)
	if err != 0 {
		return 0, ErrGetLoroValue
	}
	return float64(ret), nil
}

func (lv *LoroValue) GetI64() (int64, error) {
	var err C.uint8_t
	ret := C.loro_value_get_i64(lv.ptr, &err)
	if err != 0 {
		return 0, ErrGetLoroValue
	}
	return int64(ret), nil
}

func (lv *LoroValue) GetString() (string, error) {
	var err C.uint8_t
	ret := C.loro_value_get_string(lv.ptr, &err)
	if err != 0 {
		return "", ErrGetLoroValue
	}
	return C.GoString(ret), nil
}

func (lv *LoroValue) GetMap() (map[string]*LoroValue, error) {
	var err C.uint8_t
	ptr := C.loro_value_get_map(lv.ptr, &err)
	if err != 0 {
		return nil, ErrGetLoroValue
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

func (lv *LoroValue) GetMapDeep() (map[string]interface{}, error) {
	mapValue, err := lv.GetMap()
	if err != nil {
		return nil, err
	}
	newMapValue := make(map[string]interface{})
	for k, v := range mapValue {
		lv, err := NewLoroValue(v)
		if err != nil {
			return nil, err
		}
		newMapValue[k] = lv
	}
	return newMapValue, nil
}

func (lv *LoroValue) GetList() ([]*LoroValue, error) {
	var err C.uint8_t
	ptr := C.loro_value_get_list(lv.ptr, &err)
	if err != 0 {
		return nil, ErrGetLoroValue
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

func (lv *LoroValue) GetListDeep() ([]interface{}, error) {
	listValue, err := lv.GetList()
	if err != nil {
		return nil, err
	}
	newListValue := make([]interface{}, len(listValue))
	for i, v := range listValue {
		lv, err := NewLoroValue(v)
		if err != nil {
			return nil, err
		}
		newListValue[i] = lv
	}
	return newListValue, nil
}

func (lv *LoroValue) GetBinary() (*RustBytesVec, error) {
	var err C.uint8_t
	ptr := C.loro_value_get_binary(lv.ptr, &err)
	if err != 0 {
		return nil, ErrGetLoroValue
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
		return nil, ErrGetLoroValue
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
		return "", ErrDumpJson
	}
	return C.GoString(ptr), nil
}

// -------------- Loro Container --------------

const (
	LORO_CONTAINER_LIST         = 0
	LORO_CONTAINER_MAP          = 1
	LORO_CONTAINER_TEXT         = 2
	LORO_CONTAINER_MOVABLE_LIST = 3
	LORO_CONTAINER_TREE         = 4
	LORO_CONTAINER_UNKNOWN      = 5
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
		return nil, ErrGetLoroValue
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
		return nil, ErrGetLoroValue
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
		return nil, ErrGetLoroValue
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
		return nil, ErrGetLoroValue
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
		return nil, ErrGetLoroValue
	}
	tree := &LoroTree{ptr: ptr}
	runtime.SetFinalizer(tree, func(tree *LoroTree) {
		tree.Destroy()
	})
	return tree, nil
}

// -------------- Loro Container Value --------------

const (
	LORO_VALUE_TYPE     = 0
	LORO_CONTAINER_TYPE = 1
)

var (
	ErrFailedGetValue     = errors.New("failed to get loro value")
	ErrFailedGetContainer = errors.New("failed to get loro container")
)

type LoroContainerValueType int32

type LoroContainerOrValue struct {
	ptr unsafe.Pointer
}

func (lv *LoroContainerOrValue) Destroy() {
	C.destroy_loro_container_value(lv.ptr)
}

func (lv *LoroContainerOrValue) GetType() LoroContainerType {
	t := C.loro_container_value_get_type(lv.ptr)
	return LoroContainerType(t)
}

func (lv *LoroContainerOrValue) GetContainer() (*LoroContainer, error) {
	ptr := C.loro_container_value_get_container(lv.ptr)
	if ptr == nil {
		return nil, ErrFailedGetContainer
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
		return nil, ErrFailedGetValue
	}
	value := &LoroValue{ptr: ptr}
	runtime.SetFinalizer(value, func(value *LoroValue) {
		value.Destroy()
	})
	return value, nil
}
