use std::ffi::{c_schar, CStr};

use loro::{EncodedBlobMode, Frontiers, LoroDoc, ValueOrContainer, VersionVector};

#[no_mangle]
pub extern "C" fn loro_doc_decode_import_blob_meta(
    blob: *mut Vec<u8>,
    check_checksum: i32,
    err: *mut u8,
    psvv: *mut *mut VersionVector,
    pevv: *mut *mut VersionVector,
    sf: *mut *mut Frontiers,
    mode: *mut u8,
    start_timestamp: *mut i64,
    end_timestamp: *mut i64,
    change_num: *mut u32,
) {
    unsafe {
        let blob = &*blob;
        let check_checksum = if check_checksum == 0 { true } else { false };
        let result = LoroDoc::decode_import_blob_meta(blob, check_checksum);
        match result {
            Ok(meta) => {
                *mode = match meta.mode {
                    EncodedBlobMode::Snapshot => 0,
                    EncodedBlobMode::OutdatedSnapshot => 1,
                    EncodedBlobMode::ShallowSnapshot => 2,
                    EncodedBlobMode::OutdatedRle => 3,
                    EncodedBlobMode::Updates => 4,
                };
                *psvv = Box::into_raw(Box::new(meta.partial_start_vv.clone()));
                *pevv = Box::into_raw(Box::new(meta.partial_end_vv.clone()));
                *sf = Box::into_raw(Box::new(meta.start_frontiers.clone()));
                *start_timestamp = meta.start_timestamp;
                *end_timestamp = meta.end_timestamp;
                *change_num = meta.change_num;
            }
            Err(..) => {
                *err = 1;
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn loro_doc_get_by_path(
    doc: *mut LoroDoc,
    path_ptr: *const c_schar,
) -> *mut ValueOrContainer {
    unsafe {
        let doc = &mut *doc;
        let path = CStr::from_ptr(path_ptr).to_string_lossy();
        let result = doc.get_by_str_path(&path);
        match result {
            Some(value) => {
                let value = Box::into_raw(Box::new(value));
                value
            }
            None => std::ptr::null_mut(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_get_by_path() {
        let doc = LoroDoc::new();
        let m = doc.get_map("root");
        let t = doc.get_text("textField");
        m.insert_container("textField", t).unwrap();
        let result = doc.get_by_str_path("root/textField");
        assert!(result.is_some());
    }
}
