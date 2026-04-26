# Project Overview: CemirCol-Go

CemirCol-Go is a high-performance columnar storage library for Go, powered by a Rust core. It is designed to be a faster and more efficient alternative to Parquet for specific numeric and correlated data workloads.

## Key Features
- **mmap-based Reading**: Instant file loading without reading the entire file into memory.
- **Zero-Copy Architecture**: Data is read directly from memory-mapped files into Go slices with minimal overhead.
- **High Compression**: Utilizes Zstd compression for efficient storage (~4x smaller than equivalent Parquet files in some tests).
- **C-ABI Bridge**: Seamless integration between Go and Rust with safe memory management.
- **Dictionary Encoding Support**: Efficiently store repetitive strings (like email addresses) in a numeric-first columnar format.

## Current Status
- **Core Engine**: Stable reading/writing for Float64 and Int64.
- **Go Integration**: Fully functional Go package with automated setup scripts.
- **Benchmarking**: Proven performance advantages over Parquet in Go for specific logging use cases.
