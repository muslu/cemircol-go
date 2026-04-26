# CemirCol-Go

CemirCol yüksek performanslı sütun tabanlı veri depolama formatının Go dili için kütüphanesidir. Rust çekirdeğini (core) C-ABI üzerinden kullanarak mmap ve sıfır-kopya (zero-copy) performansını Go ekosistemine taşır.

## Kurulum

Bu kütüphane Rust çekirdeğine bağımlıdır. Kullanmadan önce Rust tarafını derlemelisiniz:

```bash
# Rust kütüphanesini derleyin (staticlib oluşturur)
cargo build --release
```

Go tarafında bağımlılığı ekleyin:

```bash
go get github.com/muslu/cemircol-go
```

## Örnek Kullanım

```go
package main

import (
	"fmt"
	"github.com/muslu/cemircol-go"
)

func main() {
	reader, _ := cemircol.NewReader("data.cemir")
	defer reader.Close()

	fmt.Println("Satır sayısı:", reader.NumRows())
	data, _ := reader.QueryFloat64("score")
	fmt.Println("Veriler:", data)
}
```

## Neden Go?

Go, özellikle backend sistemlerinde ve veri işleme hatlarında (pipelines) hızı ve eşzamanlılık (concurrency) yetenekleriyle öne çıkar. CemirCol'un Rust çekirdeğini Go ile sarmalayarak:
- Python'un yavaşlığından kurtulursunuz.
- Go'nun `goroutine` yapısı ile milyonlarca satırı paralel işleyebilirsiniz.
- Bellek yönetimini Go'nun GC'sine (çöp toplayıcısına) bırakırken, ağır veri işleme işlerini Rust'ın performansına emanet edersiniz.
