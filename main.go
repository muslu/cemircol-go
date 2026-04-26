package main

import (
	"fmt"
	"log"
	"time"
	"github.com/muslu/cemircol-go/cemircol"
)

func main() {
	// Örnek: 1 milyon satırlık veri oluştur
	const numRows = 1_000_000
	const filename = "example.cemir"

	data := make([]float64, numRows)
	for i := 0; i < numRows; i++ {
		data[i] = float64(i) * 1.23
	}

	fmt.Printf("📝 %d satır yazılıyor...\n", numRows)
	start := time.Now()
	err := cemircol.WriteFloat64(filename, "test_col", data)
	if err != nil {
		log.Fatalf("Yazma hatası: %v", err)
	}
	fmt.Printf("✅ Yazma süresi: %v\n", time.Since(start))

	fmt.Printf("📖 %s okunuyor...\n", filename)
	start = time.Now()
	reader, err := cemircol.NewReader(filename)
	if err != nil {
		log.Fatalf("Okuma hatası: %v", err)
	}
	defer reader.Close()

	val, err := reader.QueryFloat64("test_col")
	if err != nil {
		log.Fatalf("Sorgu hatası: %v", err)
	}
	
	fmt.Printf("✅ Okuma süresi: %v\n", time.Since(start))
	fmt.Printf("📊 Satır sayısı: %d\n", reader.NumRows())
	fmt.Printf("🔍 İlk 5 değer: %v\n", val[:5])
}
