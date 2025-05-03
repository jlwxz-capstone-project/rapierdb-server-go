#include <stdint.h>
#include <stddef.h>

typedef struct CLayoutID {
  uint64_t peer;
  uint32_t counter;
} CLayoutID;

extern uint32_t get_vec_len(void* ptr);
extern uint32_t get_vec_cap(void* ptr);
extern void* get_vec_data(void* ptr);
extern void* get_oplog_vv(void* ptr);
extern void* get_state_vv(void* ptr);
extern void destroy_vv(void* ptr);
extern void* get_oplog_frontiers(void* ptr);
extern void* get_state_frontiers(void* ptr);
extern void destroy_frontiers(void* ptr);
extern void* encode_frontiers(void* ptr);
extern void* encode_vv(void* ptr);
extern void* decode_frontiers(void* ptr);
extern void* decode_vv(void* ptr);
extern void* frontiers_to_vv(void* doc_ptr, void* frontiers_ptr);
extern void* vv_to_frontiers(void* doc_ptr, void* vv_ptr);
extern uint32_t get_frontiers_len(void* ptr);
extern int frontiers_contains(void* ptr, void* id_ptr);
extern void frontiers_push(void* ptr, void* id_ptr);
extern void frontiers_remove(void* ptr, void* id_ptr);
extern void* frontiers_new_empty();
extern void* vv_new_empty();
extern void* fork_doc(void* doc_ptr);
extern void* fork_doc_at(void* doc_ptr, void* frontiers_ptr);
extern void* diff_loro_doc(void* doc_ptr, void* v1_ptr, void* v2_ptr);
extern void destroy_diff_batch(void* ptr);
extern void diff_batch_events(void* ptr, void** cids_ptr, void** events_ptr);
extern void destroy_container_id(void* ptr);
extern int is_container_id_root(void* ptr);
extern int is_container_id_normal(void* ptr);
extern char* container_id_root_name(void* ptr);
extern uint64_t container_id_normal_peer(void* ptr);
extern uint32_t container_id_normal_counter(void* ptr);
extern uint8_t container_id_container_type(void* ptr);

extern void destroy_text_delta(void* ptr);
extern void destroy_map_delta(void* ptr);
extern void destroy_tree_diff(void* ptr);
extern int diff_event_get_type(void* ptr);

extern int vv_partial_cmp(void* ptr1, void* ptr2);

// Loro Doc
extern void* create_loro_doc();
extern void destroy_loro_doc(void* ptr);
extern void* get_text(void* doc_ptr, char* id_ptr);
extern void* get_list(void* doc_ptr, char* id_ptr);
extern void* get_movable_list(void* doc_ptr, char* id_ptr);
extern void* get_map(void* doc_ptr, char* id_ptr);
extern void* export_loro_doc_snapshot(void* doc_ptr);
extern void* export_loro_doc_all_updates(void* doc_ptr);
extern void* export_loro_doc_updates_from(void* doc_ptr, void* from_ptr);
extern void* export_loro_doc_updates_till(void* doc_ptr, void* till_ptr);
extern void* loro_doc_import(void* doc_ptr, void* vec_ptr);
extern void loro_doc_decode_import_blob_meta(
  void* blob,
  int check_checksum,
  uint8_t* err,
  void* psvv,
  void* pevv,
  void* sf,
  uint8_t* mode,
  int64_t* start_timestamp,
  int64_t* end_timestamp,
  uint32_t* change_num
);
extern void* loro_doc_get_by_path(void* doc_ptr, char* path_ptr);

// Loro Import Status
extern void destroy_import_status(void* ptr);
extern void* import_status_get_success(void* ptr);
extern void* import_status_get_pending(void* ptr);

// Loro Version Range
extern void destroy_version_range(void* ptr);
extern int version_range_is_empty(void* ptr);

// Loro List Diff Item
extern void destroy_list_diff_item(void* ptr);
extern int list_diff_item_get_type(void* ptr);
extern void* list_diff_item_get_insert(void* ptr, uint8_t* is_move_ptr, uint8_t* err);
extern uint32_t list_diff_item_get_delete_count(void* ptr, uint8_t* err);
extern uint32_t list_diff_item_get_retain_count(void* ptr, uint8_t* err);

// Loro Diff Event
extern void destroy_diff_event(void* ptr);
extern void* diff_event_get_list_diff(void* ptr);
extern void* diff_event_get_text_delta(void* ptr);
extern void* diff_event_get_map_delta(void* ptr);
extern void* diff_event_get_tree_diff(void* ptr);

// Rust Bytes Vec
extern void* new_vec_from_bytes(void* data_ptr, uint32_t len, uint32_t cap, uint8_t** new_data_ptr);
extern void destroy_bytes_vec(void* ptr);

// Rust Ptr Vec
extern void* new_ptr_vec();
extern void ptr_vec_push(void* ptr, void* value);
extern void* ptr_vec_get(void* ptr, uint32_t index);
extern void destroy_ptr_vec(void* ptr);
extern uint32_t get_ptr_vec_len(void* ptr);
extern uint32_t get_ptr_vec_cap(void* ptr);
extern void* get_ptr_vec_data(void* ptr);

// Loro Text
extern void* new_loro_text();
extern void destroy_loro_text(void* ptr);
extern char* loro_text_to_string(void* text_ptr, uint8_t* err);
extern void update_loro_text(void* text_ptr, char* content, uint8_t* err);
extern void insert_loro_text(void* text_ptr, uint32_t pos, char* content, uint8_t* err);
extern void insert_loro_text_utf8(void* text_ptr, uint32_t pos, char* content, uint8_t* err);
extern uint32_t loro_text_length(void* text_ptr);
extern uint32_t loro_text_length_utf8(void* text_ptr);
extern void* loro_text_to_container(void* ptr);
extern int loro_text_is_attached(void* ptr);

// Loro Map
extern void* loro_map_new_empty();
extern uint32_t loro_map_len(void* ptr);
extern void destroy_loro_map(void* ptr);
extern void* loro_map_get(void* ptr, char* key_ptr);
extern void loro_map_get_null(void* ptr, char* key_ptr, uint8_t* err);
extern int loro_map_get_bool(void* ptr, char* key_ptr, uint8_t* err);
extern double loro_map_get_double(void* ptr, char* key_ptr, uint8_t* err);
extern int64_t loro_map_get_i64(void* ptr, char* key_ptr, uint8_t* err);
extern char* loro_map_get_string(void* ptr, char* key_ptr, uint8_t* err);
extern void* loro_map_get_text(void* ptr, char* key_ptr, uint8_t* err);
extern void* loro_map_get_list(void* ptr, char* key_ptr, uint8_t* err);
extern void* loro_map_get_movable_list(void* ptr, char* key_ptr, uint8_t* err);
extern void* loro_map_get_map(void* ptr, char* key_ptr, uint8_t* err);
extern void loro_map_insert_null(void* ptr, char* key_ptr, uint8_t* err);
extern void loro_map_insert_bool(void* ptr, char* key_ptr, int bool_value, uint8_t* err);
extern void loro_map_insert_double(void* ptr, char* key_ptr, double double_value, uint8_t* err);
extern void loro_map_insert_i64(void* ptr, char* key_ptr, int64_t int_value, uint8_t* err);
extern void loro_map_insert_string(void* ptr, char* key_ptr, char* str_value, uint8_t* err);
extern void* loro_map_insert_text(void* ptr, char* key_ptr, void* text_ptr, uint8_t* err);
extern void* loro_map_insert_list(void* ptr, char* key_ptr, void* list_ptr, uint8_t* err);
extern void* loro_map_insert_movable_list(void* ptr, char* key_ptr, void* list_ptr, uint8_t* err);
extern void* loro_map_insert_map(void* ptr, char* key_ptr, void* map_ptr, uint8_t* err);
extern void* loro_map_to_container(void* ptr);
extern int loro_map_is_attached(void* ptr);
extern void* loro_map_get_items(void* ptr);
extern void* loro_map_insert_value(void* ptr, char* key_ptr, void* value_ptr, uint8_t* err);
extern void* loro_map_insert_container(void* ptr, char* key_ptr, void* value_ptr, uint8_t* err);

// Loro List
extern void* loro_list_new_empty();
extern void destroy_loro_list(void* ptr);
extern void loro_list_push_value(void* ptr, void* value_ptr, uint8_t* err);
extern void* loro_list_push_container(void* ptr, void* container_ptr, uint8_t* err);
extern void loro_list_push_null(void* ptr, uint8_t* err);
extern void loro_list_push_bool(void* ptr, int value, uint8_t* err);
extern void loro_list_push_double(void* ptr, double value, uint8_t* err);
extern void loro_list_push_i64(void* ptr, int64_t value, uint8_t* err);
extern void loro_list_push_string(void* ptr, char* value, uint8_t* err);
extern void* loro_list_push_text(void* ptr, void* text_ptr, uint8_t* err);
extern void* loro_list_push_list(void* ptr, void* list_ptr, uint8_t* err);
extern void* loro_list_push_movable_list(void* ptr, void* list_ptr, uint8_t* err);
extern void* loro_list_push_map(void* ptr, void* map_ptr, uint8_t* err);
extern void* loro_list_get(void* ptr, uint32_t index);
extern void loro_list_get_null(void* ptr, uint32_t index, uint8_t* err);
extern int loro_list_get_bool(void* ptr, uint32_t index, uint8_t* err);
extern double loro_list_get_double(void* ptr, uint32_t index, uint8_t* err);
extern int64_t loro_list_get_i64(void* ptr, uint32_t index, uint8_t* err);
extern char* loro_list_get_string(void* ptr, uint32_t index, uint8_t* err);
extern void* loro_list_get_text(void* ptr, uint32_t index, uint8_t* err);
extern void* loro_list_get_list(void* ptr, uint32_t index, uint8_t* err);
extern void* loro_list_get_movable_list(void* ptr, uint32_t index, uint8_t* err);
extern void* loro_list_get_map(void* ptr, uint32_t index, uint8_t* err);
extern uint32_t loro_list_len(void* ptr);
extern void* loro_list_to_container(void* ptr);
extern int loro_list_is_attached(void* ptr);
extern void* loro_list_get_items(void* ptr);
extern void loro_list_insert_value(void* ptr, uint32_t index, void* value_ptr, uint8_t* err);
extern void* loro_list_insert_container(void* ptr, uint32_t index, void* container_ptr, uint8_t* err);
extern void loro_list_insert_null(void* ptr, uint32_t index, uint8_t* err);
extern void loro_list_insert_bool(void* ptr, uint32_t index, int value, uint8_t* err);
extern void loro_list_insert_double(void* ptr, uint32_t index, double value, uint8_t* err);
extern void loro_list_insert_i64(void* ptr, uint32_t index, int64_t value, uint8_t* err);
extern void loro_list_insert_string(void* ptr, uint32_t index, char* value_ptr, uint8_t* err);
extern void* loro_list_insert_text(void* ptr, uint32_t index, void* text_ptr, uint8_t* err);
extern void* loro_list_insert_list(void* ptr, uint32_t index, void* list_ptr, uint8_t* err);
extern void* loro_list_insert_movable_list(void* ptr, uint32_t index, void* movable_list_ptr, uint8_t* err);
extern void* loro_list_insert_map(void* ptr, uint32_t index, void* map_ptr, uint8_t* err);
extern void loro_list_delete(void* ptr, uint32_t pos, uint32_t len, uint8_t* err);
extern void loro_list_clear(void* ptr, uint8_t* err);

// Loro Movable List
extern void* loro_movable_list_new_empty();
extern void destroy_loro_movable_list(void* ptr);
extern uint32_t loro_movable_list_len(void* ptr);
extern void loro_movable_list_push_value(void* ptr, void* value_ptr, uint8_t* err);
extern void* loro_movable_list_push_container(void* ptr, void* container_ptr, uint8_t* err);
extern void loro_movable_list_push_null(void* ptr, uint8_t* err);
extern void loro_movable_list_push_bool(void* ptr, int value, uint8_t* err);
extern void loro_movable_list_push_double(void* ptr, double value, uint8_t* err);
extern void loro_movable_list_push_i64(void* ptr, int64_t value, uint8_t* err);
extern void loro_movable_list_push_string(void* ptr, char* value, uint8_t* err);
extern void* loro_movable_list_push_text(void* ptr, void* text_ptr, uint8_t* err);
extern void* loro_movable_list_push_list(void* ptr, void* list_ptr, uint8_t* err);
extern void* loro_movable_list_push_movable_list(void* ptr, void* movable_list_ptr, uint8_t* err);
extern void* loro_movable_list_push_map(void* ptr, void* map_ptr, uint8_t* err);
extern void* loro_movable_list_get(void* ptr, uint32_t index);
extern void loro_movable_list_get_null(void* ptr, uint32_t index, uint8_t* err);
extern int loro_movable_list_get_bool(void* ptr, uint32_t index, uint8_t* err);
extern double loro_movable_list_get_double(void* ptr, uint32_t index, uint8_t* err);
extern int64_t loro_movable_list_get_i64(void* ptr, uint32_t index, uint8_t* err);
extern char* loro_movable_list_get_string(void* ptr, uint32_t index, uint8_t* err);
extern void* loro_movable_list_get_text(void* ptr, uint32_t index, uint8_t* err);
extern void* loro_movable_list_get_list(void* ptr, uint32_t index, uint8_t* err);
extern void* loro_movable_list_get_movable_list(void* ptr, uint32_t index, uint8_t* err);
extern void* loro_movable_list_get_map(void* ptr, uint32_t index, uint8_t* err);
extern void* loro_movable_list_to_container(void* ptr);
extern int loro_movable_list_is_attached(void* ptr);
extern void* loro_movable_list_get_items(void* ptr);
extern void loro_movable_list_insert_value(void* ptr, uint32_t index, void* value_ptr, uint8_t* err);
extern void* loro_movable_list_insert_container(void* ptr, uint32_t index, void* container_ptr, uint8_t* err);
extern void loro_movable_list_delete(void* ptr, uint32_t pos, uint32_t len, uint8_t* err);
extern void loro_movable_list_move(void* ptr, uint32_t from, uint32_t to, uint8_t* err);
extern void loro_movable_list_clear(void* ptr, uint8_t* err);
extern void loro_movable_list_set_value(void* ptr, uint32_t index, void* value_ptr, uint8_t* err);
extern void* loro_movable_list_set_container(void* ptr, uint32_t index, void* container_ptr, uint8_t* err);

// Loro Value
extern void destroy_loro_value(void* ptr);
extern int loro_value_get_type(void* ptr);
extern int loro_value_get_bool(void* ptr, uint8_t* err);
extern double loro_value_get_double(void* ptr, uint8_t* err);
extern int64_t loro_value_get_i64(void* ptr, uint8_t* err);
extern const char* loro_value_get_string(void* ptr, uint8_t* err);
extern void* loro_value_get_binary(void* ptr, uint8_t* err);
extern void* loro_value_get_list(void* ptr, uint8_t* err);
extern void* loro_value_get_map(void* ptr, uint8_t* err);
extern void* loro_value_get_container_id(void* ptr, uint8_t* err);
extern void* loro_value_new_null();
extern void* loro_value_new_bool(int value);
extern void* loro_value_new_double(double value);
extern void* loro_value_new_i64(int64_t value);
extern void* loro_value_new_string(const char* value);
extern void* loro_value_new_binary(void* value);
extern void* loro_value_new_list(void* value);
extern void* loro_value_new_map(void* value);
extern char* loro_value_to_json(void* ptr);
extern void* loro_value_from_json(const char* json);

// Loro Container
extern void destroy_loro_container(void* ptr);
extern uint8_t loro_container_get_type(void* ptr);
extern void* loro_container_get_list(void* ptr);
extern void* loro_container_get_map(void* ptr);
extern void* loro_container_get_text(void* ptr);
extern void* loro_container_get_movable_list(void* ptr);
extern void* loro_container_get_tree(void* ptr);

// Loro Tree
extern void destroy_loro_tree(void* ptr);
extern void* loro_tree_to_container(void* ptr);
extern int loro_tree_is_attached(void* ptr);

// Loro Container Value
extern void destroy_loro_container_value(void* ptr);
extern uint8_t loro_container_value_get_type(void* ptr);
extern void* loro_container_value_get_container(void* ptr);
extern void* loro_container_value_get_value(void* ptr);

// Loro Binary Value
extern void* loro_binary_value_new(void* data_ptr, uint32_t len);
extern void loro_binary_value_destroy(void* ptr);
