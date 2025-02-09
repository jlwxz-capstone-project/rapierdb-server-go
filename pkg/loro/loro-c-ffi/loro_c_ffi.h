#include <stdint.h>
#include <stddef.h>

typedef struct CLayoutID {
  uint64_t peer;
  uint32_t counter;
} CLayoutID;

extern void* create_loro_doc();
extern void* get_text(void* doc_ptr, char* id_ptr);
extern void* get_list(void* doc_ptr, char* id_ptr);
extern void* get_movable_list(void* doc_ptr, char* id_ptr);
extern void* get_map(void* doc_ptr, char* id_ptr);
extern void* new_loro_text();
extern char* loro_text_to_string(void* text_ptr);
extern void update_loro_text(void* text_ptr, char* content);
extern void insert_loro_text(void* text_ptr, uint32_t pos, char* content);
extern void insert_loro_text_utf8(void* text_ptr, uint32_t pos, char* content);
extern uint32_t loro_text_length(void* text_ptr);
extern uint32_t loro_text_length_utf8(void* text_ptr);
extern void* export_loro_doc_snapshot(void* doc_ptr);
extern void* export_loro_doc_all_updates(void* doc_ptr);
extern void* export_loro_doc_updates_from(void* doc_ptr, void* from_ptr);
extern void* export_loro_doc_updates_till(void* doc_ptr, void* till_ptr);
extern void loro_doc_import(void* doc_ptr, void* vec_ptr);
extern void* new_vec_from_bytes(void* data_ptr, uint32_t len, uint32_t cap, uint8_t** new_data_ptr);
extern void destroy_loro_doc(void* ptr);
extern void destroy_loro_text(void* ptr);
extern void destroy_loro_list(void* ptr);
extern void destroy_loro_movable_list(void* ptr);
extern void destroy_loro_map(void* ptr);
extern void destroy_bytes_vec(void* ptr);
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
extern size_t get_frontiers_len(void* ptr);
extern int frontiers_contains(void* ptr, void* id_ptr);
extern void frontiers_push(void* ptr, void* id_ptr);
extern void frontiers_remove(void* ptr, void* id_ptr);
extern void* frontiers_new_empty();
extern void* vv_new_empty();
extern void* fork_doc(void* doc_ptr);
extern void* fork_doc_at(void* doc_ptr, void* frontiers_ptr);
