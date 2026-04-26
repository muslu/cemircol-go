---
name: Proje Genel Bakış
description: cemircol'un ne olduğu, mimarisi, dosya formatı ve bağımlılıkları
type: project
---

cemircol, Rust/PyO3 ile yazılmış yüksek performanslı sütun tabanlı veri depolama kütüphanesidir.

**Why:** Parquet'ten daha küçük dosya boyutu ve daha hızlı sütun sorgusu sağlamak için tasarlandı.
**How to apply:** Yeni özellik eklerken mevcut `.cemir` dosya formatını bozmamaya dikkat et; geriye dönük uyumluluğu `compression` alanındaki `#[serde(default)]` gibi mekanizmalarla koru.

## Dosya Formatı (.cemir)
```
[magic: b"CEM1"] [col_data_1] [col_data_2] ... [metadata_json] [meta_len: u64 LE] [magic: b"CEM1"]
```
- Her sütun bağımsız sıkıştırılır (zstd level 22 = maksimum)
- `FileMeta` JSON footer'da: `num_rows`, `columns: HashMap<name, ColumnMeta>`, `compression`
- `ColumnMeta`: `offset`, `compressed_length`, `uncompressed_length`, `data_type`
- Desteklenen tipler: `int64`, `float64`

## Mimari
- `src/lib.rs` — PyModule giriş noktası
- `src/writer.rs` — `CemircolWriter::write()` static method; FileMeta + ColumnMeta tanımları burada
- `src/reader.rs` — `CemircolReader`; mmap + metadata parse + query()
- `cemircol/__init__.py` — Python katmanı; `from_csv`, `from_parquet` helper'ları

## Bağımlılıklar (Cargo.toml)
- `pyo3 = "0.28"` — Python binding
- `memmap2 = "0.9"` — memory-mapped file okuma
- `zstd = "0.13"` — hızlı sıkıştırma (yeni format)
- `flate2 = "1.0"` — eski zlib formatı için geriye dönük uyumluluk
- `rayon = "1.10"` — paralel sütun sıkıştırma
- `serde + serde_json` — metadata serializasyonu

## Release profili
`lto = "fat"`, `codegen-units = 1`, `opt-level = 3`, `panic = "abort"` — maksimum binary optimizasyonu
