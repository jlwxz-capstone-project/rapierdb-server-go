use loro::{Container, LoroList, LoroMap, LoroMovableList, LoroText, LoroValue, ValueOrContainer};
use std::ffi::{c_char, CStr, CString};

#[no_mangle]
pub extern "C" fn loro_map_new_empty() -> *mut LoroMap {
    let map = LoroMap::new();
    let boxed = Box::new(map);
    let ptr = Box::into_raw(boxed);
    ptr
}

#[no_mangle]
pub extern "C" fn loro_map_len(ptr: *mut LoroMap) -> usize {
    unsafe {
        let map = &mut *ptr;
        map.len()
    }
}

#[no_mangle]
pub extern "C" fn loro_map_to_container(ptr: *mut LoroMap) -> *mut Container {
    unsafe {
        let map = &mut *ptr;
        let container = Container::Map(map.clone());
        let boxed = Box::new(container);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn loro_map_is_attached(ptr: *const LoroMap) -> i32 {
    unsafe {
        let map = &*ptr;
        if map.is_attached() {
            1
        } else {
            0
        }
    }
}

#[no_mangle]
pub extern "C" fn destroy_loro_map(ptr: *mut LoroMap) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn loro_map_get(
    ptr: *const LoroMap,
    key_ptr: *const c_char,
) -> *mut ValueOrContainer {
    unsafe {
        let map = &*ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        let val = map.get(&key);
        match val {
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
pub extern "C" fn loro_map_get_null(ptr: *mut LoroMap, key_ptr: *const c_char, err: *mut u8) {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        if let Some(value) = map.get(&key) {
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
pub extern "C" fn loro_map_get_bool(
    ptr: *mut LoroMap,
    key_ptr: *const c_char,
    err: *mut u8,
) -> i32 {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        if let Some(value) = map.get(&key) {
            if let Some(value) = value.as_value() {
                if let Some(value) = value.as_bool() {
                    *err = 0;
                    return if *value { 1 } else { 0 };
                }
            }
        }
        *err = 1;
        -1
    }
}

#[no_mangle]
pub extern "C" fn loro_map_get_double(
    ptr: *mut LoroMap,
    key_ptr: *const c_char,
    err: *mut u8,
) -> f64 {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        if let Some(value) = map.get(&key) {
            if let Some(value) = value.as_value() {
                if let Some(value) = value.as_double() {
                    *err = 0;
                    return *value;
                }
            }
        }
        *err = 1;
        0.0
    }
}

#[no_mangle]
pub extern "C" fn loro_map_get_i64(ptr: *mut LoroMap, key_ptr: *const c_char, err: *mut u8) -> i64 {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        if let Some(value) = map.get(&key) {
            if let Some(value) = value.as_value() {
                if let Some(value) = value.as_i64() {
                    *err = 0;
                    return *value;
                }
            }
        }
        *err = 1;
        0
    }
}

#[no_mangle]
pub extern "C" fn loro_map_get_string(
    ptr: *mut LoroMap,
    key_ptr: *const c_char,
    err: *mut u8,
) -> *const c_char {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        if let Some(value) = map.get(&key) {
            if let Some(value) = value.as_value() {
                if let Some(value) = value.as_string() {
                    *err = 0;
                    return CString::new(value.to_string()).unwrap().into_raw();
                }
            }
        }
        *err = 1;
        std::ptr::null()
    }
}

#[no_mangle]
pub extern "C" fn loro_map_get_text(
    ptr: *mut LoroMap,
    key_ptr: *const c_char,
    err: *mut u8,
) -> *mut LoroText {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        if let Some(value) = map.get(&key) {
            if let Some(value) = value.as_container() {
                if let Some(value) = value.as_text() {
                    *err = 0;
                    return Box::into_raw(Box::new(value.clone()));
                }
            }
        }
        *err = 1;
        std::ptr::null_mut()
    }
}

#[no_mangle]
pub extern "C" fn loro_map_get_list(
    ptr: *mut LoroMap,
    key_ptr: *const c_char,
    err: *mut u8,
) -> *mut LoroList {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        if let Some(value) = map.get(&key) {
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
pub extern "C" fn loro_map_get_movable_list(
    ptr: *mut LoroMap,
    key_ptr: *const c_char,
    err: *mut u8,
) -> *mut LoroMovableList {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        if let Some(value) = map.get(&key) {
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
pub extern "C" fn loro_map_get_map(
    ptr: *mut LoroMap,
    key_ptr: *const c_char,
    err: *mut u8,
) -> *mut LoroMap {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        if let Some(value) = map.get(&key) {
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

// Insert

#[no_mangle]
pub extern "C" fn loro_map_insert_null(ptr: *mut LoroMap, key_ptr: *const c_char, err: *mut u8) {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        if map.insert(&key, LoroValue::Null).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_map_insert_bool(
    ptr: *mut LoroMap,
    key_ptr: *const c_char,
    bool_value: i32,
    err: *mut u8,
) {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        if map.insert(&key, bool_value != 0).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_map_insert_double(
    ptr: *mut LoroMap,
    key_ptr: *const c_char,
    double_value: f64,
    err: *mut u8,
) {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        if map.insert(&key, double_value).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_map_insert_i64(
    ptr: *mut LoroMap,
    key_ptr: *const c_char,
    int_value: i64,
    err: *mut u8,
) {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        if map.insert(&key, int_value).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_map_insert_string(
    ptr: *mut LoroMap,
    key_ptr: *const c_char,
    str_value: *const c_char,
    err: *mut u8,
) {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        let str_value = CStr::from_ptr(str_value).to_string_lossy().into_owned();
        if map.insert(&key, str_value).is_err() {
            *err = 1;
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_map_insert_text(
    ptr: *mut LoroMap,
    key_ptr: *const c_char,
    text_ptr: *mut LoroText,
    err: *mut u8,
) -> *mut LoroText {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        let text = &mut *text_ptr;
        if let Ok(new_text) = map.insert_container(&key, text.clone()) {
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
pub extern "C" fn loro_map_insert_list(
    ptr: *mut LoroMap,
    key_ptr: *const c_char,
    list_ptr: *mut LoroList,
    err: *mut u8,
) -> *mut LoroList {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        let list = &mut *list_ptr;
        if let Ok(new_list) = map.insert_container(&key, list.clone()) {
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
pub extern "C" fn loro_map_insert_movable_list(
    ptr: *mut LoroMap,
    key_ptr: *const c_char,
    list_ptr: *mut LoroMovableList,
    err: *mut u8,
) -> *mut LoroMovableList {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        let list = &mut *list_ptr;
        if let Ok(new_list) = map.insert_container(&key, list.clone()) {
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
pub extern "C" fn loro_map_insert_map(
    ptr: *mut LoroMap,
    key_ptr: *const c_char,
    map_ptr: *mut LoroMap,
    err: *mut u8,
) -> *mut LoroMap {
    unsafe {
        let map = &mut *ptr;
        let key = CStr::from_ptr(key_ptr).to_string_lossy().into_owned();
        let map_to_insert = &mut *map_ptr;
        if let Ok(new_map) = map.insert_container(&key, map_to_insert.clone()) {
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

#[no_mangle]
pub extern "C" fn loro_map_get_items(ptr: *mut LoroMap) -> *mut Vec<*mut u8> {
    unsafe {
        let map = &mut *ptr;
        let mut items = Vec::new();
        map.for_each(|key, value| {
            items.push(CString::new(key.to_string()).unwrap().into_raw() as *mut u8);
            items.push(Box::into_raw(Box::new(value)) as *mut u8);
        });
        let boxed = Box::new(items);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}
