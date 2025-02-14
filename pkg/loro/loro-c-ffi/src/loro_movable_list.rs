use loro::{Container, LoroList, LoroMap, LoroMovableList, LoroText, LoroValue, ValueOrContainer};
use std::ffi::{c_char, CStr, CString};

// Constructor
#[no_mangle]
pub extern "C" fn loro_movable_list_new_empty() -> *mut LoroMovableList {
    let list = LoroMovableList::new();
    let boxed = Box::new(list);
    let ptr = Box::into_raw(boxed);
    ptr
}

// Destructor
#[no_mangle]
pub extern "C" fn destroy_loro_movable_list(ptr: *mut LoroMovableList) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_to_container(ptr: *mut LoroMovableList) -> *mut Container {
    unsafe {
        let list = &mut *ptr;
        let container = Container::MovableList(list.clone());
        let boxed = Box::new(container);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_is_attached(ptr: *const LoroMovableList) -> i32 {
    unsafe {
        let list = &*ptr;
        if list.is_attached() {
            1
        } else {
            0
        }
    }
}

// Get Length
#[no_mangle]
pub extern "C" fn loro_movable_list_len(ptr: *const LoroMovableList) -> usize {
    unsafe {
        let list = &*ptr;
        list.len()
    }
}

// Loro Movable List Push
#[no_mangle]
pub extern "C" fn loro_movable_list_push_null(ptr: *mut LoroMovableList, err: *mut u8) {
    unsafe {
        let list = &mut *ptr;
        if list.push(LoroValue::Null).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_push_bool(ptr: *mut LoroMovableList, value: i32, err: *mut u8) {
    unsafe {
        let list = &mut *ptr;
        if list.push(value != 0).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_push_double(
    ptr: *mut LoroMovableList,
    value: f64,
    err: *mut u8,
) {
    unsafe {
        let list = &mut *ptr;
        if list.push(value).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_push_i64(ptr: *mut LoroMovableList, value: i64, err: *mut u8) {
    unsafe {
        let list = &mut *ptr;
        if list.push(value).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_push_string(
    ptr: *mut LoroMovableList,
    value: *const c_char,
    err: *mut u8,
) {
    unsafe {
        let list = &mut *ptr;
        let str = CStr::from_ptr(value).to_string_lossy().into_owned();
        if list.push(str).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_push_text(
    ptr: *mut LoroMovableList,
    text_ptr: *mut LoroText,
    err: *mut u8,
) -> *mut LoroText {
    unsafe {
        let list = &mut *ptr;
        let text = &mut *text_ptr;
        if let Ok(new_text) = list.push_container(text.clone()) {
            let boxed = Box::new(new_text);
            let ptr = Box::into_raw(boxed);
            *err = 0;
            ptr
        } else {
            *err = 1;
            std::ptr::null_mut()
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_push_list(
    ptr: *mut LoroMovableList,
    list_ptr: *mut LoroList,
    err: *mut u8,
) -> *mut LoroList {
    unsafe {
        let list = &mut *ptr;
        let new_list = &mut *list_ptr;
        if let Ok(new_list) = list.push_container(new_list.clone()) {
            let boxed = Box::new(new_list);
            let ptr = Box::into_raw(boxed);
            *err = 0;
            ptr
        } else {
            *err = 1;
            std::ptr::null_mut()
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_push_movable_list(
    ptr: *mut LoroMovableList,
    movable_list_ptr: *mut LoroMovableList,
    err: *mut u8,
) -> *mut LoroMovableList {
    unsafe {
        let list = &mut *ptr;
        let new_list = &mut *movable_list_ptr;
        if let Ok(new_list) = list.push_container(new_list.clone()) {
            let boxed = Box::new(new_list);
            let ptr = Box::into_raw(boxed);
            *err = 0;
            ptr
        } else {
            *err = 1;
            std::ptr::null_mut()
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_push_map(
    ptr: *mut LoroMovableList,
    map_ptr: *mut LoroMap,
    err: *mut u8,
) -> *mut LoroMap {
    unsafe {
        let list = &mut *ptr;
        let new_map = &mut *map_ptr;
        if let Ok(new_map) = list.push_container(new_map.clone()) {
            let boxed = Box::new(new_map);
            let ptr = Box::into_raw(boxed);
            *err = 0;
            ptr
        } else {
            *err = 1;
            std::ptr::null_mut()
        }
    }
}

// Loro List Get
#[no_mangle]
pub extern "C" fn loro_movable_list_get(
    ptr: *const LoroMovableList,
    index: usize,
) -> *mut ValueOrContainer {
    unsafe {
        let list = &*ptr;
        let value = list.get(index);
        match value {
            Some(val) => {
                let boxed = Box::new(val);
                let ptr = Box::into_raw(boxed);
                ptr
            }
            None => std::ptr::null_mut(),
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_get_null(
    ptr: *const LoroMovableList,
    index: usize,
    err: *mut u8,
) {
    unsafe {
        let list = &*ptr;
        let value = list.get(index);
        if let Some(value) = value {
            if let Some(value) = value.as_value() {
                if let LoroValue::Null = value {
                    *err = 0;
                    return;
                }
            }
        }
        *err = 1;
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_get_bool(
    ptr: *const LoroMovableList,
    index: usize,
    err: *mut u8,
) -> i32 {
    unsafe {
        let list = &*ptr;
        let value = list.get(index);
        if let Some(value) = value {
            if let Some(value) = value.as_value() {
                if let Some(value) = value.as_bool() {
                    *err = 0;
                    return (*value).into();
                }
            }
        }
        *err = 1;
        -1
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_get_double(
    ptr: *const LoroMovableList,
    index: usize,
    err: *mut u8,
) -> f64 {
    unsafe {
        let list = &*ptr;
        if let Some(value) = list.get(index) {
            if let Some(value) = value.as_value() {
                if let Some(n) = value.as_double() {
                    *err = 0;
                    return *n;
                }
            }
        }
        *err = 1;
        f64::NAN
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_get_i64(
    ptr: *const LoroMovableList,
    index: usize,
    err: *mut u8,
) -> i64 {
    unsafe {
        let list = &*ptr;
        if let Some(value) = list.get(index) {
            if let Some(value) = value.as_value() {
                if let Some(n) = value.as_i64() {
                    *err = 0;
                    return *n;
                }
            }
        }
        *err = 1;
        -1
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_get_string(
    ptr: *const LoroMovableList,
    index: usize,
    err: *mut u8,
) -> *const c_char {
    unsafe {
        let list = &*ptr;
        if let Some(value) = list.get(index) {
            if let Some(value) = value.as_value() {
                if let Some(s) = value.as_string() {
                    if let Ok(c_str) = CString::new(s.as_str()) {
                        *err = 0;
                        return c_str.into_raw();
                    }
                }
            }
        }
        *err = 1;
        std::ptr::null()
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_get_text(
    ptr: *const LoroMovableList,
    index: usize,
    err: *mut u8,
) -> *mut LoroText {
    unsafe {
        let list = &*ptr;
        if let Some(value) = list.get(index) {
            if let Some(value) = value.as_container() {
                if let Some(text) = value.as_text() {
                    let boxed = Box::new(text.clone());
                    let ptr = Box::into_raw(boxed);
                    *err = 0;
                    return ptr;
                }
            }
        }
        *err = 1;
        std::ptr::null_mut()
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_get_list(
    ptr: *const LoroMovableList,
    index: usize,
    err: *mut u8,
) -> *mut LoroList {
    unsafe {
        let list = &*ptr;
        if let Some(value) = list.get(index) {
            if let Some(value) = value.as_container() {
                if let Some(list) = value.as_list() {
                    let boxed = Box::new(list.clone());
                    let ptr = Box::into_raw(boxed);
                    *err = 0;
                    return ptr;
                }
            }
        }
        *err = 1;
        std::ptr::null_mut()
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_get_movable_list(
    ptr: *const LoroMovableList,
    index: usize,
    err: *mut u8,
) -> *mut LoroMovableList {
    unsafe {
        let list = &*ptr;
        if let Some(value) = list.get(index) {
            if let Some(value) = value.as_container() {
                if let Some(movable_list) = value.as_movable_list() {
                    let boxed = Box::new(movable_list.clone());
                    let ptr = Box::into_raw(boxed);
                    *err = 0;
                    return ptr;
                }
            }
        }
        *err = 1;
        std::ptr::null_mut()
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_get_map(
    ptr: *const LoroMovableList,
    index: usize,
    err: *mut u8,
) -> *mut LoroMap {
    unsafe {
        let list = &*ptr;
        if let Some(value) = list.get(index) {
            if let Some(value) = value.as_container() {
                if let Some(map) = value.as_map() {
                    let boxed = Box::new(map.clone());
                    let ptr = Box::into_raw(boxed);
                    *err = 0;
                    return ptr;
                }
            }
        }
        *err = 1;
        std::ptr::null_mut()
    }
}
