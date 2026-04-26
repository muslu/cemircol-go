#[cfg(feature = "pyo3")]
use pyo3::prelude::*;
#[cfg(feature = "pyo3")]
use pyo3::types::PyDict;
#[cfg(feature = "pyo3")]
use rayon::prelude::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
#[cfg(feature = "pyo3")]
use std::fs::File;
#[cfg(feature = "pyo3")]
use std::io::{BufWriter, Write};

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct ColumnMeta {
    pub offset: u64,
    pub compressed_length: u64,
    pub uncompressed_length: u64,
    pub data_type: String,
}

fn default_compression() -> String {
    "zlib".to_string()
}

#[derive(Serialize, Deserialize, Debug)]
pub struct FileMeta {
    pub num_rows: u64,
    pub columns: HashMap<String, ColumnMeta>,
    #[serde(default = "default_compression")]
    pub compression: String,
}

#[cfg_attr(feature = "pyo3", pyclass)]
pub struct CemircolWriter;

#[cfg_attr(feature = "pyo3", pymethods)]
impl CemircolWriter {
    #[cfg(feature = "pyo3")]
    #[staticmethod]
    fn write(filename: &str, data: &Bound<'_, PyDict>) -> PyResult<()> {
        if data.is_empty() {
            return Ok(());
        }

        let mut num_rows: u64 = 0;

        // GIL tutarak tüm sütun verilerini çek
        let mut raw_columns: Vec<(String, Vec<u8>, &'static str)> =
            Vec::with_capacity(data.len());

        for (key, value) in data.iter() {
            let col_name: String = key.extract()?;

            let (raw_bytes, dtype) = if let Ok(values) = value.extract::<Vec<i64>>() {
                let n = values.len();
                if num_rows == 0 {
                    num_rows = n as u64;
                } else if n as u64 != num_rows {
                    return Err(pyo3::exceptions::PyValueError::new_err(format!(
                        "Column '{}' length mismatch",
                        col_name
                    )));
                }
                // Sıfır kopya: Vec<i64> → &[u8] (x86 little-endian)
                let bytes = unsafe {
                    std::slice::from_raw_parts(values.as_ptr() as *const u8, n * 8).to_vec()
                };
                (bytes, "int64")
            } else if let Ok(values) = value.extract::<Vec<f64>>() {
                let n = values.len();
                if num_rows == 0 {
                    num_rows = n as u64;
                } else if n as u64 != num_rows {
                    return Err(pyo3::exceptions::PyValueError::new_err(format!(
                        "Column '{}' length mismatch",
                        col_name
                    )));
                }
                let bytes = unsafe {
                    std::slice::from_raw_parts(values.as_ptr() as *const u8, n * 8).to_vec()
                };
                (bytes, "float64")
            } else {
                return Err(pyo3::exceptions::PyTypeError::new_err(format!(
                    "Column '{}': unsupported type (expected list of int or float)",
                    col_name
                )));
            };

            raw_columns.push((col_name, raw_bytes, dtype));
        }

        // Tüm sütunları rayon ile paralel sıkıştır (zstd level 22 = max sıkıştırma)
        // Rayon thread'leri Python nesnesine dokunmaz, GIL gerekmez
        let compressed_columns: Vec<(String, Vec<u8>, &str, u64)> = raw_columns
            .into_par_iter()
            .map(|(name, raw_bytes, dtype)| {
                let uncompressed_len = raw_bytes.len() as u64;
                let compressed =
                    zstd::encode_all(&raw_bytes[..], 22).expect("zstd compression failed");
                (name, compressed, dtype, uncompressed_len)
            })
            .collect();

        // Dosyaya yaz
        let file = File::create(filename)
            .map_err(|e| pyo3::exceptions::PyIOError::new_err(e.to_string()))?;
        let mut writer = BufWriter::with_capacity(8 * 1024 * 1024, file);

        writer
            .write_all(b"CEM1")
            .map_err(|e| pyo3::exceptions::PyIOError::new_err(e.to_string()))?;

        // Offset hesapla (magic 4 byte'tan sonra)
        let mut offset: u64 = 4;
        let mut meta = FileMeta {
            num_rows,
            columns: HashMap::with_capacity(compressed_columns.len()),
            compression: "zstd".to_string(),
        };

        for (name, compressed, dtype, uncompressed_len) in &compressed_columns {
            meta.columns.insert(
                name.clone(),
                ColumnMeta {
                    offset,
                    compressed_length: compressed.len() as u64,
                    uncompressed_length: *uncompressed_len,
                    data_type: dtype.to_string(),
                },
            );
            offset += compressed.len() as u64;
        }

        for (_, compressed, _, _) in &compressed_columns {
            writer
                .write_all(compressed)
                .map_err(|e| pyo3::exceptions::PyIOError::new_err(e.to_string()))?;
        }

        let meta_json = serde_json::to_vec(&meta)
            .map_err(|e| pyo3::exceptions::PyValueError::new_err(e.to_string()))?;
        writer
            .write_all(&meta_json)
            .map_err(|e| pyo3::exceptions::PyIOError::new_err(e.to_string()))?;
        writer
            .write_all(&(meta_json.len() as u64).to_le_bytes())
            .map_err(|e| pyo3::exceptions::PyIOError::new_err(e.to_string()))?;
        writer
            .write_all(b"CEM1")
            .map_err(|e| pyo3::exceptions::PyIOError::new_err(e.to_string()))?;

        Ok(())
    }
}
