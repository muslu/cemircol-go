package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/muslu/cemircol-go/cemircol"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/writer"
)

type Row struct {
	Val float64 `parquet:"name=val, type=DOUBLE"`
}

func main() {
	const numRows = 1_000_000
	const cemirFile = "comparison.cemir"
	const parquetFile = "comparison.parquet"

	data := make([]float64, numRows)
	rows := make([]Row, numRows)
	for i := 0; i < numRows; i++ {
		v := float64(i) * 0.1
		data[i] = v
		rows[i] = Row{Val: v}
	}

	fmt.Printf("📊 Karşılaştırma Başlatılıyor (%d satır)...\n\n", numRows)

	// --- CemirCol ---
	fmt.Println("🔷 CemirCol-Go Test Ediliyor...")
	start := time.Now()
	err := cemircol.WriteFloat64(cemirFile, "val", data)
	if err != nil {
		log.Fatalf("CemirCol Yazma Hatası: %v", err)
	}
	cemirWrite := time.Since(start)

	start = time.Now()
	cReader, err := cemircol.NewReader(cemirFile)
	if err != nil {
		log.Fatalf("CemirCol Okuma Hatası: %v", err)
	}
	_, err = cReader.QueryFloat64("val")
	if err != nil {
		log.Fatalf("CemirCol Sorgu Hatası: %v", err)
	}
	cReader.Close()
	cemirRead := time.Since(start)

	// --- Parquet ---
	fmt.Println("🔶 Parquet-Go Test Ediliyor...")
	start = time.Now()
	pf, err := local.NewLocalFileWriter(parquetFile)
	if err != nil {
		log.Fatalf("Parquet Dosya Hatası: %v", err)
	}
	pw, err := writer.NewParquetWriter(pf, new(Row), 4)
	if err != nil {
		log.Fatalf("Parquet Writer Hatası: %v", err)
	}
	for _, r := range rows {
		if err := pw.Write(r); err != nil {
			log.Fatalf("Parquet Yazma Hatası: %v", err)
		}
	}
	pw.WriteStop()
	pf.Close()
	parquetWrite := time.Since(start)

	start = time.Now()
	pf2, err := local.NewLocalFileReader(parquetFile)
	if err != nil {
		log.Fatalf("Parquet Açma Hatası: %v", err)
	}
	pr, err := reader.NewParquetReader(pf2, new(Row), 4)
	if err != nil {
		log.Fatalf("Parquet Reader Hatası: %v", err)
	}
	num := int(pr.GetNumRows())
	res := make([]Row, num)
	if err := pr.Read(&res); err != nil {
		log.Fatalf("Parquet Okuma Hatası: %v", err)
	}
	pr.ReadStop()
	pf2.Close()
	parquetRead := time.Since(start)

	// --- Sonuçlar ---
	fmt.Println("\n🏁 KARŞILAŞTIRMA SONUÇLARI:")
	fmt.Println("--------------------------------------------------")
	fmt.Printf("%-15s | %-15s | %-15s\n", "İşlem", "CemirCol-Go", "Parquet-Go")
	fmt.Println("--------------------------------------------------")
	fmt.Printf("%-15s | %-15v | %-15v\n", "Yazma (Write)", cemirWrite, parquetWrite)
	fmt.Printf("%-15s | %-15v | %-15v\n", "Okuma (Read)", cemirRead, parquetRead)
	fmt.Println("--------------------------------------------------")

	cStat, _ := os.Stat(cemirFile)
	pStat, _ := os.Stat(parquetFile)
	fmt.Printf("Dosya Boyutu   | %-15s | %-15s\n", 
		fmt.Sprintf("%.2f MB", float64(cStat.Size())/1024/1024),
		fmt.Sprintf("%.2f MB", float64(pStat.Size())/1024/1024))
}
