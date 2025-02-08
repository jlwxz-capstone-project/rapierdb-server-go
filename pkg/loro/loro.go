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

func (doc *LoroDoc) Import(data []byte) {
	snapshot := NewRustVecFromBytes(data)
	C.loro_doc_import(doc.ptr, snapshot.ptr)
}

// ----------- Loro Text -----------

type LoroText struct {
	ptr unsafe.Pointer
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
