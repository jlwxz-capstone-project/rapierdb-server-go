use loro::{ContainerID, LoroBinaryValue, LoroListValue, LoroMapValue, LoroStringValue, LoroValue};
use loro_internal::container::list;
use std::ffi::{c_char, CStr, CString};

#[no_mangle]
pub extern "C" fn destroy_loro_value(ptr: *mut LoroValue) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn loro_value_get_type(ptr: *mut LoroValue) -> i32 {
    unsafe {
        let value = &*ptr;
        match value {
            LoroValue::Null => 0,
            LoroValue::Bool(..) => 1,
            LoroValue::Double(..) => 2,
            LoroValue::I64(..) => 3,
            LoroValue::String(..) => 5,
            LoroValue::Map(..) => 7,
            LoroValue::List(..) => 6,
            LoroValue::Binary(..) => 8,
            LoroValue::Container(..) => 9,
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_value_get_bool(ptr: *mut LoroValue, err: *mut u8) -> i32 {
    unsafe {
        let value = &*ptr;
        match value {
            LoroValue::Bool(value) => {
                *err = 0;
                *value as i32
            }
            _ => {
                *err = 1;
                -1
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_value_get_double(ptr: *mut LoroValue, err: *mut u8) -> f64 {
    unsafe {
        let value = &*ptr;
        match value {
            LoroValue::Double(value) => {
                *err = 0;
                *value
            }
            _ => {
                *err = 1;
                f64::NAN
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_value_get_i64(ptr: *mut LoroValue, err: *mut u8) -> i64 {
    unsafe {
        let value = &*ptr;
        match value {
            LoroValue::I64(value) => {
                *err = 0;
                *value
            }
            _ => {
                *err = 1;
                -1
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_value_get_string(ptr: *mut LoroValue, err: *mut u8) -> *const c_char {
    unsafe {
        let value = &*ptr;
        match value {
            LoroValue::String(value) => {
                let s = (**value).clone();
                match CString::new(s) {
                    Ok(cstr) => {
                        *err = 0;
                        cstr.into_raw()
                    }
                    Err(_) => {
                        *err = 1;
                        std::ptr::null()
                    }
                }
            }
            _ => {
                *err = 1;
                std::ptr::null()
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_value_get_binary(ptr: *mut LoroValue, err: *mut u8) -> *mut Vec<u8> {
    unsafe {
        let value = &*ptr;
        match value {
            LoroValue::Binary(value) => {
                let new_vec = Box::new((**value).clone());
                *err = 0;
                Box::into_raw(new_vec)
            }
            _ => {
                *err = 1;
                std::ptr::null_mut()
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_value_get_list(
    ptr: *mut LoroValue,
    err: *mut u8,
) -> *mut Vec<*mut LoroValue> {
    unsafe {
        let value = &*ptr;
        match value {
            LoroValue::List(value) => {
                let value = value.clone().unwrap();
                let mut ret = Vec::with_capacity(value.len());
                for item in value.iter() {
                    let boxed = Box::new(item.clone());
                    ret.push(Box::into_raw(boxed));
                }
                *err = 0;
                Box::into_raw(Box::new(ret))
            }
            _ => {
                *err = 1;
                std::ptr::null_mut()
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_value_get_map(ptr: *mut LoroValue, err: *mut u8) -> *mut Vec<*mut u8> {
    unsafe {
        let value = &*ptr;
        match value {
            LoroValue::Map(value) => {
                let value = value.clone().unwrap();
                let mut ret = Vec::with_capacity(value.len() * 2);
                for (k, v) in value.iter() {
                    let k_ptr = CString::new(k.clone()).unwrap().into_raw();
                    let v_ptr = Box::into_raw(Box::new(v.clone()));
                    ret.push(k_ptr as *mut u8);
                    ret.push(v_ptr as *mut u8);
                }
                *err = 0;
                Box::into_raw(Box::new(ret))
            }
            _ => {
                *err = 1;
                std::ptr::null_mut()
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_value_get_container_id(
    ptr: *mut LoroValue,
    err: *mut u8,
) -> *mut ContainerID {
    unsafe {
        let value = &*ptr;
        match value {
            LoroValue::Container(value) => {
                *err = 0;
                Box::into_raw(Box::new(value.clone()))
            }
            _ => {
                *err = 1;
                std::ptr::null_mut()
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_value_new_null() -> *mut LoroValue {
    let value = LoroValue::Null;
    Box::into_raw(Box::new(value))
}

#[no_mangle]
pub extern "C" fn loro_value_new_bool(value: i32) -> *mut LoroValue {
    let value = LoroValue::Bool(value != 0);
    Box::into_raw(Box::new(value))
}

#[no_mangle]
pub extern "C" fn loro_value_new_double(value: f64) -> *mut LoroValue {
    let value = LoroValue::Double(value);
    Box::into_raw(Box::new(value))
}

#[no_mangle]
pub extern "C" fn loro_value_new_i64(value: i64) -> *mut LoroValue {
    let value = LoroValue::I64(value);
    Box::into_raw(Box::new(value))
}

#[no_mangle]
pub extern "C" fn loro_value_new_string(value: *const c_char) -> *mut LoroValue {
    let s = unsafe { CStr::from_ptr(value).to_string_lossy().to_string() };
    let value = LoroValue::String(LoroStringValue::from(s));
    Box::into_raw(Box::new(value))
}

#[no_mangle]
pub extern "C" fn loro_value_new_binary(value: *mut Vec<u8>) -> *mut LoroValue {
    unsafe {
        let vec = &*value;
        let value = LoroValue::Binary(LoroBinaryValue::from(vec.clone()));
        Box::into_raw(Box::new(value))
    }
}

#[no_mangle]
pub extern "C" fn loro_value_new_list(value: *mut Vec<*mut LoroValue>) -> *mut LoroValue {
    unsafe {
        let vec = &*value;
        let mut value_vec = Vec::with_capacity(vec.len());
        for item in vec.iter() {
            let item_value = (**item).clone();
            value_vec.push(item_value);
        }
        let value = LoroValue::List(LoroListValue::from(value_vec));
        Box::into_raw(Box::new(value))
    }
}

#[no_mangle]
pub extern "C" fn loro_value_new_map(value: *mut Vec<*mut u8>) -> *mut LoroValue {
    unsafe {
        let vec = &*value;
        let mut value_vec = Vec::with_capacity(vec.len() / 2);
        for i in (0..vec.len()).step_by(2) {
            let key_ptr = vec[i] as *const c_char;
            let val_ptr = vec[i + 1] as *mut LoroValue;
            let key = CStr::from_ptr(key_ptr).to_string_lossy().to_string();
            let val = (&*val_ptr).clone();
            value_vec.push((key, val));
        }
        let value = LoroValue::Map(LoroMapValue::from(value_vec));
        Box::into_raw(Box::new(value))
    }
}

#[no_mangle]
pub extern "C" fn loro_value_to_json(ptr: *mut LoroValue) -> *mut c_char {
    unsafe {
        let value = &*ptr;
        let json = serde_json::to_string(value);
        match json {
            Ok(json) => {
                let cstr = CString::new(json).unwrap();
                cstr.into_raw()
            }
            Err(_) => std::ptr::null_mut(),
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_value_from_json(json: *const c_char) -> *mut LoroValue {
    unsafe {
        let json = CStr::from_ptr(json).to_string_lossy().to_string();
        let value = serde_json::from_str(&json);
        match value {
            Ok(value) => Box::into_raw(Box::new(value)),
            Err(_) => std::ptr::null_mut(),
        }
    }
}
