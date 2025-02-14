mod loro_container;
mod loro_container_value;
mod loro_diff_event;
mod loro_list;
mod loro_list_diff_item;
mod loro_map;
mod loro_movable_list;
mod loro_text;
mod loro_text_delta;
mod loro_tree;
mod loro_value;

use loro::event::{Diff, DiffBatch, ListDiffItem, MapDelta};
use loro::{
    ContainerID, ContainerTrait, Counter, ExportMode, Frontiers, ImportBlobMetadata,
    LoroBinaryValue, LoroDoc, LoroList, LoroListValue, LoroMap, LoroMovableList, LoroStringValue,
    LoroText, LoroValue, PeerID, TextDelta, TreeDiff, UpdateOptions, ValueOrContainer,
    VersionVector, ID,
};
use std::ffi::{CStr, CString};
use std::os::raw::c_char;

// re-exports
pub use loro_container::*;
pub use loro_container_value::*;
pub use loro_diff_event::*;
pub use loro_list::*;
pub use loro_list_diff_item::*;
pub use loro_map::*;
pub use loro_movable_list::*;
pub use loro_text::*;
pub use loro_text_delta::*;
pub use loro_tree::*;
pub use loro_value::*;

#[no_mangle]
pub extern "C" fn create_loro_doc() -> *mut LoroDoc {
    let doc = LoroDoc::new();
    let boxed = Box::new(doc);
    let ptr = Box::into_raw(boxed);
    ptr
}

#[no_mangle]
pub extern "C" fn destroy_loro_doc(ptr: *mut LoroDoc) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn get_text(doc_ptr: *mut LoroDoc, id_ptr: *const c_char) -> *mut LoroText {
    unsafe {
        let doc = &mut *doc_ptr;
        let id = CStr::from_ptr(id_ptr).to_string_lossy().into_owned();
        let text = doc.get_text(id);
        let boxed = Box::new(text);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn get_list(doc_ptr: *mut LoroDoc, id_ptr: *const c_char) -> *mut LoroList {
    unsafe {
        let doc = &mut *doc_ptr;
        let id = CStr::from_ptr(id_ptr).to_string_lossy().into_owned();
        let list = doc.get_list(id);
        let boxed = Box::new(list);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn get_movable_list(
    doc_ptr: *mut LoroDoc,
    id_ptr: *const c_char,
) -> *mut LoroMovableList {
    unsafe {
        let doc = &mut *doc_ptr;
        let id = CStr::from_ptr(id_ptr).to_string_lossy().into_owned();
        let list = doc.get_movable_list(id);
        let boxed = Box::new(list);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn get_map(doc_ptr: *mut LoroDoc, id_ptr: *const c_char) -> *mut LoroMap {
    unsafe {
        let doc = &mut *doc_ptr;
        let id = CStr::from_ptr(id_ptr).to_string_lossy().into_owned();
        let map = doc.get_map(id);
        let boxed = Box::new(map);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn export_loro_doc_snapshot(doc_ptr: *mut LoroDoc) -> *mut Vec<u8> {
    unsafe {
        let doc = &mut *doc_ptr;
        let snapshot = doc.export(ExportMode::snapshot()).unwrap();
        let boxed = Box::new(snapshot);
        let boxed_ptr = Box::into_raw(boxed);
        boxed_ptr
    }
}

#[no_mangle]
pub extern "C" fn export_loro_doc_all_updates(doc_ptr: *mut LoroDoc) -> *mut Vec<u8> {
    unsafe {
        let doc = &mut *doc_ptr;
        let snapshot = doc.export(ExportMode::all_updates()).unwrap();
        let boxed = Box::new(snapshot);
        let boxed_ptr = Box::into_raw(boxed);
        boxed_ptr
    }
}

#[no_mangle]
pub extern "C" fn export_loro_doc_updates_from(
    doc_ptr: *mut LoroDoc,
    from_ptr: *mut VersionVector,
) -> *mut Vec<u8> {
    unsafe {
        let doc = &mut *doc_ptr;
        let from = &*from_ptr;
        let snapshot = doc.export(ExportMode::updates(from)).unwrap();
        let boxed = Box::new(snapshot);
        let boxed_ptr = Box::into_raw(boxed);
        boxed_ptr
    }
}

#[no_mangle]
pub extern "C" fn export_loro_doc_updates_till(
    doc_ptr: *mut LoroDoc,
    till_ptr: *mut VersionVector,
) -> *mut Vec<u8> {
    unsafe {
        let doc = &mut *doc_ptr;
        let till = &*till_ptr;
        let snapshot = doc.export(ExportMode::updates_till(till)).unwrap();
        let boxed = Box::new(snapshot);
        let boxed_ptr = Box::into_raw(boxed);
        boxed_ptr
    }
}

#[no_mangle]
pub extern "C" fn new_vec_from_bytes(
    data_ptr: *mut u8,
    len: usize,
    cap: usize,
    new_data_ptr: *mut *mut u8,
) -> *mut Vec<u8> {
    unsafe {
        let mut new_vec = Vec::with_capacity(cap);
        new_vec.set_len(len);
        std::ptr::copy_nonoverlapping(data_ptr, new_vec.as_mut_ptr(), len);
        let boxed = Box::new(new_vec);
        let ptr = Box::into_raw(boxed);
        *new_data_ptr = (*ptr).as_mut_ptr();
        ptr
    }
}

#[no_mangle]
pub extern "C" fn loro_doc_import(doc_ptr: *mut LoroDoc, vec_ptr: *mut Vec<u8>) {
    unsafe {
        let doc = &mut *doc_ptr;
        let vec = &mut *vec_ptr;
        doc.import(vec).unwrap();
    }
}

#[no_mangle]
pub extern "C" fn destroy_bytes_vec(ptr: *mut Vec<u8>) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn get_vec_len(ptr: *mut Vec<u8>) -> usize {
    unsafe {
        let vec = &*ptr;
        vec.len()
    }
}

#[no_mangle]
pub extern "C" fn get_vec_cap(ptr: *mut Vec<u8>) -> usize {
    unsafe {
        let vec = &*ptr;
        vec.capacity()
    }
}

#[no_mangle]
pub extern "C" fn get_vec_data(ptr: *mut Vec<u8>) -> *mut u8 {
    unsafe {
        let vec = &mut *ptr;
        vec.as_mut_ptr()
    }
}

#[no_mangle]
pub extern "C" fn new_ptr_vec() -> *mut Vec<*mut u8> {
    let vec = Vec::new();
    let boxed = Box::new(vec);
    let ptr = Box::into_raw(boxed);
    ptr
}

#[no_mangle]
pub extern "C" fn ptr_vec_push(ptr: *mut Vec<*mut u8>, value: *mut u8) {
    unsafe {
        let vec = &mut *ptr;
        vec.push(value);
    }
}

#[no_mangle]
pub extern "C" fn ptr_vec_get(ptr: *mut Vec<*mut u8>, index: usize) -> *mut u8 {
    unsafe {
        let vec = &mut *ptr;
        vec[index]
    }
}

#[no_mangle]
pub extern "C" fn destroy_ptr_vec(ptr: *mut Vec<*mut u8>) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn get_ptr_vec_len(ptr: *mut Vec<*mut u8>) -> usize {
    unsafe {
        let vec = &*ptr;
        vec.len()
    }
}

#[no_mangle]
pub extern "C" fn get_ptr_vec_cap(ptr: *mut Vec<*mut u8>) -> usize {
    unsafe {
        let vec = &*ptr;
        vec.capacity()
    }
}

#[no_mangle]
pub extern "C" fn get_ptr_vec_data(ptr: *mut Vec<*mut u8>) -> *mut *mut u8 {
    unsafe {
        let vec = &mut *ptr;
        vec.as_mut_ptr()
    }
}

#[no_mangle]
pub extern "C" fn get_oplog_vv(ptr: *mut LoroDoc) -> *mut VersionVector {
    unsafe {
        let doc = &*ptr;
        let vv = doc.oplog_vv();
        let boxed = Box::new(vv);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn get_state_vv(ptr: *mut LoroDoc) -> *mut VersionVector {
    unsafe {
        let doc = &*ptr;
        let vv = doc.state_vv();
        let boxed = Box::new(vv);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn destroy_vv(ptr: *mut VersionVector) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn vv_to_frontiers(
    doc_ptr: *mut LoroDoc,
    vv_ptr: *mut VersionVector,
) -> *mut Frontiers {
    unsafe {
        let doc = &*doc_ptr;
        let vv = &*vv_ptr;
        let frontiers = doc.vv_to_frontiers(vv);
        let boxed = Box::new(frontiers);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn frontiers_to_vv(
    doc_ptr: *mut LoroDoc,
    frontiers_ptr: *mut Frontiers,
) -> *mut VersionVector {
    unsafe {
        let doc = &*doc_ptr;
        let frontiers = &*frontiers_ptr;
        let vv = doc.frontiers_to_vv(frontiers).unwrap();
        let boxed = Box::new(vv);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn get_oplog_frontiers(ptr: *mut LoroDoc) -> *mut Frontiers {
    unsafe {
        let doc = &*ptr;
        let frontiers = doc.oplog_frontiers();
        let boxed = Box::new(frontiers);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn get_state_frontiers(ptr: *mut LoroDoc) -> *mut Frontiers {
    unsafe {
        let doc = &*ptr;
        let frontiers = doc.state_frontiers();
        let boxed = Box::new(frontiers);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn destroy_frontiers(ptr: *mut Frontiers) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn encode_frontiers(ptr: *mut Frontiers) -> *mut Vec<u8> {
    unsafe {
        let frontiers = &*ptr;
        let encoded = frontiers.encode();
        let boxed = Box::new(encoded);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn encode_vv(ptr: *mut VersionVector) -> *mut Vec<u8> {
    unsafe {
        let vv = &*ptr;
        let encoded = vv.encode();
        let boxed = Box::new(encoded);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn decode_frontiers(ptr: *mut Vec<u8>) -> *mut Frontiers {
    unsafe {
        let encoded = &*ptr;
        let frontiers = Frontiers::decode(encoded).unwrap();
        let boxed = Box::new(frontiers);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn decode_vv(ptr: *mut Vec<u8>) -> *mut VersionVector {
    unsafe {
        let encoded = &*ptr;
        let vv = VersionVector::decode(encoded).unwrap();
        let boxed = Box::new(vv);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn get_frontiers_len(ptr: *mut Frontiers) -> usize {
    unsafe {
        let frontiers = &*ptr;
        frontiers.len()
    }
}

#[no_mangle]
pub extern "C" fn fork_doc(doc_ptr: *mut LoroDoc) -> *mut LoroDoc {
    unsafe {
        let doc = &mut *doc_ptr;
        let forked = doc.fork();
        let boxed = Box::new(forked);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn fork_doc_at(
    doc_ptr: *mut LoroDoc,
    frontiers_ptr: *mut Frontiers,
) -> *mut LoroDoc {
    unsafe {
        let doc = &mut *doc_ptr;
        let frontiers = &*frontiers_ptr;
        let forked = doc.fork_at(frontiers);
        let boxed = Box::new(forked);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[repr(C, packed)]
#[derive(PartialEq, Eq, Hash, Clone, Copy)]
pub struct CLayoutID {
    pub peer: PeerID,
    pub counter: Counter,
}

impl From<ID> for CLayoutID {
    fn from(id: ID) -> Self {
        Self {
            peer: id.peer,
            counter: id.counter,
        }
    }
}

impl From<CLayoutID> for ID {
    fn from(id: CLayoutID) -> Self {
        Self {
            peer: id.peer,
            counter: id.counter,
        }
    }
}

#[no_mangle]
pub extern "C" fn frontiers_new_empty() -> *mut Frontiers {
    let frontiers = Frontiers::new();
    let boxed = Box::new(frontiers);
    let ptr = Box::into_raw(boxed);
    ptr
}

#[no_mangle]
pub extern "C" fn vv_new_empty() -> *mut VersionVector {
    let vv = VersionVector::new();
    let boxed = Box::new(vv);
    let ptr = Box::into_raw(boxed);
    ptr
}

#[no_mangle]
pub extern "C" fn frontiers_contains(ptr: *mut Frontiers, id_ptr: *const CLayoutID) -> i32 {
    unsafe {
        let frontiers = &*ptr;
        let id: ID = (*id_ptr).into();
        if frontiers.contains(&id) {
            1
        } else {
            0
        }
    }
}

#[no_mangle]
pub extern "C" fn frontiers_push(ptr: *mut Frontiers, id_ptr: *const CLayoutID) {
    unsafe {
        let frontiers = &mut *ptr;
        let id: ID = (*id_ptr).into();
        frontiers.push(id);
    }
}

#[no_mangle]
pub extern "C" fn frontiers_remove(ptr: *mut Frontiers, id_ptr: *const CLayoutID) {
    unsafe {
        let frontiers = &mut *ptr;
        let id: ID = (*id_ptr).into();
        frontiers.remove(&id);
    }
}

#[no_mangle]
pub extern "C" fn diff_loro_doc(
    doc_ptr: *mut LoroDoc,
    v1_ptr: *mut Frontiers,
    v2_ptr: *mut Frontiers,
) -> *mut DiffBatch {
    unsafe {
        let doc1 = &*doc_ptr;
        let v1 = &*v1_ptr;
        let v2 = &*v2_ptr;
        let diff = doc1.diff(v1, v2).unwrap();
        let boxed = Box::new(diff);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn destroy_diff_batch(ptr: *mut DiffBatch) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn diff_batch_events(
    ptr: *mut DiffBatch,
    cids_ptr: *mut *mut Vec<*mut u8>,
    events_ptr: *mut *mut Vec<*mut u8>,
) {
    unsafe {
        let diff = &*ptr;
        let mut cids = Box::new(Vec::new());
        let mut events = Box::new(Vec::new());
        for (cid, event) in diff.iter() {
            let cid = Box::into_raw(Box::new(cid.clone()));
            let event = Box::into_raw(Box::new(event.clone()));
            cids.push(cid);
            events.push(event);
        }
        // make rust happy
        *cids_ptr = Box::into_raw(cids) as *mut _;
        *events_ptr = Box::into_raw(events) as *mut _;
    }
}

#[no_mangle]
pub extern "C" fn destroy_container_id(ptr: *mut ContainerID) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn is_container_id_root(ptr: *mut ContainerID) -> i32 {
    unsafe {
        let cid = &*ptr;
        match cid {
            ContainerID::Root { .. } => 1,
            _ => 0,
        }
    }
}

#[no_mangle]
pub extern "C" fn is_container_id_normal(ptr: *mut ContainerID) -> i32 {
    unsafe {
        let cid = &*ptr;
        match cid {
            ContainerID::Normal { .. } => 1,
            _ => 0,
        }
    }
}

#[no_mangle]
pub extern "C" fn container_id_root_name(ptr: *mut ContainerID) -> *mut c_char {
    unsafe {
        let cid = &*ptr;
        match cid {
            ContainerID::Root { name, .. } => {
                let ptr = CString::new(name.as_str()).unwrap().into_raw();
                ptr
            }
            _ => panic!("ContainerID is not root"),
        }
    }
}

#[no_mangle]
pub extern "C" fn container_id_normal_peer(ptr: *mut ContainerID) -> PeerID {
    unsafe {
        let cid = &*ptr;
        match cid {
            ContainerID::Normal { peer, .. } => *peer,
            _ => panic!("ContainerID is not normal"),
        }
    }
}

#[no_mangle]
pub extern "C" fn container_id_normal_counter(ptr: *mut ContainerID) -> Counter {
    unsafe {
        let cid = &*ptr;
        match cid {
            ContainerID::Normal { counter, .. } => *counter,
            _ => panic!("ContainerID is not normal"),
        }
    }
}

#[no_mangle]
pub extern "C" fn container_id_container_type(ptr: *mut ContainerID) -> u8 {
    unsafe {
        let cid = &*ptr;
        cid.container_type().to_u8()
    }
}

#[no_mangle]
pub extern "C" fn destroy_map_delta(ptr: *mut MapDelta) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn destroy_tree_diff(ptr: *mut TreeDiff) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn loro_list_to_value(ptr: *mut LoroList) -> *mut LoroValue {
    unsafe {
        let list = &*ptr;
        let value = list.get_value();
        let boxed = Box::new(value);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn loro_map_to_value(ptr: *mut LoroMap) -> *mut LoroValue {
    unsafe {
        let map = &*ptr;
        let value = map.get_value();
        let boxed = Box::new(value);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn loro_movable_list_to_value(ptr: *mut LoroMovableList) -> *mut LoroValue {
    unsafe {
        let list = &*ptr;
        let value = list.get_value();
        let boxed = Box::new(value);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[no_mangle]
pub extern "C" fn loro_text_to_value(ptr: *mut LoroText) -> *mut LoroValue {
    unsafe {
        let text = &*ptr;
        let value = text.get_richtext_value();
        let boxed = Box::new(value);
        let ptr = Box::into_raw(boxed);
        ptr
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_create_and_edit_many_loro_docs() {
        let time_start = std::time::Instant::now();
        let mut docs = Vec::new();
        for i in 0..100000 {
            let doc = LoroDoc::new();
            let text = doc.get_text("test");
            let content = format!("Hello, world! {}", i);
            text.update(&content, UpdateOptions::default()).unwrap();
            let snapshot = doc.export(ExportMode::snapshot()).unwrap();
            let doc2 = LoroDoc::new();
            doc2.import(&snapshot).unwrap();
            docs.push(doc2);
        }
        let time_end = std::time::Instant::now();
        println!("Time taken: {:?}", time_end.duration_since(time_start));
    }

    #[test]
    fn test_loro_diff() {
        let time_start = std::time::Instant::now();
        let mut cids = Vec::new();
        let mut events = Vec::new();
        for _ in 0..100000 {
            let doc = LoroDoc::new();
            let f1 = doc.state_frontiers();
            let text = doc.get_text("test");
            text.update("Hello, world!", UpdateOptions::default())
                .unwrap();
            let f2 = doc.state_frontiers();
            let diff = doc.diff(&f1, &f2).unwrap();
            for (cid, event) in diff.iter() {
                cids.push(cid.clone());
                events.push(event.clone());
            }
        }
        println!("cids: {:?}", cids.len());
        println!("events: {:?}", events.len());
        let time_end = std::time::Instant::now();
        println!("Time taken: {:?}", time_end.duration_since(time_start));
    }

    #[test]
    fn test_heap_alloc() {
        let arr = [0; 1000000];
        let boxed = Box::new(arr);
        println!("arr len: {}", boxed.len());
    }
}
