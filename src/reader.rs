use flate2::read::ZlibDecoder;
use memmap2::{Mmap, MmapOptions};
#[cfg(feature = "pyo3")]
use pyo3::prelude::*;
#[cfg(feature = "pyo3")]
use pyo3::types::{PyByteArray, PyList};
use std::fs::File;
use std::io::{Cursor, Read};

use crate::writer::FileMeta;

#[cfg_attr(feature = "pyo3", pyclass)]
pub struct CemircolReader {
    mmap: Mmap,
    metadata: FileMeta,
}

#[cfg_attr(feature = "pyo3", pymethods)]
impl CemircolReader {
    pub fn open(file_path: &str) -> Result<Self, String> {
        let file = File::open(file_path).map_err(|e| e.to_string())?;

        let mmap = unsafe {
            MmapOptions::new()
                .map(&file)
                .map_err(|e| e.to_string())?
        };

        let len = mmap.len();
        if len < 16 {
            return Err("Invalid file: too small".to_string());
        }

        if &mmap[len - 4..] != b"CEM1" {
            return Err("Invalid file format: missing CEM1 magic".to_string());
        }

        let meta_len_bytes: [u8; 8] = mmap[len - 12..len - 4].try_into().map_err(|_| {
            "Failed to read metadata length".to_string()
        })?;
        let meta_len = u64::from_le_bytes(meta_len_bytes) as usize;

        let meta_start = len - 12 - meta_len;
        let meta_bytes = &mmap[meta_start..meta_start + meta_len];
        let metadata: FileMeta = serde_json::from_slice(meta_bytes).map_err(|e| {
            format!("Invalid metadata: {}", e)
        })?;

        Ok(Self { mmap, metadata })
    }

    #[cfg(feature = "pyo3")]
    #[new]
    fn new(file_path: &str) -> PyResult<Self> {
        Self::open(file_path).map_err(|e| pyo3::exceptions::PyIOError::new_err(e))
    }

    pub fn read_column_raw<T: Clone>(&self, column: &str) -> Result<Vec<T>, String> {
        let col_meta = self.metadata.columns.get(column).ok_or_else(|| {
            format!("Column '{}' not found", column)
        })?;

        let start = col_meta.offset as usize;
        let end = start + col_meta.compressed_length as usize;
        let compressed: &[u8] = &self.mmap[start..end];
        let uncompressed_length = col_meta.uncompressed_length as usize;
        let compression = self.metadata.compression.as_str();

        let mut buf = vec![0u8; uncompressed_length];
        match compression {
            "zstd" => {
                zstd::stream::copy_decode(compressed, Cursor::new(&mut buf))
                    .map_err(|e| e.to_string())?;
            }
            _ => {
                let mut decoder = ZlibDecoder::new(compressed);
                decoder.read_exact(&mut buf).map_err(|e| e.to_string())?;
            }
        }

        let n = uncompressed_length / 8;
        let data = unsafe {
            std::slice::from_raw_parts(buf.as_ptr() as *const T, n).to_vec()
        };
        Ok(data)
    }

    /// Tek sütun sorgula.
    /// Sıfır-kopya pipeline: mmap → doğrudan PyByteArray içine decompress →
    /// numpy sıfır-kopya view. Hiçbir ara Rust buffer yok.
    #[cfg(feature = "pyo3")]
    fn query<'py>(&self, py: Python<'py>, column: &str) -> PyResult<Bound<'py, PyAny>> {
        let col_meta = self.metadata.columns.get(column).ok_or_else(|| {
            pyo3::exceptions::PyKeyError::new_err(format!("Column '{}' not found", column))
        })?;

        let start = col_meta.offset as usize;
        let end = start + col_meta.compressed_length as usize;
        let compressed: &[u8] = &self.mmap[start..end];
        let uncompressed_length = col_meta.uncompressed_length as usize;
        let compression = self.metadata.compression.as_str();

        // PyByteArray'e doğrudan decompress et — ara Vec<u8> yok
        let py_bytearray = PyByteArray::new_with(py, uncompressed_length, |buf| {
            match compression {
                "zstd" => {
                    // zstd → Cursor<&mut [u8]> üzerine stream decode
                    zstd::stream::copy_decode(compressed, Cursor::new(buf))
                        .map(|_| ())
                        .map_err(|e| pyo3::exceptions::PyIOError::new_err(e.to_string()))
                }
                _ => {
                    // Eski zlib formatı — read_exact ile mevcut buffer'a yaz
                    let mut decoder = ZlibDecoder::new(compressed);
                    decoder
                        .read_exact(buf)
                        .map_err(|e| pyo3::exceptions::PyIOError::new_err(e.to_string()))
                }
            }
        })?;

        let np_dtype = match col_meta.data_type.as_str() {
            "int64" => "int64",
            "float64" => "float64",
            other => {
                return Err(pyo3::exceptions::PyTypeError::new_err(format!(
                    "Unsupported data type: {other}"
                )))
            }
        };

        // numpy.frombuffer: PyByteArray üzerinde sıfır-kopya view
        if let Ok(np) = py.import("numpy") {
            if let Ok(arr) = np.call_method1("frombuffer", (&py_bytearray, np_dtype)) {
                return Ok(arr);
            }
        }

        // Fallback: array.array (tek internal memcopy, PyList'ten çok hızlı)
        let array_code = if np_dtype == "int64" { "q" } else { "d" };
        if let Ok(array_mod) = py.import("array") {
            if let Ok(arr_cls) = array_mod.getattr("array") {
                if let Ok(arr) = arr_cls.call1((array_code,)) {
                    let _ = arr.call_method1("frombytes", (&py_bytearray,));
                    return Ok(arr);
                }
            }
        }

        // Son çare: PyList
        let raw = unsafe { py_bytearray.as_bytes() };
        match col_meta.data_type.as_str() {
            "int64" => {
                let values = unsafe {
                    std::slice::from_raw_parts(raw.as_ptr() as *const i64, raw.len() / 8)
                };
                Ok(PyList::new(py, values)?.into_any())
            }
            _ => {
                let values = unsafe {
                    std::slice::from_raw_parts(raw.as_ptr() as *const f64, raw.len() / 8)
                };
                Ok(PyList::new(py, values)?.into_any())
            }
        }
    }

    pub fn columns(&self) -> Vec<String> {
        self.metadata.columns.keys().cloned().collect()
    }

    pub fn num_rows(&self) -> u64 {
        self.metadata.num_rows
    }
}
