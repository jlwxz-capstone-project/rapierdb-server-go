use loro::{Container, LoroValue, ValueOrContainer};

#[no_mangle]
pub extern "C" fn destroy_loro_container_value(ptr: *mut ValueOrContainer) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn loro_container_value_get_type(ptr: *mut ValueOrContainer) -> u8 {
    unsafe {
        let value = &*ptr;
        match value {
            ValueOrContainer::Value(_) => 0,
            ValueOrContainer::Container(_) => 1,
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_container_value_get_container(ptr: *mut ValueOrContainer) -> *mut Container {
    unsafe {
        let value = &*ptr;
        match value {
            ValueOrContainer::Value(v) => std::ptr::null_mut(),
            ValueOrContainer::Container(c) => {
                let boxed = Box::new(c.clone());
                Box::into_raw(boxed)
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_container_value_get_value(ptr: *mut ValueOrContainer) -> *mut LoroValue {
    unsafe {
        let value = &*ptr;
        match value {
            ValueOrContainer::Value(v) => {
                let boxed = Box::new(v.clone());
                Box::into_raw(boxed)
            }
            ValueOrContainer::Container(c) => std::ptr::null_mut(),
        }
    }
}
