#include <stdint.h>

extern void* create_loro_doc();
extern void* get_text(void* doc_ptr, char* id_ptr);
extern void* get_list(void* doc_ptr, char* id_ptr);
extern void* get_movable_list(void* doc_ptr, char* id_ptr);
extern void* get_map(void* doc_ptr, char* id_ptr);
extern char* loro_text_to_string(void* text_ptr);
extern void update_loro_text(void* text_ptr, char* content);
extern void* export_loro_doc_snapshot(void* doc_ptr);
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