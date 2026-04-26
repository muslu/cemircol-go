// Package cemircol provides Go bindings for the CemirCol high-performance columnar storage engine.
// It uses CGO to interface with a Rust core library and provides mmap-based fast reading.
package cemircol

/*
#cgo LDFLAGS: -L${SRCDIR}/../target/release -lcemircol -Wl,-rpath,${SRCDIR}/../target/release
#include <stdlib.h>

typedef void* cemircol_reader_t;

extern cemircol_reader_t cemircol_reader_new(const char* filename);
extern void cemircol_reader_free(cemircol_reader_t reader);
extern unsigned long long cemircol_reader_num_rows(cemircol_reader_t reader);
extern double* cemircol_reader_query_float64(cemircol_reader_t reader, const char* column, size_t* out_len);
extern long long* cemircol_reader_query_int64(cemircol_reader_t reader, const char* column, size_t* out_len);
extern int cemircol_writer_write_float64(const char* filename, const char* column, const double* data, size_t len);
extern int cemircol_writer_write_int64(const char* filename, const char* column, const long long* data, size_t len);
extern void cemircol_free_data(void* ptr, size_t len, int is_float);
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// Reader represents a CemirCol file reader.
type Reader struct {
	ptr C.cemircol_reader_t
}

// NewReader opens a CemirCol file and returns a Reader.
// It uses memory mapping (mmap) for high-performance data access.
func NewReader(filename string) (*Reader, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	ptr := C.cemircol_reader_new(cFilename)
	if ptr == nil {
		return nil, fmt.Errorf("failed to open cemircol file: %s", filename)
	}
	return &Reader{ptr: ptr}, nil
}

// Close releases the resources associated with the reader and unmaps the file.
func (r *Reader) Close() {
	if r.ptr != nil {
		C.cemircol_reader_free(r.ptr)
		r.ptr = nil
	}
}

// NumRows returns the total number of rows stored in the CemirCol file.
func (r *Reader) NumRows() uint64 {
	if r.ptr == nil {
		return 0
	}
	return uint64(C.cemircol_reader_num_rows(r.ptr))
}

// QueryFloat64 reads a float64 column into a Go slice.
// It provides zero-copy performance by copying data directly from memory-mapped blocks.
func (r *Reader) QueryFloat64(column string) ([]float64, error) {
	if r.ptr == nil {
		return nil, fmt.Errorf("reader is closed")
	}

	cColumn := C.CString(column)
	defer C.free(unsafe.Pointer(cColumn))

	var outLen C.size_t
	ptr := C.cemircol_reader_query_float64(r.ptr, cColumn, &outLen)
	if ptr == nil {
		return nil, fmt.Errorf("column '%s' not found or error reading", column)
	}

	cSlice := unsafe.Slice((*float64)(unsafe.Pointer(ptr)), int(outLen))
	goSlice := make([]float64, int(outLen))
	copy(goSlice, cSlice)
	C.cemircol_free_data(unsafe.Pointer(ptr), outLen, 1)

	return goSlice, nil
}

// QueryInt64 reads an int64 column into a Go slice.
// It provides zero-copy performance by copying data directly from memory-mapped blocks.
func (r *Reader) QueryInt64(column string) ([]int64, error) {
	if r.ptr == nil {
		return nil, fmt.Errorf("reader is closed")
	}

	cColumn := C.CString(column)
	defer C.free(unsafe.Pointer(cColumn))

	var outLen C.size_t
	ptr := C.cemircol_reader_query_int64(r.ptr, cColumn, &outLen)
	if ptr == nil {
		return nil, fmt.Errorf("column '%s' not found or error reading", column)
	}

	cSlice := unsafe.Slice((*int64)(unsafe.Pointer(ptr)), int(outLen))
	goSlice := make([]int64, int(outLen))
	copy(goSlice, cSlice)
	C.cemircol_free_data(unsafe.Pointer(ptr), outLen, 0)

	return goSlice, nil
}

// WriteFloat64 creates a CemirCol file with a single float64 column.
// The file is compressed using Zstd for efficient storage.
func WriteFloat64(filename, column string, data []float64) error {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))
	cColumn := C.CString(column)
	defer C.free(unsafe.Pointer(cColumn))

	res := C.cemircol_writer_write_float64(cFilename, cColumn, (*C.double)(unsafe.Pointer(&data[0])), C.size_t(len(data)))
	if res != 0 {
		return fmt.Errorf("failed to write cemircol file (error code: %d)", res)
	}
	return nil
}

// WriteInt64 creates a CemirCol file with a single int64 column.
// The file is compressed using Zstd for efficient storage.
func WriteInt64(filename, column string, data []int64) error {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))
	cColumn := C.CString(column)
	defer C.free(unsafe.Pointer(cColumn))

	res := C.cemircol_writer_write_int64(cFilename, cColumn, (*C.longlong)(unsafe.Pointer(&data[0])), C.size_t(len(data)))
	if res != 0 {
		return fmt.Errorf("failed to write cemircol file (error code: %d)", res)
	}
	return nil
}
