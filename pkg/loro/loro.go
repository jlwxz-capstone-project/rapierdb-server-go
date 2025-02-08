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

// ----------- Rust Bytes Vec -----------

type RustBytesVec struct {
	ptr      unsafe.Pointer
	dataPtr  unsafe.Pointer
	len      uint32
	capacity uint32
}

func (vec *RustBytesVec) Destroy() {
	C.destroy_bytes_vec(vec.ptr)
}

func (vec *RustBytesVec) GetSize() uint32 {
	return vec.len
}

func (vec *RustBytesVec) GetCapacity() uint32 {
	return vec.capacity
}

func (vec *RustBytesVec) Bytes() []byte {
	return unsafe.Slice((*byte)(vec.dataPtr), vec.len)
}

// ----------- Loro Doc -----------

type LoroDoc struct {
	ptr unsafe.Pointer
}

func (doc *LoroDoc) Destroy() {
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
	var dataPtr *C.uint8_t
	var len C.uint32_t
	var cap C.uint32_t

	ptr := C.export_loro_doc_snapshot(
		doc.ptr,
		&dataPtr,
		&len,
		&cap,
	)

	bytesVec := &RustBytesVec{
		ptr:      ptr,
		dataPtr:  unsafe.Pointer(dataPtr),
		len:      uint32(len),
		capacity: uint32(cap),
	}
	runtime.SetFinalizer(bytesVec, func(vec *RustBytesVec) {
		vec.Destroy()
	})
	return bytesVec
}

// ----------- Loro Text -----------

type LoroText struct {
	ptr unsafe.Pointer
}

func (text *LoroText) Destroy() {
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

// ----------- Loro Map -----------

type LoroMap struct {
	ptr unsafe.Pointer
}

func (m *LoroMap) Destroy() {
	C.destroy_loro_map(m.ptr)
}

// ----------- Loro List -----------

type LoroList struct {
	ptr unsafe.Pointer
}

func (list *LoroList) Destroy() {
	C.destroy_loro_list(list.ptr)
}

// ----------- Loro Movable List -----------

type LoroMovableList struct {
	ptr unsafe.Pointer
}

func (movableList *LoroMovableList) Destroy() {
	C.destroy_loro_movable_list(movableList.ptr)
}
