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

//*** Create ffiapi_message array functions (interal Go use). ***

static inline ffiapi_message* create_ffiapi_message_array(allocator alloc, size_t len) {
	return alloc(sizeof(ffiapi_message) * len);
}

static inline void set_ffiapi_message_array_element(
	ffiapi_message* array,
	size_t i,
	ffiapi_string file,
	ptrdiff_t line,
	ptrdiff_t column,
	ptrdiff_t length,
	ffiapi_string text
) {
	ffiapi_message msg = {
		.file = file,
		.line = line,
		.column = column,
		.length = length,
		.text = text,
	};
	array[i] = msg;
}

//*** Create ffiapi_output_file array functions (interal Go use). ***

static inline ffiapi_output_file* create_ffiapi_output_file_array(allocator alloc, size_t len) {
	return alloc(sizeof(ffiapi_output_file) * len);
}

static inline void set_ffiapi_output_file_array_element(ffiapi_output_file* array, size_t i, ffiapi_string path, ffiapi_string data) {
	ffiapi_output_file file = {
		.path = path,
		.data = data,
	};
	array[i] = file;
}

//*** Access input C array element functions (interal Go use). ***

static inline ffiapi_define get_ffiapi_define_array_element(ffiapi_define* array, size_t i) {
	return array[i];
}

static inline ffiapi_engine get_ffiapi_engine_array_element(ffiapi_engine* array, size_t i) {
	return array[i];
}

static inline ffiapi_loader get_ffiapi_loader_array_element(ffiapi_loader* array, size_t i) {
	return array[i];
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

func copyToCString(alloc C.allocator, bytes []byte) C.ffiapi_string {
	var godataPtr unsafe.Pointer = nil
	if len(bytes) > 0 {
		godataPtr = unsafe.Pointer(&bytes[0])
	}
	return C.create_ffiapi_string_from_bytes(alloc, C.size_t(len(bytes)), godataPtr)
}

func copyToCMessageArray(alloc C.allocator, messages []api.Message) (*C.ffiapi_message, C.size_t) {
	clen := C.size_t(len(messages))
	carray := C.create_ffiapi_message_array(alloc, clen)
	for i, msg := range messages {
		loc := msg.Location
		if loc == nil {
			loc = &api.Location{}
		}
		C.set_ffiapi_message_array_element(
			carray,
			C.size_t(i),
			C.create_ffiapi_string(alloc, loc.File),
			C.ptrdiff_t(loc.Line),
			C.ptrdiff_t(loc.Column),
			C.ptrdiff_t(loc.Length),
			C.create_ffiapi_string(alloc, msg.Text),
		)
	}
	return carray, clen
}

func copyToCOutputFileArray(alloc C.allocator, outputFiles []api.OutputFile) (*C.ffiapi_output_file, C.size_t) {
	clen := C.size_t(len(outputFiles))
	carray := C.create_ffiapi_output_file_array(alloc, clen)
	for i, file := range outputFiles {
		C.set_ffiapi_output_file_array_element(
			carray,
			C.size_t(i),
			C.create_ffiapi_string(alloc, file.Path),
			copyToCString(alloc, file.Contents),
		)
	}
	return carray, clen
}

func convertCDefineArrayToMap(cptr *C.ffiapi_define, clen C.size_t) map[string]string {
	godefinesLen := int(clen)
	godefines := make(map[string]string, godefinesLen)
	for i := 0; i < godefinesLen; i++ {
		define := C.get_ffiapi_define_array_element(cptr, C.size_t(i))
		godefines[define.from] = define.to
	}
	return godefines
}

func convertCEngineArrayToSlice(cptr *C.ffiapi_engine, clen C.size_t) []api.Engine {
	goenginesLen := int(clen)
	goengines := make([]api.Engine, goenginesLen)
	for i := 0; i < goenginesLen; i++ {
		engine := C.get_ffiapi_engine_array_element(cptr, C.size_t(i))
		goengines[i] = api.Engine{
			Name:    api.EngineName(engine.name),
			Version: engine.version,
		}
	}
	return goengines
}

func convertCLoaderArrayToMap(cptr *C.ffiapi_loader, clen C.size_t) map[string]api.Loader {
	goloadersLen := int(clen)
	goloaders := make(map[string]api.Loader, goloadersLen)
	for i := 0; i < goloadersLen; i++ {
		loader := C.get_ffiapi_loader_array_element(cptr, C.size_t(i))
		goloaders[loader.name] = api.Loader(loader.loader)
	}
	return goloaders
}

func callBuildApi(
	alloc C.allocator,
	cb C.build_api_callback,
	cbData unsafe.Pointer,
	buildOptions api.BuildOptions,
) {
	goresult := api.Build(buildOptions)

	coutputFiles, coutputFilesLen := copyToCOutputFileArray(alloc, goresult.OutputFiles)
	cerrors, cerrorsLen := copyToCMessageArray(alloc, goresult.Errors)
	cwarnings, cwarningsLen := copyToCMessageArray(alloc, goresult.Warnings)

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
		Engines:   convertCEngineArrayToSlice(engines, enginesLen),
		Strict: api.StrictOptions{
			NullishCoalescing: bool(strictNullishCoalescing),
			ClassFields:       bool(strictClassFields),
		},

		MinifyWhitespace:  bool(minifyWhitespace),
		MinifyIdentifiers: bool(minifyIdentifiers),
		MinifySyntax:      bool(minifySyntax),

		JSXFactory:  jsxFactory,
		JSXFragment: jsxFragment,

		Defines:       convertCDefineArrayToMap(defines, definesLen),
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
		Loaders:           convertCLoaderArrayToMap(loaders, loadersLen),
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

	cjs := copyToCString(alloc, goresult.JS)
	cjsSourceMap := copyToCString(alloc, goresult.JSSourceMap)
	cerrors, cerrorsLen := copyToCMessageArray(alloc, goresult.Errors)
	cwarnings, cwarningsLen := copyToCMessageArray(alloc, goresult.Warnings)

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
		Engines:   convertCEngineArrayToSlice(engines, enginesLen),
		Strict: api.StrictOptions{
			NullishCoalescing: bool(strictNullishCoalescing),
			ClassFields:       bool(strictClassFields),
		},

		MinifyWhitespace:  bool(minifyWhitespace),
		MinifyIdentifiers: bool(minifyIdentifiers),
		MinifySyntax:      bool(minifySyntax),

		JSXFactory:  jsxFactory,
		JSXFragment: jsxFragment,

		Defines:       convertCDefineArrayToMap(defines, definesLen),
		PureFunctions: pureFunctions,

		Sourcefile: sourceFile,
		Loader:     api.Loader(loader),
	})
}

func main() {}
