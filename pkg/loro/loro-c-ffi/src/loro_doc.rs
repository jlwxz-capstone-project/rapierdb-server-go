use loro::{EncodedBlobMode, Frontiers, LoroDoc, VersionVector};

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
