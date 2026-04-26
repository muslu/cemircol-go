use crate::reader::CemircolReader;
#[cfg(feature = "pyo3")]
use crate::writer::CemircolWriter;
use libc::{c_char, c_void};
use std::ffi::CStr;
use std::ptr;

#[no_mangle]
pub extern "C" fn cemircol_reader_new(filename: *const c_char) -> *mut CemircolReader {
    if filename.is_null() {
        return ptr::null_mut();
    }
    let c_str = unsafe { CStr::from_ptr(filename) };
    let filename_str = match c_str.to_str() {
        Ok(s) => s,
        Err(_) => return ptr::null_mut(),
    };

    // CemircolReader::new returns a PyResult, we need a pure Rust version or handle it
    // For simplicity, let's assume we add a 'new_rust' method or refactor
    // For now, I'll implement the logic here or call a non-pyo3 version
    match CemircolReader::open(filename_str) {
        Ok(reader) => Box::into_raw(Box::new(reader)),
        Err(_) => ptr::null_mut(),
    }
}

#[no_mangle]
pub extern "C" fn cemircol_reader_free(reader: *mut CemircolReader) {
    if !reader.is_null() {
        unsafe {
            let _ = Box::from_raw(reader);
        }
    }
}

#[no_mangle]
pub extern "C" fn cemircol_reader_num_rows(reader: *const CemircolReader) -> u64 {
    if reader.is_null() {
        return 0;
    }
    let reader = unsafe { &*reader };
    reader.num_rows()
}

#[no_mangle]
pub extern "C" fn cemircol_reader_query_float64(
    reader: *const CemircolReader,
    column: *const c_char,
    out_len: *mut usize,
) -> *mut f64 {
    if reader.is_null() || column.is_null() {
        return ptr::null_mut();
    }
    let reader = unsafe { &*reader };
    let c_str = unsafe { CStr::from_ptr(column) };
    let col_name = match c_str.to_str() {
        Ok(s) => s,
        Err(_) => return ptr::null_mut(),
    };

    match reader.read_column_raw(col_name) {
        Ok(data) => {
            let mut data = data.into_boxed_slice();
            unsafe {
                *out_len = data.len();
                let ptr = data.as_mut_ptr();
                std::mem::forget(data);
                ptr as *mut f64
            }
        }
        Err(_) => ptr::null_mut(),
    }
}

#[no_mangle]
pub extern "C" fn cemircol_writer_write_float64(
    filename: *const c_char,
    col_name: *const c_char,
    data: *const f64,
    len: usize,
) -> i32 {
    if filename.is_null() || col_name.is_null() || data.is_null() {
        return -1;
    }
    let filename_str = unsafe { CStr::from_ptr(filename).to_str().unwrap_or("") };
    let col_name_str = unsafe { CStr::from_ptr(col_name).to_str().unwrap_or("") };
    let data_slice = unsafe { std::slice::from_raw_parts(data as *const u8, len * 8) };

    let mut columns = Vec::new();
    columns.push((col_name_str.to_string(), data_slice.to_vec(), "float64".to_string()));

    match crate::writer::CemircolWriter::write_file(filename_str, columns, len as u64) {
        Ok(_) => 0,
        Err(_) => -2,
    }
}

#[no_mangle]
pub extern "C" fn cemircol_free_data(ptr: *mut c_void, len: usize, is_float: bool) {
    if ptr.is_null() {
        return;
    }
    unsafe {
        if is_float {
            let _ = Box::from_raw(std::slice::from_raw_parts_mut(ptr as *mut f64, len));
        } else {
            let _ = Box::from_raw(std::slice::from_raw_parts_mut(ptr as *mut i64, len));
        }
    }
}
