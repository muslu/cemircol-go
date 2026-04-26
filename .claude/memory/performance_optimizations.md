---
name: Performans Optimizasyonları
description: Rust tarafında yapılan tüm optimizasyonlar, gerekçeleri ve ölçülen sonuçları
type: project
---

v0.1.3 → v0.1.4 arasında yapılan Rust optimizasyonları (10M satır benchmark).

**Why:** CemirCol okuma hızı Parquet'in gerisindeydi (0.33s vs 0.19s) ve dosya boyutu zlib ile 27MB'dı.
**How to apply:** Yeni sıkıştırma algoritması eklenirse FileMeta.compression alanını genişlet ve reader'da yeni branch ekle. PyList dönüşünden kaçın — her zaman numpy/array.array tercih et.

## Sonuçlar
| Metrik | Önce | Sonra |
|---|---|---|
| Dosya boyutu | 27.57 MB | 5.52 MB (5x küçük) |
| CemirCol okuma | 0.329 s | 0.108 s (3x hızlı) |
| Parquet okuma | 0.193 s | 0.113 s |

CemirCol artık Parquet'ten %5 daha hızlı okuma yapıyor, 15x daha küçük dosya üretiyor.

## Optimizasyon 1: zlib → zstd level 22 (maksimum sıkıştırma)
- **Ne:** `flate2::ZlibEncoder` → `zstd::encode_all(..., 22)`
- **Neden:** zstd level 22 zlib level 9'dan çok daha iyi sıkıştırma sağlar. Decompression hızı ise level'dan bağımsız olarak yüksek kalır.
- **Geriye uyumluluk:** `FileMeta.compression` alanı eklendi (`#[serde(default = "default_zlib")]`). Eski dosyalar `"zlib"` ile okunur.

## Optimizasyon 2: Rayon paralel sıkıştırma (writer)
- **Ne:** Sütun verisi önce GIL tutarak Python'dan alınır, sonra `raw_columns.into_par_iter()` ile tüm sütunlar paralel sıkıştırılır.
- **Neden:** Çok sütunlu yazmalarda tüm CPU core'larını kullanır. Sıkıştırma saf Rust — GIL gerekmez.

## Optimizasyon 3: Unsafe zero-copy byte cast (writer)
- **Ne:** `values.iter().flat_map(|v| v.to_le_bytes()).collect()` → `std::slice::from_raw_parts(values.as_ptr() as *const u8, n * 8).to_vec()`
- **Neden:** Tek memcopy, per-element iterasyon yükü yok. x86 little-endian varsayımı (güvenli).

## Optimizasyon 4: BufWriter 8MB buffer (writer)
- **Ne:** `File::create` → `BufWriter::with_capacity(8 * 1024 * 1024, file)`
- **Neden:** Küçük yazmaları biriktirip toplu flush ile syscall sayısını azaltır.

## Optimizasyon 5: PyByteArray sıfır-kopya decompress (reader) ← EN BÜYÜK KAZANIM
- **Ne:** `Vec<u8>` ara buffer yok. `PyByteArray::new_with(py, len, |buf| { zstd::stream::copy_decode(compressed, Cursor::new(buf)) })` ile doğrudan Python belleğine decompress.
- **Neden:** 80MB'lık `Vec<u8>` allocation + copy ortadan kalktı. Peak memory yarıya indi.

## Optimizasyon 6: numpy.frombuffer sıfır-kopya view (reader) ← EN BÜYÜK KAZANIM
- **Ne:** PyList (10M Python nesnesi) yerine `numpy.frombuffer(py_bytearray, dtype)` → sıfır-kopya numpy view.
- **Neden:** 10M Python float nesnesi yaratmak ~0.15s alıyordu. numpy view anlık.
- **Fallback sırası:** numpy → array.array → PyList (son çare)

## Optimizasyon 7: Unsafe slice cast (reader fallback)
- **Ne:** `chunks_exact(8).map(...).collect()` → `std::slice::from_raw_parts(ptr as *const f64, len/8)`
- **Neden:** PyList fallback'te bile ara Vec<f64> allocation yok.
