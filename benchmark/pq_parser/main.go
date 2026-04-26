package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/writer"
)

type MailRecord struct {
	Timestamp int64  `parquet:"name=timestamp, type=INT64"`
	Recipient string `parquet:"name=recipient, type=BYTE_ARRAY, convertedtype=UTF8"`
	Status    string `parquet:"name=status, type=BYTE_ARRAY, convertedtype=UTF8"`
}

func main() {
	file, err := os.Open("mail.log")
	if err != nil {
		log.Fatalf("Log dosyası bulunamadı. Önce generate_logs.go çalıştırın: %v", err)
	}
	defer file.Close()

	fmt.Println("🔍 Parquet: mail.log ayrıştırılıyor...")

	reQueueID := regexp.MustCompile(`([A-Z0-9]{10}):`)
	reTo := regexp.MustCompile(`to=<([^>]+)>`)
	reStatus := regexp.MustCompile(`status=([a-z]+)`)

	queueData := make(map[string]*MailRecord)
	var finalEntries []MailRecord
	currentYear := time.Now().Year()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		qMatch := reQueueID.FindStringSubmatch(line)
		if len(qMatch) < 2 {
			continue
		}
		qid := qMatch[1]

		if _, ok := queueData[qid]; !ok {
			parts := strings.Fields(line)
			if len(parts) < 3 {
				continue
			}
			tsStr := fmt.Sprintf("%s %s %s %d", parts[0], parts[1], parts[2], currentYear)
			t, _ := time.Parse("Jan 2 15:04:05 2006", tsStr)
			queueData[qid] = &MailRecord{Timestamp: t.Unix()}
		}

		if strings.Contains(line, "to=<") {
			toMatch := reTo.FindStringSubmatch(line)
			statusMatch := reStatus.FindStringSubmatch(line)
			if len(toMatch) > 1 && len(statusMatch) > 1 {
				queueData[qid].Recipient = strings.ToLower(toMatch[1])
				queueData[qid].Status = statusMatch[1]
				finalEntries = append(finalEntries, *queueData[qid])
			}
		}
	}

	fmt.Println("💾 Veriler Parquet formatına yazılıyor...")

	fw, err := local.NewLocalFileWriter("logs.parquet")
	if err != nil {
		log.Fatalf("Parquet dosya hatası: %v", err)
	}
	pw, err := writer.NewParquetWriter(fw, new(MailRecord), 4)
	if err != nil {
		log.Fatalf("Parquet writer hatası: %v", err)
	}

	for _, entry := range finalEntries {
		if err := pw.Write(entry); err != nil {
			log.Fatalf("Yazma hatası: %v", err)
		}
	}

	pw.WriteStop()
	fw.Close()

	fmt.Printf("✅ %d mail kaydı Parquet dosyasına kaydedildi.\n", len(finalEntries))
}
