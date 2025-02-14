use loro::{Container, LoroText, UpdateOptions};
use std::ffi::{c_char, CStr, CString};

#[no_mangle]
pub extern "C" fn loro_text_to_string(text_ptr: *mut LoroText, err: *mut u8) -> *mut c_char {
    unsafe {
        let text = &mut *text_ptr;
        let s = text.to_string();
        match CString::new(s) {
            Ok(cstr) => {
                *err = 0;
                cstr.into_raw()
            }
            Err(_) => {
                *err = 1;
                std::ptr::null_mut()
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn new_loro_text() -> *mut LoroText {
    let text = LoroText::new();
    let boxed = Box::new(text);
    let ptr = Box::into_raw(boxed);
    ptr
}

#[no_mangle]
pub extern "C" fn loro_text_to_container(ptr: *mut LoroText) -> *mut Container {
    unsafe {
        let text = &mut *ptr;
        let container = Container::Text(text.clone());
        let boxed = Box::new(container);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn loro_text_is_attached(ptr: *const LoroText) -> i32 {
    unsafe {
        let text = &*ptr;
        if text.is_attached() {
            1
        } else {
            0
        }
    }
}

#[no_mangle]
pub extern "C" fn destroy_loro_text(ptr: *mut LoroText) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn update_loro_text(text_ptr: *mut LoroText, content: *const c_char, err: *mut u8) {
    unsafe {
        let text = &mut *text_ptr;
        let s = CStr::from_ptr(content).to_string_lossy().into_owned();
        if text.update(&s, UpdateOptions::default()).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn insert_loro_text(
    text_ptr: *mut LoroText,
    pos: usize,
    content: *const c_char,
    err: *mut u8,
) {
    unsafe {
        let text = &mut *text_ptr;
        let s = CStr::from_ptr(content).to_string_lossy().into_owned();
        if text.insert(pos, &s).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn insert_loro_text_utf8(
    text_ptr: *mut LoroText,
    pos: usize,
    content: *const c_char,
    err: *mut u8,
) {
    unsafe {
        let text = &mut *text_ptr;
        let s = CStr::from_ptr(content).to_string_lossy().into_owned();
        if text.insert_utf8(pos, &s).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_text_length(text_ptr: *mut LoroText) -> usize {
    unsafe {
        let text = &mut *text_ptr;
        text.len_unicode()
    }
}

#[no_mangle]
pub extern "C" fn loro_text_length_utf8(text_ptr: *mut LoroText) -> usize {
    unsafe {
        let text = &mut *text_ptr;
        text.len_utf8()
    }
}
