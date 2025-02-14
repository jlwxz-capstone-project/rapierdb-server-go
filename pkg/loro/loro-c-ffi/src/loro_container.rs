use loro::{Container, LoroList, LoroMap, LoroMovableList, LoroText, LoroTree};

#[no_mangle]
pub extern "C" fn destroy_loro_container(ptr: *mut Container) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn loro_container_get_type(ptr: *mut Container) -> u8 {
    unsafe {
        let container = &*ptr;
        match container {
            Container::List(..) => 0,
            Container::Map(..) => 1,
            Container::Text(..) => 2,
            Container::MovableList(..) => 3,
            Container::Tree(..) => 4,
            Container::Unknown(..) => 5,
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_container_get_list(ptr: *mut Container) -> *mut LoroList {
    unsafe {
        let container = &*ptr;
        match container {
            Container::List(list) => {
                let boxed = Box::new(list.clone());
                Box::into_raw(boxed)
            }
            _ => std::ptr::null_mut(),
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_container_get_map(ptr: *mut Container) -> *mut LoroMap {
    unsafe {
        let container = &*ptr;
        match container {
            Container::Map(map) => {
                let boxed = Box::new(map.clone());
                Box::into_raw(boxed)
            }
            _ => std::ptr::null_mut(),
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_container_get_text(ptr: *mut Container) -> *mut LoroText {
    unsafe {
        let container = &*ptr;
        match container {
            Container::Text(text) => {
                let boxed = Box::new(text.clone());
                Box::into_raw(boxed)
            }
            _ => std::ptr::null_mut(),
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_container_get_movable_list(ptr: *mut Container) -> *mut LoroMovableList {
    unsafe {
        let container = &*ptr;
        match container {
            Container::MovableList(list) => {
                let boxed = Box::new(list.clone());
                Box::into_raw(boxed)
            }
            _ => std::ptr::null_mut(),
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_container_get_tree(ptr: *mut Container) -> *mut LoroTree {
    unsafe {
        let container = &*ptr;
        match container {
            Container::Tree(tree) => {
                let boxed = Box::new(tree.clone());
                Box::into_raw(boxed)
            }
            _ => std::ptr::null_mut(),
        }
    }
}
