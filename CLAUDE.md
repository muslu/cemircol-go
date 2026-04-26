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
- **Benchmark**: `cd benchmark && go run generate_logs.go && go run postfix_parser.go && go run postfix_logger.go`
- **Parquet Comparison**: `cd benchmark && go run parquet_parser.go && go run parquet_logger.go`
- **Publish**: `./publish.sh`

## Code Style
- **Rust**: Follow standard `cargo fmt` style. Use `#[cfg(feature = "pyo3")]` for Python-specific code.
- **Go**: Follow `gofmt` style. Ensure `cgo` flags point to the correct library paths using `${SRCDIR}`.
- **C-ABI**: Ensure pointers are correctly freed using `cemircol_free_data` to avoid memory leaks.
