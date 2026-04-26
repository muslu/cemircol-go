# CemirCol-Go 🚀

**CemirCol-Go**, yüksek performanslı sütun tabanlı (columnar) veri depolama formatının Go dili için optimize edilmiş kütüphanesidir. Rust çekirdeğini (core) C-ABI üzerinden kullanarak **mmap** ve **sıfır-kopya (zero-copy)** performansını Go ekosistemine taşır.

[![Go Report Card](https://goreportcard.com/badge/github.com/muslu/cemircol-go)](https://goreportcard.com/report/github.com/muslu/cemircol-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## ✨ Özellikler

*   🚀 **Ultra Hızlı Okuma:** mmap sayesinde dosyayı belleğe yüklemeden anında erişim.
*   💎 **Sıfır-Kopya (Zero-Copy):** Rust'tan Go'ya veri aktarırken ek bellek kopyalaması yapmaz.
*   📦 **Yüksek Sıkıştırma:** Zstd algoritması ile Parquet'den daha küçük dosya boyutları.
*   🛠️ **Dictionary Encoding:** Tekrar eden metin verileri (e-posta, durum kodları vb.) için optimize edilmiş depolama.
*   🔌 **Kolay Entegrasyon:** Otomatik kurulum scriptleri ve basit API.

## 📊 Performans (Benchmark)

Üç farklı gerçek dünya senaryosunda CemirCol-Go ile Parquet-Go karşılaştırması.  
Tüm testler aynı donanımda çalıştırılmış; sonuçlar deterministik rastgele veri ile üretilmiştir.

> Testleri kendiniz çalıştırmak için `benchmark/` dizinine göz atın.

---

### Senaryo 1 — IoT Sensör Verisi

**1.000.000 satır · 4 kolon** (`timestamp`, `device_id`, `temperature`, `humidity`)  
Script: `benchmark/sensor/main.go`

| İşlem | CemirCol-Go | Parquet-Go | Fark |
| :--- | :---: | :---: | :--- |
| **Yazma (Write)** | **68 ms** | 171 ms | **2.5x daha hızlı** |
| **Okuma (Read)** | **43 ms** | 219 ms | **5.1x daha hızlı** |
| **Filtre Sorgusu** | **3.2 ms** | 3.9 ms | **1.2x daha hızlı** |
| **Dosya Boyutu** | **16.9 MB** | 22.5 MB | **1.3x daha küçük** |
| **Sıkıştırma Oranı** | **1.80x** | 1.36x | Ham 30.5 MB veriden |

---

### Senaryo 2 — Finansal OHLCV Verisi

**2.000.000 satır · 6 kolon** (`timestamp`, `open`, `high`, `low`, `close`, `volume`)  
Script: `benchmark/finance/main.go`

| İşlem | CemirCol-Go | Parquet-Go | Fark |
| :--- | :---: | :---: | :--- |
| **Yazma (Write)** | **231 ms** | 458 ms | **2.0x daha hızlı** |
| **Okuma (Read)** | **129 ms** | 630 ms | **4.9x daha hızlı** |
| **Analitik Sorgu** | **11.4 ms** | 12.6 ms | **1.1x daha hızlı** |
| **Dosya Boyutu** | **64.9 MB** | 79.8 MB | **1.2x daha küçük** |
| **Sıkıştırma Oranı** | **1.41x** | 1.15x | Ham 91.6 MB veriden |

---

### Senaryo 3 — Web Erişim Logu Analitiği

**3.000.000 satır · 5 kolon** (`timestamp`, `user_id`, `status_code`, `response_ms`, `bytes_sent`)  
Script: `benchmark/weblog/main.go`

| İşlem | CemirCol-Go | Parquet-Go | Fark |
| :--- | :---: | :---: | :--- |
| **Yazma (Write)** | **355 ms** | 564 ms | **1.6x daha hızlı** |
| **Okuma (Read)** | **168 ms** | 846 ms | **5.0x daha hızlı** |
| **Analitik Sorgu** | 11.8 ms | **10.6 ms** | Parquet 1.1x daha hızlı |
| **Dosya Boyutu** | **44.0 MB** | 67.4 MB | **1.5x daha küçük** |
| **Sıkıştırma Oranı** | **2.60x** | 1.70x | Ham 114.4 MB veriden |

---

### Özet

| | Yazma | Okuma | Dosya Boyutu |
| :--- | :---: | :---: | :---: |
| **Ortalama Hız Farkı** | **~2x hızlı** | **~5x hızlı** | **~1.3x küçük** |

CemirCol-Go okuma performansında Parquet'e karşı tutarlı biçimde **~5x** üstünlük sağlar.  
Dosya boyutunda Parquet'in kendi sıkıştırması rekabetçi kalsa da mmap tabanlı okuma,  
büyük veri setlerinde belleğe alma (deserialization) maliyetini tamamen ortadan kaldırır.

## 🛠️ Kurulum

Bu kütüphane Rust çekirdeğine bağımlıdır. Tüm kurulum ve derleme işlemlerini tek komutla yapabilirsiniz:

```bash
# Otomatik kurulum ve derleme (Rust & Go)
./setup.sh
```

## 💻 Örnek Kullanım

```go
package main

import (
	"fmt"
	"github.com/muslu/cemircol-go/cemircol"
)

func main() {
	// Veri yazma (Int64 örneği)
	data := []int64{100, 200, 300, 400, 500}
	cemircol.WriteInt64("data.cemir", "score", data)

	// Veri okuma
	reader, _ := cemircol.NewReader("data.cemir")
	defer reader.Close()

	fmt.Println("Toplam Satır:", reader.NumRows())
	
	// Sütun sorgulama (Sıfır-kopya performansıyla)
	scores, _ := reader.QueryInt64("score")
	fmt.Println("Skorlar:", scores)
}
```

## 📂 Proje Yapısı

*   `cemircol/`: Go sarmalayıcı (CGO).
*   `src/`: Rust çekirdek implementasyonu (Reader, Writer, C-ABI).
*   `benchmark/`: Performans testleri ve karşılaştırma scriptleri.
    *   `sensor/` — IoT sensör verisi (1M satır, 4 kolon)
    *   `finance/` — Finansal OHLCV verisi (2M satır, 6 kolon)
    *   `weblog/` — Web erişim logu analitiği (3M satır, 5 kolon)
    *   `compare_all/` — Tek kolon float64 büyük veri karşılaştırması
    *   `gen/`, `parser/`, `logger/`, `pq_parser/`, `pq_logger/` — Log işleme örnekleri
*   `setup.sh`: Geliştirme ortamı kurulum scripti.
*   `publish.sh`: Git yayınlama ve versiyonlama scripti.

## 📄 Lisans

Bu proje MIT lisansı ile lisanslanmıştır. Daha fazla bilgi için `LICENSE` dosyasına bakınız.
