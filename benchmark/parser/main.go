package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/muslu/cemircol-go/cemircol"
)

type PostfixLogEntry struct {
	Timestamp int64
	Recipient string
	Status    string
}

func main() {
	file, err := os.Open("mail.log")
	if err != nil {
		log.Fatalf("Log dosyası açılamadı: %v", err)
	}
	defer file.Close()

	fmt.Println("🔍 mail.log ayrıştırılıyor...")

	reQueueID := regexp.MustCompile(`([A-Z0-9]{10}):`)
	reTo := regexp.MustCompile(`to=<([^>]+)>`)
	reStatus := regexp.MustCompile(`status=([a-z]+)`)

	queueData := make(map[string]*PostfixLogEntry)
	var finalEntries []PostfixLogEntry
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
			queueData[qid] = &PostfixLogEntry{Timestamp: t.Unix()}
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

	fmt.Println("💾 Veriler CemirCol formatına yazılıyor...")

	emailDict := make(map[string]int64)
	var emailList []string
	statusDict := map[string]int64{"sent": 0, "deferred": 1, "bounced": 2, "expired": 3}

	var timestamps []int64
	var rcptIDs []int64
	var statusIDs []int64

	for _, entry := range finalEntries {
		if _, ok := emailDict[entry.Recipient]; !ok {
			emailDict[entry.Recipient] = int64(len(emailList))
			emailList = append(emailList, entry.Recipient)
		}
		timestamps = append(timestamps, entry.Timestamp)
		rcptIDs = append(rcptIDs, emailDict[entry.Recipient])
		statusIDs = append(statusIDs, statusDict[entry.Status])
	}

	cemircol.WriteInt64("logs_time.cemir", "time", timestamps)
	cemircol.WriteInt64("logs_rcpt.cemir", "rcpt", rcptIDs)
	cemircol.WriteInt64("logs_status.cemir", "status", statusIDs)

	dictFile, _ := os.Create("emails.dict")
	json.NewEncoder(dictFile).Encode(emailList)
	dictFile.Close()

	fmt.Printf("✅ %d mail kaydı işlendi ve CemirCol dosyalarına kaydedildi.\n", len(finalEntries))
}
