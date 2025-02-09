package loro

/*
#cgo LDFLAGS: -L./loro-c-ffi/target/release -lloro_c_ffi
#include <stdlib.h>
#include "loro-c-ffi/loro_c_ffi.h"
*/
import "C"
import (
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

func NewRustVecFromBytes(data []byte) *RustBytesVec {
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

// ----------- Loro Doc -----------

type LoroDoc struct {
	ptr unsafe.Pointer
}

func (doc *LoroDoc) Destroy() {
	// fmt.Println("destroying loro doc")
	C.destroy_loro_doc(doc.ptr)
}

func NewLoroDoc() *LoroDoc {
	ret := C.create_loro_doc()
	loroDoc := &LoroDoc{
		ptr: unsafe.Pointer(ret),
	}
	runtime.SetFinalizer(loroDoc, func(doc *LoroDoc) {
		doc.Destroy()
	})
	return loroDoc
}

func (doc *LoroDoc) GetText(id string) *LoroText {
	idPtr := C.CString(id)
	defer C.free(unsafe.Pointer(idPtr))
	ret := C.get_text(doc.ptr, idPtr)
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
	ret := C.get_list(doc.ptr, idPtr)
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
	ret := C.get_movable_list(doc.ptr, idPtr)
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
	ret := C.get_map(doc.ptr, idPtr)
	loroMap := &LoroMap{
		ptr: unsafe.Pointer(ret),
	}
	runtime.SetFinalizer(loroMap, func(m *LoroMap) {
		m.Destroy()
	})
	return loroMap
}

func (doc *LoroDoc) ExportSnapshot() *RustBytesVec {
	ptr := C.export_loro_doc_snapshot(doc.ptr)
	bytesVec := &RustBytesVec{
		ptr: ptr,
	}
	runtime.SetFinalizer(bytesVec, func(vec *RustBytesVec) {
		vec.Destroy()
	})
	return bytesVec
}

func (doc *LoroDoc) ExportAllUpdates() *RustBytesVec {
	ptr := C.export_loro_doc_all_updates(doc.ptr)
	bytesVec := &RustBytesVec{
		ptr: ptr,
	}
	runtime.SetFinalizer(bytesVec, func(vec *RustBytesVec) {
		vec.Destroy()
	})
	return bytesVec
}

func (doc *LoroDoc) ExportUpdatesFrom(from *VersionVector) *RustBytesVec {
	ptr := C.export_loro_doc_updates_from(doc.ptr, from.ptr)
	bytesVec := &RustBytesVec{
		ptr: ptr,
	}
	runtime.SetFinalizer(bytesVec, func(vec *RustBytesVec) {
		vec.Destroy()
	})
	return bytesVec
}

func (doc *LoroDoc) ExportUpdatesTill(till *VersionVector) *RustBytesVec {
	ptr := C.export_loro_doc_updates_till(doc.ptr, till.ptr)
	bytesVec := &RustBytesVec{
		ptr: ptr,
	}
	runtime.SetFinalizer(bytesVec, func(vec *RustBytesVec) {
		vec.Destroy()
	})
	return bytesVec
}

func (doc *LoroDoc) Import(data []byte) {
	snapshot := NewRustVecFromBytes(data)
	C.loro_doc_import(doc.ptr, snapshot.ptr)
}

func (doc *LoroDoc) GetOplogVv() *VersionVector {
	ptr := C.get_oplog_vv(doc.ptr)
	vv := &VersionVector{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(vv, func(vv *VersionVector) {
		vv.Destroy()
	})
	return vv
}

func (doc *LoroDoc) GetStateVv() *VersionVector {
	ptr := C.get_state_vv(doc.ptr)
	vv := &VersionVector{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(vv, func(vv *VersionVector) {
		vv.Destroy()
	})
	return vv
}

func (doc *LoroDoc) GetOplogFrontiers() *Frontiers {
	ptr := C.get_oplog_frontiers(doc.ptr)
	frontier := &Frontiers{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(frontier, func(f *Frontiers) {
		f.Destroy()
	})
	return frontier
}

func (doc *LoroDoc) GetStateFrontiers() *Frontiers {
	ptr := C.get_state_frontiers(doc.ptr)
	frontier := &Frontiers{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(frontier, func(f *Frontiers) {
		f.Destroy()
	})
	return frontier
}

func (doc *LoroDoc) FrontiersToVv(frontiers *Frontiers) *VersionVector {
	ptr := C.frontiers_to_vv(doc.ptr, frontiers.ptr)
	vv := &VersionVector{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(vv, func(vv *VersionVector) {
		vv.Destroy()
	})
	return vv
}

func (doc *LoroDoc) VvToFrontiers(vv *VersionVector) *Frontiers {
	ptr := C.vv_to_frontiers(doc.ptr, vv.ptr)
	frontiers := &Frontiers{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(frontiers, func(f *Frontiers) {
		f.Destroy()
	})
	return frontiers
}

func (doc *LoroDoc) Fork() *LoroDoc {
	ptr := C.fork_doc(doc.ptr)
	loroDoc := &LoroDoc{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(loroDoc, func(doc *LoroDoc) {
		doc.Destroy()
	})
	return loroDoc
}

func (doc *LoroDoc) ForkAt(frontiers *Frontiers) *LoroDoc {
	ptr := C.fork_doc_at(doc.ptr, frontiers.ptr)
	loroDoc := &LoroDoc{
		ptr: unsafe.Pointer(ptr),
	}
	runtime.SetFinalizer(loroDoc, func(doc *LoroDoc) {
		doc.Destroy()
	})
	return loroDoc
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

func (text *LoroText) ToString() string {
	ret := C.loro_text_to_string(text.ptr)
	return C.GoString(ret)
}

func (text *LoroText) UpdateText(content string) {
	contentPtr := C.CString(content)
	defer C.free(unsafe.Pointer(contentPtr))
	C.update_loro_text(text.ptr, contentPtr)
}

func (text *LoroText) InsertText(content string, pos uint32) {
	contentPtr := C.CString(content)
	defer C.free(unsafe.Pointer(contentPtr))
	C.insert_loro_text(text.ptr, C.uint32_t(pos), contentPtr)
}

func (text *LoroText) InsertTextUtf8(content string, pos uint32) {
	contentPtr := C.CString(content)
	defer C.free(unsafe.Pointer(contentPtr))
	C.insert_loro_text_utf8(text.ptr, C.uint32_t(pos), contentPtr)
}

func (text *LoroText) GetLength() uint32 {
	return uint32(C.loro_text_length(text.ptr))
}

func (text *LoroText) GetLengthUtf8() uint32 {
	return uint32(C.loro_text_length_utf8(text.ptr))
}

// ----------- Loro Map -----------

type LoroMap struct {
	ptr unsafe.Pointer
}

func (m *LoroMap) Destroy() {
	// fmt.Println("destroying loro map")
	C.destroy_loro_map(m.ptr)
}

// ----------- Loro List -----------

type LoroList struct {
	ptr unsafe.Pointer
}

func (list *LoroList) Destroy() {
	// fmt.Println("destroying loro list")
	C.destroy_loro_list(list.ptr)
}

// ----------- Loro Movable List -----------

type LoroMovableList struct {
	ptr unsafe.Pointer
}

func (movableList *LoroMovableList) Destroy() {
	// fmt.Println("destroying loro movable list")
	C.destroy_loro_movable_list(movableList.ptr)
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
