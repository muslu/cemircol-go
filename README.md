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

1 milyon `float64` satırı üzerinde yapılan karşılaştırma sonuçları:

| İşlem | CemirCol-Go | Parquet-Go (xitongsys) | Fark |
| :--- | :--- | :--- | :--- |
| **Yazma (Write)** | **~22ms** | ~75ms | **3.4x Daha Hızlı** |
| **Okuma (Read)** | **~12ms** | ~130ms | **10.8x Daha Hızlı** |
| **Dosya Boyutu** | **1.02 MB** | 3.94 MB | **3.8x Daha Küçük** |

*Not: Benchmark sonuçları veri tipine ve donanıma göre değişiklik gösterebilir. Testleri kendiniz çalıştırmak için `benchmark/` dizinine göz atın.*

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
*   `benchmark/`: Performans testleri ve log işleme örnekleri.
*   `setup.sh`: Geliştirme ortamı kurulum scripti.
*   `publish.sh`: Git yayınlama ve versiyonlama scripti.

## 📄 Lisans

Bu proje MIT lisansı ile lisanslanmıştır. Daha fazla bilgi için `LICENSE` dosyasına bakınız.
