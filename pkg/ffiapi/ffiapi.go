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

// WARNING: This must match the GoSlice declaration, including the GoInt type.
// We can't use GoSlice or GoInt because they are not available in the preamble.
typedef struct ffiapi_gostring_goslice {
	_GoString_* data;
	ptrdiff_t len;
	ptrdiff_t cap;
} ffiapi_gostring_goslice;

typedef struct ffiapi_define {
	_GoString_ name;
	_GoString_ value;
} ffiapi_define;

typedef struct ffiapi_engine {
	uint8_t name;
	_GoString_ version;
} ffiapi_engine;

typedef struct ffiapi_loader {
	_GoString_ name;
	uint8_t loader;
} ffiapi_loader;

//*** Option structs. ***

typedef struct ffiapi_build_options {
	uint8_t source_map;
	uint8_t target;
	ffiapi_engine* engines;
	size_t engines_len;
	bool strict_nullish_coalescing;
	bool strict_class_fields;

	bool minify_whitespace;
	bool minify_identifiers;
	bool minify_syntax;

	_GoString_ jsx_factory;
	_GoString_ jsx_fragment;

	ffiapi_define* defines;
	size_t defines_len;
	ffiapi_gostring_goslice pure_functions;

	_GoString_ global_name;
	bool bundle;
	bool splitting;
	_GoString_ outfile;
	_GoString_ metafile;
	_GoString_ outdir;
	uint8_t platform;
	uint8_t format;
	ffiapi_gostring_goslice externals;
	ffiapi_loader* loaders;
	size_t loaders_len;
	ffiapi_gostring_goslice resolve_extensions;
	_GoString_ tsconfig;

	ffiapi_gostring_goslice entry_points;
} ffiapi_build_options;

typedef struct ffiapi_transform_options {
	uint8_t source_map;
	uint8_t sources_content;

	uint8_t target;
	uint8_t format;
	_GoString_ global_name;
	ffiapi_engine* engines;
	size_t engines_len;

	bool minify_whitespace;
	bool minify_identifiers;
	bool minify_syntax;
	uint8_t charset;
	uint8_t tree_shaking;

	_GoString_ jsx_factory;
	_GoString_ jsx_fragment;
	_GoString_ tsconfig_raw;
	_GoString_ footer;
	_GoString_ banner;

	ffiapi_define* defines;
	size_t defines_len;
	ffiapi_gostring_goslice pure;
	bool avoid_tdz;
	bool keep_names;

	_GoString_ source_file;
	uint8_t loader;
} ffiapi_transform_options;

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
	ffiapi_string code,
	ffiapi_string map,
	ffiapi_message* errors,
	size_t errors_len,
	ffiapi_message* warnings,
	size_t warnings_len
);

//*** Create ffiapi_string functions (interal Go use). ***

static inline ffiapi_string create_ffiapi_string_from_bytes(allocator alloc, size_t len, void const* godata) {
	char* data = (char*) alloc(len);
	if (data != NULL && godata != NULL) {
		memcpy(data, godata, len);
	}
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

static inline void* call_alloc(allocator alloc, size_t bytes) {
	return alloc(bytes);
}

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
	ffiapi_string code,
	ffiapi_string map,
	ffiapi_message* errors,
	size_t errors_len,
	ffiapi_message* warnings,
	size_t warnings_len
) {
	f(cb_data, code, map, errors, errors_len, warnings, warnings_len);
}
*/
import "C"
import (
	"github.com/evanw/esbuild/pkg/api"
	"unsafe"
)

func asStringSlice(cptr *C.ffiapi_gostring_goslice) []string {
	return *(*[]string)(unsafe.Pointer(cptr))
}

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
		govalue[define.name] = define.value
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
	opt *C.ffiapi_build_options,
) {
	go callBuildApi(alloc, cb, cbData, api.BuildOptions{
		Sourcemap: api.SourceMap(opt.source_map),
		Target:    api.Target(opt.target),
		Engines:   fromCEngineArray(opt.engines, opt.engines_len),
		Strict: api.StrictOptions{
			NullishCoalescing: bool(opt.strict_nullish_coalescing),
			ClassFields:       bool(opt.strict_class_fields),
		},

		MinifyWhitespace:  bool(opt.minify_whitespace),
		MinifyIdentifiers: bool(opt.minify_identifiers),
		MinifySyntax:      bool(opt.minify_syntax),

		JSXFactory:  opt.jsx_factory,
		JSXFragment: opt.jsx_fragment,

		Defines:       fromCDefineArray(opt.defines, opt.defines_len),
		PureFunctions: asStringSlice(&opt.pure_functions),

		GlobalName:        opt.global_name,
		Bundle:            bool(opt.bundle),
		Splitting:         bool(opt.splitting),
		Outfile:           opt.outfile,
		Metafile:          opt.metafile,
		Outdir:            opt.outdir,
		Platform:          api.Platform(opt.platform),
		Format:            api.Format(opt.format),
		Externals:         asStringSlice(&opt.externals),
		Loaders:           fromCLoaderArray(opt.loaders, opt.loaders_len),
		ResolveExtensions: asStringSlice(&opt.resolve_extensions),
		Tsconfig:          opt.tsconfig,

		EntryPoints: asStringSlice(&opt.entry_points),
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

	cjs := toCString(alloc, goresult.Code)
	cjsSourceMap := toCString(alloc, goresult.Map)
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
	opt *C.ffiapi_transform_options,
) {
	go callTransformApi(alloc, cb, cbData, code, api.TransformOptions{
		Sourcemap:      api.SourceMap(opt.source_map),
		SourcesContent: api.SourcesContent(opt.sources_content),

		Target:     api.Target(opt.target),
		Format:     api.Format(opt.format),
		GlobalName: opt.global_name,
		Engines:    fromCEngineArray(opt.engines, opt.engines_len),

		MinifyWhitespace:  bool(opt.minify_whitespace),
		MinifyIdentifiers: bool(opt.minify_identifiers),
		MinifySyntax:      bool(opt.minify_syntax),
		Charset:           api.Charset(opt.charset),
		TreeShaking:       api.TreeShaking(opt.tree_shaking),

		JSXFactory:  opt.jsx_factory,
		JSXFragment: opt.jsx_fragment,
		TsconfigRaw: opt.tsconfig_raw,
		Footer:      opt.footer,
		Banner:      opt.banner,

		Define:    fromCDefineArray(opt.defines, opt.defines_len),
		Pure:      asStringSlice(&opt.pure_functions),
		AvoidTDZ:  bool(opt.avoid_tdz),
		KeepNames: bool(opt.keep_names),

		Sourcefile: opt.source_file,
		Loader:     api.Loader(opt.loader),
	})
}

func main() {}
