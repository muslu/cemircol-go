package cemircol

/*
#cgo LDFLAGS: -L${SRCDIR}/../target/release -lcemircol
#include <stdlib.h>

typedef void* cemircol_reader_t;

extern cemircol_reader_t cemircol_reader_new(const char* filename);
extern void cemircol_reader_free(cemircol_reader_t reader);
extern unsigned long long cemircol_reader_num_rows(cemircol_reader_t reader);
extern double* cemircol_reader_query_float64(cemircol_reader_t reader, const char* column, size_t* out_len);
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
func NewReader(filename string) (*Reader, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	ptr := C.cemircol_reader_new(cFilename)
	if ptr == nil {
		return nil, fmt.Errorf("failed to open cemircol file: %s", filename)
	}
	return &Reader{ptr: ptr}, nil
}

// Close releases the resources associated with the reader.
func (r *Reader) Close() {
	if r.ptr != nil {
		C.cemircol_reader_free(r.ptr)
		r.ptr = nil
	}
}

// NumRows returns the number of rows in the file.
func (r *Reader) NumRows() uint64 {
	if r.ptr == nil {
		return 0
	}
	return uint64(C.cemircol_reader_num_rows(r.ptr))
}

// QueryFloat64 reads a float64 column into a Go slice.
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

	// Create a Go slice pointing to the C memory
	cSlice := unsafe.Slice((*float64)(unsafe.Pointer(ptr)), int(outLen))
	
	// Copy the data to Go-managed memory to ensure safety and allow freeing the C memory
	goSlice := make([]float64, int(outLen))
	copy(goSlice, cSlice)
	
	// Free the buffer allocated by Rust
	C.cemircol_free_data(unsafe.Pointer(ptr), outLen, 1)
	
	return goSlice, nil
}
