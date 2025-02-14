use loro::{
    event::{Diff, ListDiffItem, MapDelta},
    TextDelta, TreeDiff,
};

#[no_mangle]
pub extern "C" fn destroy_diff_event(ptr: *mut u8) {
    // make rust happy
    unsafe {
        let ptr2 = ptr as *mut Diff;
        let _ = Box::from_raw(ptr2);
    }
}

#[no_mangle]
pub extern "C" fn diff_event_get_list_diff(ptr: *mut Diff) -> *mut Vec<*mut ListDiffItem> {
    unsafe {
        let diff = &*ptr;
        match diff {
            Diff::List(list) => {
                let items = list
                    .iter()
                    .map(|item| {
                        let boxed = Box::new(item.clone());
                        let ptr = Box::into_raw(boxed);
                        ptr
                    })
                    .collect();
                let boxed = Box::new(items);
                Box::into_raw(boxed)
            }
            _ => std::ptr::null_mut(),
        }
    }
}

#[no_mangle]
pub extern "C" fn diff_event_get_text_delta(ptr: *mut Diff) -> *mut Vec<*mut TextDelta> {
    unsafe {
        let diff = &*ptr;
        match diff {
            Diff::Text(delta) => {
                let items = delta
                    .iter()
                    .map(|item| {
                        let boxed = Box::new(item.clone());
                        let ptr = Box::into_raw(boxed);
                        ptr
                    })
                    .collect();
                let boxed = Box::new(items);
                Box::into_raw(boxed)
            }
            _ => std::ptr::null_mut(),
        }
    }
}

#[no_mangle]
pub extern "C" fn diff_event_get_map_delta(ptr: *mut Diff) -> *mut MapDelta {
    unsafe {
        let diff = &*ptr;
        match diff {
            Diff::Map(map) => {
                let boxed = Box::new(map.clone());
                Box::into_raw(boxed)
            }
            _ => std::ptr::null_mut(),
        }
    }
}

#[no_mangle]
pub extern "C" fn diff_event_get_tree_diff(ptr: *mut Diff) -> *mut TreeDiff {
    unsafe {
        let diff = &*ptr;
        match diff {
            Diff::Tree(tree) => {
                let tree = tree.clone().into_owned();
                let boxed = Box::new(tree);
                Box::into_raw(boxed)
            }
            _ => std::ptr::null_mut(),
        }
    }
}

#[no_mangle]
pub extern "C" fn diff_event_get_type(ptr: *mut Diff) -> i32 {
    unsafe {
        let diff = &*ptr;
        match diff {
            Diff::List(..) => 0,
            Diff::Text(..) => 1,
            Diff::Map(..) => 2,
            Diff::Tree(..) => 3,
            Diff::Unknown => 4,
        }
    }
}
