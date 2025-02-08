use loro::{ExportMode, LoroDoc, LoroList, LoroMap, LoroMovableList, LoroText, UpdateOptions};
use std::ffi::{CStr, CString};
use std::os::raw::c_char;

#[no_mangle]
pub extern "C" fn create_loro_doc() -> *mut LoroDoc {
    let doc = LoroDoc::new();
    let boxed = Box::new(doc);
    let ptr = Box::into_raw(boxed);
    ptr
}

#[no_mangle]
pub extern "C" fn destroy_loro_doc(ptr: *mut LoroDoc) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn get_text(doc_ptr: *mut LoroDoc, id_ptr: *const c_char) -> *mut LoroText {
    unsafe {
        let doc = &mut *doc_ptr;
        let id = CStr::from_ptr(id_ptr).to_string_lossy().into_owned();
        let text = doc.get_text(id);
        let boxed = Box::new(text);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn destroy_loro_text(ptr: *mut LoroText) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn loro_text_to_string(text_ptr: *mut LoroText) -> *mut c_char {
    unsafe {
        let text = &mut *text_ptr;
        let s = text.to_string();
        let ptr = CString::new(s).unwrap().into_raw();
        ptr
    }
}

#[no_mangle]
pub extern "C" fn update_loro_text(text_ptr: *mut LoroText, content: *const c_char) {
    unsafe {
        let text = &mut *text_ptr;
        let s = CStr::from_ptr(content).to_string_lossy().into_owned();
        text.update(&s, UpdateOptions::default()).unwrap();
    }
}

#[no_mangle]
pub extern "C" fn get_list(doc_ptr: *mut LoroDoc, id_ptr: *const c_char) -> *mut LoroList {
    unsafe {
        let doc = &mut *doc_ptr;
        let id = CStr::from_ptr(id_ptr).to_string_lossy().into_owned();
        let list = doc.get_list(id);
        let boxed = Box::new(list);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn destroy_loro_list(ptr: *mut LoroList) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn get_movable_list(
    doc_ptr: *mut LoroDoc,
    id_ptr: *const c_char,
) -> *mut LoroMovableList {
    unsafe {
        let doc = &mut *doc_ptr;
        let id = CStr::from_ptr(id_ptr).to_string_lossy().into_owned();
        let list = doc.get_movable_list(id);
        let boxed = Box::new(list);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn destroy_loro_movable_list(ptr: *mut LoroMovableList) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn get_map(doc_ptr: *mut LoroDoc, id_ptr: *const c_char) -> *mut LoroMap {
    unsafe {
        let doc = &mut *doc_ptr;
        let id = CStr::from_ptr(id_ptr).to_string_lossy().into_owned();
        let map = doc.get_map(id);
        let boxed = Box::new(map);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn destroy_loro_map(ptr: *mut LoroMap) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn export_loro_doc_snapshot(
    doc_ptr: *mut LoroDoc,
    arr_ptr: *mut *mut u8,
    len_ptr: *mut usize,
    cap_ptr: *mut usize,
) -> *mut Vec<u8> {
    unsafe {
        let doc = &mut *doc_ptr;
        let snapshot = doc.export(ExportMode::snapshot()).unwrap();
        let boxed = Box::new(snapshot);
        let boxed_ptr = Box::into_raw(boxed);
        let arr = (*boxed_ptr).as_mut_ptr();
        let len = (*boxed_ptr).len();
        let cap = (*boxed_ptr).capacity();
        *arr_ptr = arr;
        *len_ptr = len;
        *cap_ptr = cap;
        boxed_ptr
    }
}

#[no_mangle]
pub extern "C" fn destroy_bytes_vec(ptr: *mut Vec<u8>) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}
