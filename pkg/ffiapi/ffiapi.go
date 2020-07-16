package main

// #include "ffiapi.h"
import "C"
import (
  "github.com/evanw/esbuild/pkg/api"
  "unsafe"
)

func copyToCMessageArray(messages []Message) {
  carray = C.create_ffiapi_message_array(alloc, C.size_t(len(goresult.Errors)))
  for i, err := range goresult.Errors {
    C.set_ffiapi_message_array_element(alloc, carray, C.size_t(i), err.Location.File, C.int(err.Location.Line), C.int(err.Location.Col), C.int(err.Location.Length), err.Text)
  }
  return carray
}

func callTransformApi(
  alloc C.allocator,
  cb C.transform_api_callback,
  cbData unsafe.Pointer,
  code string,
  transformOptions api.TransformOptions,
) {
  goresult := api.Transform(code, transformOptions)

  cjs = C.create_ffiapi_string_from_bytes(alloc, C.size_t(len(goresult.JS)), unsafe.Pointer(&goresult.JS[0]))
  cerrors = copyToCMessageArray(goresult.Errors)
  cwarnings = copyToCMessageArray(goresult.Warnings)

  C.call_transform_api_callback(cb, cbData, cjs, cerrors, cwarnings)
}

//export GoTransform
func GoTransform(
  alloc C.allocator,
  cb C.transform_api_callback,
  cbData unsafe.Pointer,
  code string,

  sourceMap C.uint8_t,
  target C.uint8_t,
  engines *C.ffiapi_engine,
  enginesLen C.size_t,
  strictNullishCoalescing C.bool,
  strictClassFields C.bool,

  minifyWhitespace C.bool,
  minifyIdentifiers C.bool,
  minifySyntax C.bool,

  jsxFactory string,
  jsxFragment string,

  defines *C.ffiapi_define,
  definesLen C.size_t,
  pureFunctions *string,
  pureFunctionsLen C.size_t,

  sourceFile string,
  loader C.uint8_t,
) {
  goengines = make([]api.Engine, enginesLen)
  for i := 0; i < enginesLen; i++ {
    engine := C.get_ffiapi_engine_array_element(engines, C.size_t(i))
    goengines[i] = api.Engine{
      Name: engine._name,
      Version: engine._version,
    }
  }
  go callTransformApi(alloc, cb, cbData, code, api.TransformOptions{
    Sourcemap: sourceMap,
    Target: target,
    Engines: goengines,
    Strict: api.StrictOptions{
      NullishCoalescing: strictNullishCoalescing,
      ClassFields: strictClassFields,
    },

    MinifyWhitespace: minifyWhitespace,
    MinifyIdentifiers: minifyIdentifiers,
    MinifySyntax: minifySyntax,

    JSXFactory: jsxFactory,
    JSXFragment: jsxFragment,
  })
}

func main() {}
