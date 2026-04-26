// src/lib.rs
pub mod reader;
pub mod writer;
pub mod c_api;

#[cfg(feature = "pyo3")]
use pyo3::prelude::*;

/// CemirCol — High-performance columnar storage with zlib compression.
#[cfg(feature = "pyo3")]
#[pymodule]
fn _cemircol(m: &Bound<'_, PyModule>) -> PyResult<()> {
    m.add_class::<writer::CemircolWriter>()?;
    m.add_class::<reader::CemircolReader>()?;
    Ok(())
}