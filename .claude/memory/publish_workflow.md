---
name: Yayınlama Süreci
description: PyPI'a yeni sürüm yayınlama adımları — maturin, twine, sürüm senkronizasyonu
type: reference
---

Yeni sürüm yayınlarken şu dosyalar senkronize güncellenmelidir:
1. `Cargo.toml` → `version = "x.y.z"`
2. `pyproject.toml` → `version = "x.y.z"`

**Why:** Cargo ve PyPI sürüm numaraları ayrı — ikisi farklıysa `maturin build` yanlış wheel ismi üretir.
**How to apply:** Her ikisini de aynı anda güncelle, commit mesajına sürüm numarasını yaz.

## Komutlar
```bash
# Geliştirme ortamına kur (hızlı iterasyon)
maturin develop --release

# PyPI için wheel oluştur + yükle
./publish.sh
# → twine kullanıcı adı: __token__
# → twine şifre: pypi-... (PyPI API token)
```

## Dikkat
- `target/wheels/` klasörü root-owned olabilir (önceki sudo ile oluşturulmuş). `maturin develop` /tmp kullanır — bu durumda tercih et.
- `pip install -e . --no-build-isolation` yerine `maturin develop --release` tercih et (izin sorunlarını aşar).
