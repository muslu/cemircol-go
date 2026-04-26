# CLAUDE.md — cemircol

Bu dosya Claude Code'un bu projede çalışırken okuması gereken bağlamı sağlar.

## Hafıza Dosyaları

Proje hafızası `.claude/memory/` klasöründe tutulur:

- [Proje Genel Bakış](.claude/memory/project_overview.md) — dosya formatı, mimari, bağımlılıklar
- [Performans Optimizasyonları](.claude/memory/performance_optimizations.md) — yapılan optimizasyonlar ve gerekçeleri
- [Kullanıcı Profili](.claude/memory/user_profile.md) — proje sahipleri hakkında bilgi
- [Yayınlama Süreci](.claude/memory/publish_workflow.md) — PyPI yayınlama adımları

## Hızlı Başvuru

### Derleme
```bash
maturin develop --release   # geliştirme ortamı
./publish.sh                # PyPI yayınlama
```

### Benchmark
```bash
python benchmark.py
```

### Kritik Kurallar

1. **Dosya formatını bozma.** Yeni alan eklerken `#[serde(default)]` kullan — eski dosyalar okunabilir kalmalı.
2. **PyList döndürme.** `query()` her zaman numpy array veya array.array döndürmeli; 10M nesne yaratmak yasak.
3. **Versiyon senkronizasyonu.** `Cargo.toml` ve `pyproject.toml` versiyonları her zaman aynı olmalı.
4. **`target/wheels/` izin sorunu.** Bu klasör root-owned olabilir. `maturin develop --release` tercih et, `pip install -e .` değil.
5. **`allow_threads` pyo3 0.28'de yok.** GIL release için alternatif ara. Rayon thread'leri GIL'e dokunmaz, doğrudan çağır.

## Mevcut Performans (v0.1.4, 10M satır)

| | Dosya Boyutu | Okuma Süresi |
|---|---|---|
| **CemirCol** | **5.52 MB** | **0.108 s** |
| Parquet | 85.09 MB | 0.113 s |
| CSV | 182.61 MB | 6.6 s |
| JSON | 192.15 MB | 1.2 s |
