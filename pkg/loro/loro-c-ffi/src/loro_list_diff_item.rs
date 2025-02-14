use loro::{event::ListDiffItem, LoroValue, ValueOrContainer};

#[no_mangle]
pub extern "C" fn destroy_list_diff_item(ptr: *mut ListDiffItem) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn list_diff_item_get_type(ptr: *mut ListDiffItem) -> i32 {
    unsafe {
        let item = &*ptr;
        match item {
            ListDiffItem::Insert { .. } => 0,
            ListDiffItem::Delete { .. } => 1,
            ListDiffItem::Retain { .. } => 2,
        }
    }
}

#[no_mangle]
pub extern "C" fn list_diff_item_get_insert(
    ptr: *mut ListDiffItem,
    is_move_ptr: *mut u8,
    err: *mut u8,
) -> *mut Vec<*mut ValueOrContainer> {
    unsafe {
        let item = &*ptr;
        match item {
            ListDiffItem::Insert { insert, is_move } => {
                *err = 0;
                *is_move_ptr = if *is_move { 1 } else { 0 };
                let mut vec = Vec::new();
                for item in insert {
                    let boxed = Box::new(item.clone());
                    let ptr = Box::into_raw(boxed);
                    vec.push(ptr);
                }
                let boxed = Box::new(vec);
                Box::into_raw(boxed)
            }
            _ => {
                *err = 1;
                std::ptr::null_mut()
            }
        }
    }
}
#[no_mangle]
pub extern "C" fn list_diff_item_get_delete_count(ptr: *mut ListDiffItem, err: *mut u8) -> usize {
    unsafe {
        let item = &*ptr;
        match item {
            ListDiffItem::Delete { delete, .. } => {
                *err = 0;
                *delete
            }
            _ => {
                *err = 1;
                0
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn list_diff_item_get_retain_count(ptr: *mut ListDiffItem, err: *mut u8) -> usize {
    unsafe {
        let item = &*ptr;
        match item {
            ListDiffItem::Retain { retain, .. } => {
                *err = 0;
                *retain
            }
            _ => {
                *err = 1;
                0
            }
        }
    }
}
