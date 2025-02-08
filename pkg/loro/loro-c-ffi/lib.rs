use loro::{
    Counter, ExportMode, Frontiers, LoroDoc, LoroList, LoroMap, LoroMovableList, LoroText, PeerID,
    UpdateOptions, VersionVector, ID,
};
use std::ffi::{CStr, CString};
use std::os::raw::c_char;

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
pub extern "C" fn destroy_loro_text(ptr: *mut LoroText) {
    unsafe {
        let _ = Box::from_raw(ptr);
    }
}

#[no_mangle]
pub extern "C" fn loro_text_to_string(text_ptr: *mut LoroText) -> *mut c_char {
    unsafe {
        let text = &mut *text_ptr;
        let s = text.to_string();
        let ptr = CString::new(s).unwrap().into_raw();
        ptr
    }
}

#[no_mangle]
pub extern "C" fn update_loro_text(text_ptr: *mut LoroText, content: *const c_char) {
    unsafe {
        let text = &mut *text_ptr;
        let s = CStr::from_ptr(content).to_string_lossy().into_owned();
        text.update(&s, UpdateOptions::default()).unwrap();
    }
}

#[no_mangle]
pub extern "C" fn insert_loro_text(text_ptr: *mut LoroText, pos: usize, content: *const c_char) {
    unsafe {
        let text = &mut *text_ptr;
        let s = CStr::from_ptr(content).to_string_lossy().into_owned();
        text.insert(pos, &s).unwrap();
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
pub extern "C" fn destroy_loro_list(ptr: *mut LoroList) {
    unsafe {
        let _ = Box::from_raw(ptr);
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
pub extern "C" fn destroy_loro_movable_list(ptr: *mut LoroMovableList) {
    unsafe {
        let _ = Box::from_raw(ptr);
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
pub extern "C" fn destroy_loro_map(ptr: *mut LoroMap) {
    unsafe {
        let _ = Box::from_raw(ptr);
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
}
