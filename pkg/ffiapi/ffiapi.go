package main

/*
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <string.h>

// A string with data allocated and owned by C, not Go.
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
	ptrdiff_t line;
	ptrdiff_t column;
	ptrdiff_t length;
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

typedef void (*transform_api_callback) (
	void* cb_data,
	size_t out_len,
	ffiapi_message* errors,
	size_t errors_len,
	ffiapi_message* warnings,
	size_t warnings_len
);

static inline ffiapi_string create_ffiapi_string(allocator alloc, _GoString_ gostr) {
	size_t len = gostr.n;
	char const* godata = gostr.p;
	char* data = (char*) alloc(len);
	memcpy(data, godata, len);
	struct ffiapi_string str = {
		.len = len,
		.data = data,
	};
	return str;
}

static inline ffiapi_string create_ffiapi_string_from_bytes(allocator alloc, size_t len, void* godata) {
	char* data = (char*) alloc(len);
	memcpy(data, godata, len);
	struct ffiapi_string str = {
		.len = len,
		.data = data,
	};
	return str;
}

static inline ffiapi_message* create_ffiapi_message_array(allocator alloc, size_t len) {
	return alloc(sizeof(ffiapi_message) * len);
}

static inline void set_ffiapi_message_array_element(allocator alloc, ffiapi_message* array, size_t i, ffiapi_string file, ptrdiff_t line, ptrdiff_t column, ptrdiff_t length, ffiapi_string text) {
	ffiapi_message msg = {
		.file = file,
		.line = line,
		.column = column,
		.length = length,
		.text = text,
	};
	array[i] = msg;
}

static inline ffiapi_engine get_ffiapi_engine_array_element(ffiapi_engine* array, size_t i) {
	return array[i];
}

static inline ffiapi_define get_ffiapi_define_array_element(ffiapi_define* array, size_t i) {
	return array[i];
}

static inline void call_transform_api_callback(
	transform_api_callback f,
	void* cb_data,
	size_t out_len,
	ffiapi_message* errors,
	size_t errors_len,
	ffiapi_message* warnings,
	size_t warnings_len
) {
	f(cb_data, out_len, errors, errors_len, warnings, warnings_len);
}
*/
import "C"
import (
	"github.com/evanw/esbuild/pkg/api"
	"unsafe"
)

func copyToCMessageArray(alloc C.allocator, messages []api.Message) (*C.ffiapi_message, C.size_t) {
	clen := C.size_t(len(messages))
	carray := C.create_ffiapi_message_array(alloc, clen)
	for i, msg := range messages {
		C.set_ffiapi_message_array_element(
			alloc,
			carray,
			C.size_t(i),
			C.create_ffiapi_string(alloc, msg.Location.File),
			C.ptrdiff_t(msg.Location.Line),
			C.ptrdiff_t(msg.Location.Column),
			C.ptrdiff_t(msg.Location.Length),
			C.create_ffiapi_string(alloc, msg.Text),
		)
	}
	return carray, clen
}

func callTransformApi(
	alloc C.allocator,
	cb C.transform_api_callback,
	cbData unsafe.Pointer,
	out unsafe.Pointer,
	code string,
	transformOptions api.TransformOptions,
) {
	goresult := api.Transform(code, transformOptions)

	// TODO Could output be larger than input?
	coutLen := C.size_t(len(goresult.JS))
	C.memcpy(out, unsafe.Pointer(&goresult.JS[0]), coutLen)

	cerrors, cerrorsLen := copyToCMessageArray(alloc, goresult.Errors)
	cwarnings, cwarningsLen := copyToCMessageArray(alloc, goresult.Warnings)

	C.call_transform_api_callback(cb, cbData, coutLen, cerrors, cerrorsLen, cwarnings, cwarningsLen)
}

//export GoTransform
func GoTransform(
	alloc C.allocator,
	cb C.transform_api_callback,
	cbData unsafe.Pointer,
	out unsafe.Pointer,
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
	pureFunctions []string,

	sourceFile string,
	loader C.uint8_t,
) {
	goenginesLen := int(enginesLen)
	goengines := make([]api.Engine, goenginesLen)
	for i := 0; i < goenginesLen; i++ {
		engine := C.get_ffiapi_engine_array_element(engines, C.size_t(i))
		goengines[i] = api.Engine{
			Name:    api.EngineName(engine.name),
			Version: engine.version,
		}
	}
	godefines := make(map[string]string)
	godefinesLen := int(definesLen)
	for i := 0; i < godefinesLen; i++ {
		define := C.get_ffiapi_define_array_element(defines, C.size_t(i))
		godefines[define.from] = define.to
	}
	go callTransformApi(alloc, cb, cbData, out, code, api.TransformOptions{
		Sourcemap: api.SourceMap(sourceMap),
		Target:    api.Target(target),
		Engines:   goengines,
		Strict: api.StrictOptions{
			NullishCoalescing: bool(strictNullishCoalescing),
			ClassFields:       bool(strictClassFields),
		},

		MinifyWhitespace:  bool(minifyWhitespace),
		MinifyIdentifiers: bool(minifyIdentifiers),
		MinifySyntax:      bool(minifySyntax),

		JSXFactory:  jsxFactory,
		JSXFragment: jsxFragment,

		Defines:       godefines,
		PureFunctions: pureFunctions,

		Sourcefile: sourceFile,
		Loader:     api.Loader(loader),
	})
}

func main() {}
