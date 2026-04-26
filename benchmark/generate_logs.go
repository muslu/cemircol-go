package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	f, _ := os.Create("mail.log")
	defer f.Close()

	statuses := []string{"sent", "deferred", "bounced", "expired"}
	domains := []string{"gmail.com", "yahoo.com", "outlook.com", "example.com"}

	now := time.Now()

	for i := 0; i < 100000; i++ {
		queueID := fmt.Sprintf("%010X", i+1000000)
		timestamp := now.Add(time.Duration(i) * time.Second).Format("Jan 2 15:04:05")
		
		sender := fmt.Sprintf("sender%d@myserver.com", rand.Intn(10))
		rcpt := fmt.Sprintf("user%d@%s", rand.Intn(100), domains[rand.Intn(len(domains))])
		status := statuses[rand.Intn(len(statuses))]

		// Log line 1: qmgr (from)
		fmt.Fprintf(f, "%s server postfix/qmgr[123]: %s: from=<%s>, size=%d, nrcpt=1\n", 
			timestamp, queueID, sender, rand.Intn(5000)+500)
		
		// Log line 2: smtp (to + status)
		delay := rand.Float64() * 2
		fmt.Fprintf(f, "%s server postfix/smtp[456]: %s: to=<%s>, relay=mx.%s, delay=%.1f, status=%s\n", 
			timestamp, queueID, rcpt, domains[rand.Intn(len(domains))], delay, status)
	}
	fmt.Println("✅ 1000 satırlık mail.log oluşturuldu.")
}
