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

type Record struct {
	Value float64 `parquet:"name=value, type=DOUBLE"`
}

func main() {
	const numRows = 5_000_000
	const cemirFile = "benchmark_data.cemir"
	const parquetFile = "benchmark_data.parquet"

	fmt.Printf("🚀 5 Milyon satırlık (%d MB) veri ile benchmark başlatılıyor...\n\n", numRows*8/1024/1024)

	// Veri Hazırlama
	data := make([]float64, numRows)
	pData := make([]Record, numRows)
	for i := 0; i < numRows; i++ {
		v := float64(i) * 1.5
		data[i] = v
		pData[i] = Record{Value: v}
	}

	// --- CEMIRCOL TEST ---
	fmt.Println("🔷 CemirCol Ölçülüyor...")
	
	// Yazma
	start := time.Now()
	if err := cemircol.WriteFloat64(cemirFile, "val", data); err != nil {
		log.Fatalf("CemirCol Write Error: %v", err)
	}
	cemirWriteTime := time.Since(start)

	// Okuma
	start = time.Now()
	cReader, err := cemircol.NewReader(cemirFile)
	if err != nil {
		log.Fatalf("CemirCol Reader Error: %v", err)
	}
	readValues, err := cReader.QueryFloat64("val")
	if err != nil {
		log.Fatalf("CemirCol Query Error: %v", err)
	}
	cemirReadTime := time.Since(start)
	cReader.Close()
	_ = readValues

	cemirStat, _ := os.Stat(cemirFile)
	cemirSize := float64(cemirStat.Size()) / 1024 / 1024

	// --- PARQUET TEST ---
	fmt.Println("🔶 Parquet Ölçülüyor...")

	// Yazma
	start = time.Now()
	fw, err := local.NewLocalFileWriter(parquetFile)
	if err != nil {
		log.Fatalf("Parquet File Error: %v", err)
	}
	pw, err := writer.NewParquetWriter(fw, new(Record), 4)
	if err != nil {
		log.Fatalf("Parquet Writer Error: %v", err)
	}
	for _, r := range pData {
		if err := pw.Write(r); err != nil {
			log.Fatalf("Parquet Write Error: %v", err)
		}
	}
	pw.WriteStop()
	fw.Close()
	parquetWriteTime := time.Since(start)

	// Okuma
	start = time.Now()
	fr, err := local.NewLocalFileReader(parquetFile)
	if err != nil {
		log.Fatalf("Parquet Reader Error: %v", err)
	}
	pr, err := reader.NewParquetReader(fr, new(Record), 4)
	if err != nil {
		log.Fatalf("Parquet Reader Error: %v", err)
	}
	num := int(pr.GetNumRows())
	pRes := make([]Record, num)
	if err := pr.Read(&pRes); err != nil {
		log.Fatalf("Parquet Read Error: %v", err)
	}
	pr.ReadStop()
	fr.Close()
	parquetReadTime := time.Since(start)

	parquetStat, _ := os.Stat(parquetFile)
	parquetSize := float64(parquetStat.Size()) / 1024 / 1024

	// --- SONUÇLAR ---
	fmt.Println("\n🏆 BENCHMARK SONUÇLARI (5M Satır)")
	fmt.Println("----------------------------------------------------------------------")
	fmt.Printf("%-15s | %-15s | %-15s | %-10s\n", "Metric", "CemirCol-Go", "Parquet-Go", "Ratio")
	fmt.Println("----------------------------------------------------------------------")
	fmt.Printf("%-15s | %-15v | %-15v | %-10.2fx\n", "Write Time", cemirWriteTime, parquetWriteTime, float64(parquetWriteTime)/float64(cemirWriteTime))
	fmt.Printf("%-15s | %-15v | %-15v | %-10.2fx\n", "Read Time", cemirReadTime, parquetReadTime, float64(parquetReadTime)/float64(cemirReadTime))
	fmt.Printf("%-15s | %-15.2f MB | %-15.2f MB | %-10.2fx\n", "File Size", cemirSize, parquetSize, parquetSize/cemirSize)
	fmt.Println("----------------------------------------------------------------------")

	// Temizlik
	os.Remove(cemirFile)
	os.Remove(parquetFile)
}
