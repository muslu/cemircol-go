# Performance Optimizations

## Columnar Memory Layout
Data is stored in contiguous blocks per column, allowing for excellent CPU cache locality and efficient SIMD processing.

## Zstd Compression
Switched to Zstd (level 3) for the C-ABI writer to balance compression speed and ratio. This significantly reduces disk I/O while maintaining high throughput.

## Go/Rust Bridge (CGO)
- **Pinned Pointers**: Uses `unsafe.Pointer` to share data across the ABI boundary without intermediate allocations.
- **Explicit Freeing**: Implemented a cleanup mechanism to free Rust-allocated memory from Go, preventing leaks while avoiding Go GC overhead for large data blocks.
- **RPATH Handling**: Configured `cgo` LDFLAGS with RPATH to ensure the shared library is correctly located at runtime.

## Benchmark Results (100k - 1M rows)
- **CemirCol Read**: ~75-80M rows/s.
- **Parquet Comparison**: CemirCol demonstrated ~10x faster read speeds and ~4x smaller file sizes for simple numeric/dictionary-encoded logs.
