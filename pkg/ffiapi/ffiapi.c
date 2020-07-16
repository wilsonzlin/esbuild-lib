#include "ffiapi.h"

ffiapi_string create_ffiapi_string(allocator alloc, _GoString_ gostr) {
  size_t len = _GoStringLen(gostr);
  char* godata = _GoStringPtr(gostr);
  char* data = (char*) alloc(len);
  memcpy(data, godata, len);
  struct ffiapi_string str = {
    .len = len,
    .data = data,
  };
  return str;
}

ffiapi_string create_ffiapi_string_from_bytes(allocator alloc, size_t len, void* godata) {
  char* data = (char*) alloc(len);
  memcpy(data, godata, len);
  struct ffiapi_string str = {
    .len = len,
    .data = data,
  };
  return str;
}

ffiapi_message* create_ffiapi_message_array(allocator alloc, size_t len) {
  return alloc(sizeof(ffiapi_message) * len);
}

void set_ffiapi_message_array_element(allocator alloc, ffiapi_message* array, size_t i, _GoString_ file, int line, int column, int length, _GoString_ text) {
  array[i] = ffiapi_message {
    .file = create_ffiapi_string(alloc, file),
    .line = line,
    .column = column,
    .length = length,
    .text = create_ffiapi_string(alloc, text),
  };
}

ffiapi_engine get_ffiapi_engine_array_element(ffiapi_engine* array, size_t i) {
  return array[i];
}

void call_transform_api_callback(
  transform_api_callback f,
  void* cb_data,
  ffiapi_string js,
  ffiapi_message* errors,
  ffiapi_message* warnings
) {
  f(cb_data, js, errors, warnings);
}
