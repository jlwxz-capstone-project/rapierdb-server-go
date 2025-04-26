use loro::{Container, LoroList, LoroMap, LoroMovableList, LoroText, LoroValue, ValueOrContainer};
use std::ffi::{c_char, CStr, CString};

// Constructor
#[no_mangle]
pub extern "C" fn loro_list_new_empty() -> *mut LoroList {
    let list = LoroList::new();
    let boxed = Box::new(list);
    let ptr = Box::into_raw(boxed);
    ptr
}

// Destructor
#[no_mangle]
pub extern "C" fn destroy_loro_list(ptr: *mut LoroList) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn loro_list_to_container(ptr: *mut LoroList) -> *mut Container {
    unsafe {
        let list = &mut *ptr;
        let container = Container::List(list.clone());
        let boxed = Box::new(container);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn loro_list_is_attached(ptr: *const LoroList) -> i32 {
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
pub extern "C" fn loro_list_len(ptr: *const LoroList) -> usize {
    unsafe {
        let list = &*ptr;
        list.len()
    }
}

// Loro List Push
#[no_mangle]
pub extern "C" fn loro_list_push_value(
    ptr: *mut LoroList,
    value_ptr: *mut LoroValue,
    err: *mut u8,
) {
    unsafe {
        let list = &mut *ptr;
        let value = &mut *value_ptr;
        if list.push(value.clone()).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_list_push_container(
    ptr: *mut LoroList,
    container_ptr: *mut Container,
    err: *mut u8,
) -> *mut Container {
    unsafe {
        let list = &mut *ptr;
        let container = &mut *container_ptr;
        if let Ok(new_container) = list.push_container(container.clone()) {
            *err = 0;
            let boxed = Box::new(new_container);
            let ptr = Box::into_raw(boxed);
            ptr
        } else {
            *err = 1;
            std::ptr::null_mut()
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_list_push_null(ptr: *mut LoroList, err: *mut u8) {
    unsafe {
        let list = &mut *ptr;
        if list.push(LoroValue::Null).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_list_push_bool(ptr: *mut LoroList, value: i32, err: *mut u8) {
    unsafe {
        let list = &mut *ptr;
        if list.push(value != 0).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_list_push_double(ptr: *mut LoroList, value: f64, err: *mut u8) {
    unsafe {
        let list = &mut *ptr;
        if list.push(value).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_list_push_i64(ptr: *mut LoroList, value: i64, err: *mut u8) {
    unsafe {
        let list = &mut *ptr;
        if list.push(value).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_list_push_string(ptr: *mut LoroList, value: *const c_char, err: *mut u8) {
    unsafe {
        let list = &mut *ptr;
        let str = CStr::from_ptr(value).to_string_lossy().into_owned();
        if list.push(str).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_list_push_text(
    ptr: *mut LoroList,
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
pub extern "C" fn loro_list_push_list(
    ptr: *mut LoroList,
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
pub extern "C" fn loro_list_push_movable_list(
    ptr: *mut LoroList,
    list_ptr: *mut LoroMovableList,
    err: *mut u8,
) -> *mut LoroMovableList {
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
pub extern "C" fn loro_list_push_map(
    ptr: *mut LoroList,
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
pub extern "C" fn loro_list_get(ptr: *const LoroList, index: usize) -> *mut ValueOrContainer {
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
pub extern "C" fn loro_list_get_null(ptr: *const LoroList, index: usize, err: *mut u8) {
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
pub extern "C" fn loro_list_get_bool(ptr: *const LoroList, index: usize, err: *mut u8) -> i32 {
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
pub extern "C" fn loro_list_get_double(ptr: *const LoroList, index: usize, err: *mut u8) -> f64 {
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
pub extern "C" fn loro_list_get_i64(ptr: *const LoroList, index: usize, err: *mut u8) -> i64 {
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
pub extern "C" fn loro_list_get_string(
    ptr: *const LoroList,
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
pub extern "C" fn loro_list_get_text(
    ptr: *const LoroList,
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
pub extern "C" fn loro_list_get_list(
    ptr: *const LoroList,
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
pub extern "C" fn loro_list_get_movable_list(
    ptr: *const LoroList,
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
pub extern "C" fn loro_list_get_map(
    ptr: *const LoroList,
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

#[no_mangle]
pub extern "C" fn loro_list_get_items(ptr: *mut LoroList) -> *mut Vec<*mut ValueOrContainer> {
    unsafe {
        let list = &mut *ptr;
        let mut items = Vec::new();
        list.for_each(|item| {
            let boxed = Box::new(item);
            let ptr = Box::into_raw(boxed);
            items.push(ptr);
        });
        let boxed = Box::new(items);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

// Loro List Insert
#[no_mangle]
pub extern "C" fn loro_list_insert_value(
    ptr: *mut LoroList,
    index: usize,
    value_ptr: *mut LoroValue,
    err: *mut u8,
) {
    unsafe {
        let list = &mut *ptr;
        let value = &mut *value_ptr;
        let res = list.insert(index, value.clone());
        if res.is_err() {
            *err = 1;
        } else {
            *err = 0;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_list_insert_container(
    ptr: *mut LoroList,
    index: usize,
    container_ptr: *mut Container,
    err: *mut u8,
) -> *mut Container {
    unsafe {
        let list = &mut *ptr;
        let container = &mut *container_ptr;
        if let Ok(new_container) = list.insert_container(index, container.clone()) {
            *err = 0;
            let boxed = Box::new(new_container);
            let ptr = Box::into_raw(boxed);
            ptr
        } else {
            *err = 1;
            std::ptr::null_mut()
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_list_insert_null(ptr: *mut LoroList, index: usize, err: *mut u8) {
    unsafe {
        let list = &mut *ptr;
        let res = list.insert(index, LoroValue::Null);
        if res.is_err() {
            *err = 1;
        } else {
            *err = 0;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_list_insert_bool(
    ptr: *mut LoroList,
    index: usize,
    value: bool,
    err: *mut u8,
) {
    unsafe {
        let list = &mut *ptr;
        let res = list.insert(index, value);
        if res.is_err() {
            *err = 1;
        } else {
            *err = 0;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_list_insert_i64(ptr: *mut LoroList, index: usize, value: i64, err: *mut u8) {
    unsafe {
        let list = &mut *ptr;
        let res = list.insert(index, value);
        if res.is_err() {
            *err = 1;
        } else {
            *err = 0;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_list_insert_double(
    ptr: *mut LoroList,
    index: usize,
    value: f64,
    err: *mut u8,
) {
    unsafe {
        let list = &mut *ptr;
        let res = list.insert(index, value);
        if res.is_err() {
            *err = 1;
        } else {
            *err = 0;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_list_insert_string(
    ptr: *mut LoroList,
    index: usize,
    value_ptr: *const c_char,
    err: *mut u8,
) {
    unsafe {
        let list = &mut *ptr;
        let value = CStr::from_ptr(value_ptr).to_string_lossy().into_owned();
        let res = list.insert(index, value);
        if res.is_err() {
            *err = 1;
        } else {
            *err = 0;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_list_insert_text(
    ptr: *mut LoroList,
    index: usize,
    text_ptr: *mut LoroText,
    err: *mut u8,
) -> *mut LoroText {
    unsafe {
        let list = &mut *ptr;
        let text = &mut *text_ptr;
        let res = list.insert_container(index, text.clone());
        if let Ok(new_text) = res {
            *err = 0;
            let boxed = Box::new(new_text);
            let ptr = Box::into_raw(boxed);
            ptr
        } else {
            *err = 1;
            std::ptr::null_mut()
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_list_insert_list(
    ptr: *mut LoroList,
    index: usize,
    list_ptr: *mut LoroList,
    err: *mut u8,
) -> *mut LoroList {
    unsafe {
        let list = &mut *ptr;
        let new_list = &mut *list_ptr;
        let res = list.insert_container(index, new_list.clone());
        if let Ok(new_list) = res {
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
pub extern "C" fn loro_list_insert_movable_list(
    ptr: *mut LoroList,
    index: usize,
    movable_list_ptr: *mut LoroMovableList,
    err: *mut u8,
) -> *mut LoroMovableList {
    unsafe {
        let list = &mut *ptr;
        let movable_list = &mut *movable_list_ptr;
        let res = list.insert_container(index, movable_list.clone());
        if let Ok(new_movable_list) = res {
            let boxed = Box::new(new_movable_list);
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
pub extern "C" fn loro_list_insert_map(
    ptr: *mut LoroList,
    index: usize,
    map_ptr: *mut LoroMap,
    err: *mut u8,
) -> *mut LoroMap {
    unsafe {
        let list = &mut *ptr;
        let map = &mut *map_ptr;
        let res = list.insert_container(index, map.clone());
        if let Ok(new_map) = res {
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

// Loro List Delete
#[no_mangle]
pub extern "C" fn loro_list_delete(ptr: *mut LoroList, pos: usize, len: usize, err: *mut u8) {
    unsafe {
        let list = &mut *ptr;
        let res = list.delete(pos, len);
        if res.is_err() {
            *err = 1;
        } else {
            *err = 0;
        }
    }
}

// Loro List Clear
#[no_mangle]
pub extern "C" fn loro_list_clear(ptr: *mut LoroList, err: *mut u8) {
    unsafe {
        let list = &mut *ptr;
        let res = list.clear();
        if res.is_err() {
            *err = 1;
        } else {
            *err = 0;
        }
    }
}
