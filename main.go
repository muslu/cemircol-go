package main

import (
	"fmt"
	"log"
	"github.com/muslu/cemircol-go"
)

func main() {
	// Önce Rust tarafını derlemeniz gerekir:
	// cargo build --release
	
	reader, err := cemircol.NewReader("data.cemir")
	if err != nil {
		log.Fatalf("Hata: %v", err)
	}
	defer reader.Close()

	fmt.Printf("Dosya açıldı. Satır sayısı: %d\n", reader.NumRows())

	data, err := reader.QueryFloat64("val")
	if err != nil {
		fmt.Printf("Sütun okunamadı: %v\n", err)
	} else {
		fmt.Printf("Okunan veri (ilk 5): %v\n", data[:min(5, len(data))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
