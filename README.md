# CemirCol-Go

CemirCol yüksek performanslı sütun tabanlı veri depolama formatının Go dili için kütüphanesidir. Rust çekirdeğini (core) C-ABI üzerinden kullanarak mmap ve sıfır-kopya (zero-copy) performansını Go ekosistemine taşır.

## Kurulum

Bu kütüphane Rust çekirdeğine bağımlıdır. Kullanmadan önce Rust tarafını derlemelisiniz:

```bash
# Kurulum ve derleme için:
./setup.sh
```

Go tarafında bağımlılığı ekleyin:

```bash
go get github.com/muslu/cemircol-go/cemircol
```

## Performans (Benchmark)

10 milyon satırlık (`float64`) veri üzerinde yapılan test sonuçları:

- **Yazma Hızı:** ~50 Milyon satır/sn
- **Okuma Hızı:** ~75 Milyon satır/sn (mmap + zero-copy)
- **Dosya Boyutu:** ~80 MB (ham veri 80MB, zstd sıkıştırma ile veriye göre değişir)

Testi çalıştırmak için: `go run benchmark.go`

## Parquet ile Karşılaştırma (1 Milyon Satır)

`cemircol-go` ve `parquet-go` (xitongsys) kütüphaneleri arasında 1 milyon `float64` satırı ile yapılan karşılaştırma:

| İşlem | CemirCol-Go | Parquet-Go | Fark |
| :--- | :--- | :--- | :--- |
| **Yazma (Write)** | ~22ms | ~75ms | **3.4x Daha Hızlı** |
| **Okuma (Read)** | **~12ms** | ~130ms | **10.8x Daha Hızlı** |
| **Dosya Boyutu** | **1.02 MB** | 3.94 MB | **3.8x Daha Küçük** |

*Not: Okuma hızındaki muazzam fark, CemirCol'un mmap ve sıfır-kopya (zero-copy) mimarisinden kaynaklanmaktadır.*

Testi çalıştırmak için: `go run comparison_benchmark.go`

## Örnek Kullanım

```go
package main

import (
	"fmt"
	"github.com/muslu/cemircol-go/cemircol"
)

func main() {
	// Veri yazma
	data := []float64{1.1, 2.2, 3.3}
	cemircol.WriteFloat64("data.cemir", "val", data)

	// Veri okuma
	reader, _ := cemircol.NewReader("data.cemir")
	defer reader.Close()

	fmt.Println("Satır sayısı:", reader.NumRows())
	val, _ := reader.QueryFloat64("val")
	fmt.Println("Veriler:", val)
}
```

## Neden Go?

Go, özellikle backend sistemlerinde ve veri işleme hatlarında (pipelines) hızı ve eşzamanlılık (concurrency) yetenekleriyle öne çıkar. CemirCol'un Rust çekirdeğini Go ile sarmalayarak:
- Python'un yavaşlığından kurtulursunuz.
- Go'nun `goroutine` yapısı ile milyonlarca satırı paralel işleyebilirsiniz.
- Bellek yönetimini Go'nun GC'sine (çöp toplayıcısına) bırakırken, ağır veri işleme işlerini Rust'ın performansına emanet edersiniz.
