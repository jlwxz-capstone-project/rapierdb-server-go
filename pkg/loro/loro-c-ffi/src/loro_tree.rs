use loro::{Container, LoroTree};

#[no_mangle]
pub extern "C" fn destroy_loro_tree(ptr: *mut LoroTree) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn loro_tree_to_container(ptr: *mut LoroTree) -> *mut Container {
    unsafe {
        let tree = &mut *ptr;
        let container = Container::Tree(tree.clone());
        let boxed = Box::new(container);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn loro_tree_is_attached(ptr: *const LoroTree) -> i32 {
    unsafe {
        let tree = &*ptr;
        if tree.is_attached() {
            1
        } else {
            0
        }
    }
}
