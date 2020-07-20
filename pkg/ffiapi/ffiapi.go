package main

/*
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <string.h>

//*** Result structs. ***

typedef struct ffiapi_string {
	size_t len;
	// This memory is allocated from Go but using provided allocator, and as such owned by the other party.
	// It does NOT point to a Go string or Go memory, and must be managed.
	char* data;
} ffiapi_string;

typedef struct ffiapi_message {
	ffiapi_string file;
	ptrdiff_t line;
	ptrdiff_t column;
	ptrdiff_t length;
	ffiapi_string text;
} ffiapi_message;

typedef struct ffiapi_output_file {
	ffiapi_string path;
	ffiapi_string data;
} ffiapi_output_file;

//*** Input structs. ***

typedef struct ffiapi_define {
	_GoString_ from;
	_GoString_ to;
} ffiapi_define;

typedef struct ffiapi_engine {
	uint8_t name;
	_GoString_ version;
} ffiapi_engine;

typedef struct ffiapi_loader {
	_GoString_ name;
	uint8_t loader;
} ffiapi_loader;

//*** Input functions. ***

typedef void* (*allocator) (size_t bytes);

static inline void* call_alloc(allocator alloc, size_t bytes) {
	return alloc(bytes);
}

typedef void (*build_api_callback) (
	void* cb_data,
	ffiapi_output_file* output_files,
	size_t output_files_len,
	ffiapi_message* errors,
	size_t errors_len,
	ffiapi_message* warnings,
	size_t warnings_len
);

typedef void (*transform_api_callback) (
	void* cb_data,
	ffiapi_string js,
	ffiapi_string js_source_map,
	ffiapi_message* errors,
	size_t errors_len,
	ffiapi_message* warnings,
	size_t warnings_len
);

//*** Create ffiapi_string functions (interal Go use). ***

static inline ffiapi_string create_ffiapi_string_from_bytes(allocator alloc, size_t len, void const* godata) {
	char* data = (char*) alloc(len);
	memcpy(data, godata, len);
	struct ffiapi_string str = {
		.len = len,
		.data = data,
	};
	return str;
}

static inline ffiapi_string create_ffiapi_string(allocator alloc, _GoString_ gostr) {
	size_t len = gostr.n;
	char const* godata = gostr.p;
	return create_ffiapi_string_from_bytes(alloc, len, godata);
}

//*** Call callback functions (interal Go use). ***

static inline void call_build_api_callback(
	build_api_callback f,
	void* cb_data,
	ffiapi_output_file* output_files,
	size_t output_files_len,
	ffiapi_message* errors,
	size_t errors_len,
	ffiapi_message* warnings,
	size_t warnings_len
) {
	f(cb_data, output_files, output_files_len, errors, errors_len, warnings, warnings_len);
}

static inline void call_transform_api_callback(
	transform_api_callback f,
	void* cb_data,
	ffiapi_string js,
	ffiapi_string js_source_map,
	ffiapi_message* errors,
	size_t errors_len,
	ffiapi_message* warnings,
	size_t warnings_len
) {
	f(cb_data, js, js_source_map, errors, errors_len, warnings, warnings_len);
}
*/
import "C"
import (
	"github.com/evanw/esbuild/pkg/api"
	"unsafe"
)

func toCString(alloc C.allocator, bytes []byte) C.ffiapi_string {
	var godataPtr unsafe.Pointer = nil
	if len(bytes) > 0 {
		godataPtr = unsafe.Pointer(&bytes[0])
	}
	return C.create_ffiapi_string_from_bytes(alloc, C.size_t(len(bytes)), godataPtr)
}

func toCMessageArray(alloc C.allocator, messages []api.Message) (*C.ffiapi_message, C.size_t) {
	clen := C.size_t(len(messages))
	carray := (*C.ffiapi_message)(C.call_alloc(alloc, C.sizeof_ffiapi_message*clen))
	carraySlice := (*[1 << 30]C.ffiapi_message)(unsafe.Pointer(carray))[:clen:clen]
	for i, msg := range messages {
		loc := msg.Location
		if loc == nil {
			loc = &api.Location{}
		}
		carraySlice[i].file = C.create_ffiapi_string(alloc, loc.File)
		carraySlice[i].line = C.ptrdiff_t(loc.Line)
		carraySlice[i].column = C.ptrdiff_t(loc.Column)
		carraySlice[i].length = C.ptrdiff_t(loc.Length)
		carraySlice[i].text = C.create_ffiapi_string(alloc, msg.Text)
	}
	return carray, clen
}

func toCOutputFileArray(alloc C.allocator, outputFiles []api.OutputFile) (*C.ffiapi_output_file, C.size_t) {
	clen := C.size_t(len(outputFiles))
	carray := (*C.ffiapi_output_file)(C.call_alloc(alloc, C.sizeof_ffiapi_output_file*clen))
	carraySlice := (*[1 << 30]C.ffiapi_output_file)(unsafe.Pointer(carray))[:clen:clen]
	for i, file := range outputFiles {
		carraySlice[i].path = C.create_ffiapi_string(alloc, file.Path)
		carraySlice[i].data = toCString(alloc, file.Contents)
	}
	return carray, clen
}

func fromCDefineArray(cptr *C.ffiapi_define, clen C.size_t) map[string]string {
	length := int(clen)
	slice := (*[1 << 30]C.ffiapi_define)(unsafe.Pointer(cptr))[:length:length]
	govalue := make(map[string]string, length)
	for i := 0; i < length; i++ {
		define := slice[i]
		govalue[define.from] = define.to
	}
	return govalue
}

func fromCEngineArray(cptr *C.ffiapi_engine, clen C.size_t) []api.Engine {
	length := int(clen)
	slice := (*[1 << 30]C.ffiapi_engine)(unsafe.Pointer(cptr))[:length:length]
	govalue := make([]api.Engine, length)
	for i := 0; i < length; i++ {
		engine := slice[i]
		govalue[i] = api.Engine{
			Name:    api.EngineName(engine.name),
			Version: engine.version,
		}
	}
	return govalue
}

func fromCLoaderArray(cptr *C.ffiapi_loader, clen C.size_t) map[string]api.Loader {
	length := int(clen)
	slice := (*[1 << 30]C.ffiapi_loader)(unsafe.Pointer(cptr))[:length:length]
	govalue := make(map[string]api.Loader, length)
	for i := 0; i < length; i++ {
		loader := slice[i]
		govalue[loader.name] = api.Loader(loader.loader)
	}
	return govalue
}

func callBuildApi(
	alloc C.allocator,
	cb C.build_api_callback,
	cbData unsafe.Pointer,
	buildOptions api.BuildOptions,
) {
	goresult := api.Build(buildOptions)

	coutputFiles, coutputFilesLen := toCOutputFileArray(alloc, goresult.OutputFiles)
	cerrors, cerrorsLen := toCMessageArray(alloc, goresult.Errors)
	cwarnings, cwarningsLen := toCMessageArray(alloc, goresult.Warnings)

	C.call_build_api_callback(cb, cbData, coutputFiles, coutputFilesLen, cerrors, cerrorsLen, cwarnings, cwarningsLen)
}

//export GoBuild
func GoBuild(
	alloc C.allocator,
	cb C.build_api_callback,
	cbData unsafe.Pointer,

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

	globalName string,
	bundle C.bool,
	splitting C.bool,
	outfile string,
	metafile string,
	outdir string,
	platform C.uint8_t,
	format C.uint8_t,
	externals []string,
	loaders *C.ffiapi_loader,
	loadersLen C.size_t,
	resolveExtensions []string,
	tsconfig string,

	entryPoints []string,
) {
	go callBuildApi(alloc, cb, cbData, api.BuildOptions{
		Sourcemap: api.SourceMap(sourceMap),
		Target:    api.Target(target),
		Engines:   fromCEngineArray(engines, enginesLen),
		Strict: api.StrictOptions{
			NullishCoalescing: bool(strictNullishCoalescing),
			ClassFields:       bool(strictClassFields),
		},

		MinifyWhitespace:  bool(minifyWhitespace),
		MinifyIdentifiers: bool(minifyIdentifiers),
		MinifySyntax:      bool(minifySyntax),

		JSXFactory:  jsxFactory,
		JSXFragment: jsxFragment,

		Defines:       fromCDefineArray(defines, definesLen),
		PureFunctions: pureFunctions,

		GlobalName:        globalName,
		Bundle:            bool(bundle),
		Splitting:         bool(splitting),
		Outfile:           outfile,
		Metafile:          metafile,
		Outdir:            outdir,
		Platform:          api.Platform(platform),
		Format:            api.Format(format),
		Externals:         externals,
		Loaders:           fromCLoaderArray(loaders, loadersLen),
		ResolveExtensions: resolveExtensions,
		Tsconfig:          tsconfig,

		EntryPoints: entryPoints,
	})
}

func callTransformApi(
	alloc C.allocator,
	cb C.transform_api_callback,
	cbData unsafe.Pointer,
	code string,
	transformOptions api.TransformOptions,
) {
	goresult := api.Transform(code, transformOptions)

	cjs := toCString(alloc, goresult.JS)
	cjsSourceMap := toCString(alloc, goresult.JSSourceMap)
	cerrors, cerrorsLen := toCMessageArray(alloc, goresult.Errors)
	cwarnings, cwarningsLen := toCMessageArray(alloc, goresult.Warnings)

	C.call_transform_api_callback(cb, cbData, cjs, cjsSourceMap, cerrors, cerrorsLen, cwarnings, cwarningsLen)
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
	pureFunctions []string,

	sourceFile string,
	loader C.uint8_t,
) {
	go callTransformApi(alloc, cb, cbData, code, api.TransformOptions{
		Sourcemap: api.SourceMap(sourceMap),
		Target:    api.Target(target),
		Engines:   fromCEngineArray(engines, enginesLen),
		Strict: api.StrictOptions{
			NullishCoalescing: bool(strictNullishCoalescing),
			ClassFields:       bool(strictClassFields),
		},

		MinifyWhitespace:  bool(minifyWhitespace),
		MinifyIdentifiers: bool(minifyIdentifiers),
		MinifySyntax:      bool(minifySyntax),

		JSXFactory:  jsxFactory,
		JSXFragment: jsxFragment,

		Defines:       fromCDefineArray(defines, definesLen),
		PureFunctions: pureFunctions,

		Sourcefile: sourceFile,
		Loader:     api.Loader(loader),
	})
}

func main() {}
