use loro::TextDelta;

#[no_mangle]
pub extern "C" fn destroy_text_delta(ptr: *mut TextDelta) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn text_delta_get_type(ptr: *mut TextDelta) -> i32 {
    unsafe {
        let delta = &*ptr;
        match delta {
            TextDelta::Insert { .. } => 0,
            TextDelta::Delete { .. } => 1,
            TextDelta::Retain { .. } => 2,
        }
    }
}
