#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

typedef struct ffiapi_string {
  size_t len;
  char* data;
} ffiapi_string;

typedef struct ffiapi_output_file {
  struct ffiapi_string path;
  void* data;
} ffiapi_output_file;

typedef struct ffiapi_message {
  struct ffiapi_string file;
  int line;
  int column;
  int length;
  struct ffiapi_string text;
} ffiapi_message;

typedef struct ffiapi_engine {
  uint8_t name;
  _GoString_ version;
} ffiapi_engine;

typedef struct ffiapi_define {
  _GoString_ from;
  _GoString_ to;
} ffiapi_define;

typedef void* (*allocator) (size_t bytes);

ffiapi_string create_ffiapi_string(allocator alloc, _GoString_ gostr);

ffiapi_string create_ffiapi_string_from_bytes(allocator alloc, size_t len, void* godata);

ffiapi_message* create_ffiapi_message_array(allocator alloc, size_t len);

void set_ffiapi_message_array_element(allocator alloc, ffiapi_message* array, size_t i, _GoString_ file, int line, int column, int length, _GoString_ text);

ffiapi_engine get_ffiapi_engine_array_element(ffiapi_engine* array, size_t i);

typedef void (*transform_api_callback) (
  void* cb_data,
  ffiapi_string js,
  ffiapi_message* errors,
  ffiapi_message* warnings
);

void call_transform_api_callback(
  transform_api_callback f,
  void* cb_data,
  ffiapi_string js,
  ffiapi_message* errors,
  ffiapi_message* warnings
);
