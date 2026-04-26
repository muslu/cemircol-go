# CemirCol-Go AI Guide

## Build & Setup
- **Initial Setup**: Run `./setup.sh` to install Rust/Go dependencies and build the core library.
- **Rust Core**: `cargo build --release` (generates `target/release/libcemircol.a` and `.so`).
- **Go Build**: `go build ./...`
- **Full Clean Build**: `cargo clean && ./setup.sh`

## Project Structure
- `cemircol/`: Go wrapper for the Rust core using CGO.
- `src/`: Rust implementation of the columnar storage engine.
  - `reader.rs`: mmap-based fast reading and query logic.
  - `writer.rs`: Zstd compressed columnar writing logic.
  - `c_api.rs`: C-ABI exposure for Go.
- `benchmark/`: Performance testing and Parquet comparison scripts.

## Key Commands
- **Benchmark**: `go run benchmark/gen/main.go && go run benchmark/parser/main.go && go run benchmark/logger/main.go`
- **Full Comparison**: `go run benchmark/compare_all/main.go`
- **Parquet Comparison**: `go run benchmark/pq_parser/main.go && go run benchmark/pq_logger/main.go`
- **Publish**: `./publish.sh`

## Code Style
- **Rust**: Follow standard `cargo fmt` style. Use `#[cfg(feature = "pyo3")]` for Python-specific code.
- **Go**: Follow `gofmt` style. Ensure `cgo` flags point to the correct library paths using `${SRCDIR}`.
- **C-ABI**: Ensure pointers are correctly freed using `cemircol_free_data` to avoid memory leaks.
