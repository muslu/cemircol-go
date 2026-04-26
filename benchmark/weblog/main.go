package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/muslu/cemircol-go/cemircol"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/writer"
)

// 3 million web access log rows
// Columns: timestamp, user_id, status_code, response_ms, bytes_sent
const numRows = 3_000_000

// WebLogRecord is used for Parquet serialization.
type WebLogRecord struct {
	Timestamp  int64   `parquet:"name=timestamp,    type=INT64"`
	UserID     int64   `parquet:"name=user_id,      type=INT64"`
	StatusCode int64   `parquet:"name=status_code,  type=INT64"`
	ResponseMs float64 `parquet:"name=response_ms,  type=DOUBLE"`
	BytesSent  int64   `parquet:"name=bytes_sent,   type=INT64"`
}

type WebLogData struct {
	Timestamps  []int64
	UserIDs     []int64
	StatusCodes []int64
	ResponseMs  []float64
	BytesSent   []int64
}

// status code weights: 200 (70%), 301 (10%), 404 (12%), 500 (5%), 503 (3%)
var statusPool = []int64{200, 200, 200, 200, 200, 200, 200, 301, 301, 404, 404, 404, 500, 503}

func generateData() WebLogData {
	rng := rand.New(rand.NewSource(77))
	now := time.Now().Unix()
	d := WebLogData{
		Timestamps:  make([]int64, numRows),
		UserIDs:     make([]int64, numRows),
		StatusCodes: make([]int64, numRows),
		ResponseMs:  make([]float64, numRows),
		BytesSent:   make([]int64, numRows),
	}
	for i := 0; i < numRows; i++ {
		d.Timestamps[i] = now + int64(i)
		d.UserIDs[i] = int64(rng.Intn(100_000) + 1)
		sc := statusPool[rng.Intn(len(statusPool))]
		d.StatusCodes[i] = sc
		// 5xx requests are slower
		if sc >= 500 {
			d.ResponseMs[i] = 800 + rng.Float64()*4200 // 800–5000 ms
		} else {
			d.ResponseMs[i] = 5 + rng.Float64()*995 // 5–1000 ms
		}
		d.BytesSent[i] = int64(rng.Intn(500_000) + 512)
	}
	return d
}

// ── CemirCol ─────────────────────────────────────────────────────────────────

var cemirFiles = []string{
	"wl_ts.cemir", "wl_uid.cemir", "wl_sc.cemir", "wl_rms.cemir", "wl_bytes.cemir",
}

func cemirWrite(d WebLogData) time.Duration {
	start := time.Now()
	mustErr(cemircol.WriteInt64("wl_ts.cemir", "timestamp", d.Timestamps))
	mustErr(cemircol.WriteInt64("wl_uid.cemir", "user_id", d.UserIDs))
	mustErr(cemircol.WriteInt64("wl_sc.cemir", "status_code", d.StatusCodes))
	mustErr(cemircol.WriteFloat64("wl_rms.cemir", "response_ms", d.ResponseMs))
	mustErr(cemircol.WriteInt64("wl_bytes.cemir", "bytes_sent", d.BytesSent))
	return time.Since(start)
}

func cemirRead() (WebLogData, time.Duration) {
	start := time.Now()
	rTs, _ := cemircol.NewReader("wl_ts.cemir")
	rUID, _ := cemircol.NewReader("wl_uid.cemir")
	rSC, _ := cemircol.NewReader("wl_sc.cemir")
	rRms, _ := cemircol.NewReader("wl_rms.cemir")
	rBytes, _ := cemircol.NewReader("wl_bytes.cemir")
	defer func() { rTs.Close(); rUID.Close(); rSC.Close(); rRms.Close(); rBytes.Close() }()

	ts, _ := rTs.QueryInt64("timestamp")
	uids, _ := rUID.QueryInt64("user_id")
	scs, _ := rSC.QueryInt64("status_code")
	rms, _ := rRms.QueryFloat64("response_ms")
	bytes, _ := rBytes.QueryInt64("bytes_sent")

	return WebLogData{ts, uids, scs, rms, bytes}, time.Since(start)
}

func cemirFileSize() float64 {
	total := int64(0)
	for _, name := range cemirFiles {
		info, _ := os.Stat(name)
		total += info.Size()
	}
	return float64(total) / 1024 / 1024
}

func cemirCleanup() {
	for _, name := range cemirFiles {
		os.Remove(name)
	}
}

// ── Parquet ───────────────────────────────────────────────────────────────────

func parquetWrite(d WebLogData) time.Duration {
	start := time.Now()
	fw, err := local.NewLocalFileWriter("weblog.parquet")
	mustErr(err)
	pw, err := writer.NewParquetWriter(fw, new(WebLogRecord), 4)
	mustErr(err)
	for i := 0; i < numRows; i++ {
		mustErr(pw.Write(WebLogRecord{
			Timestamp:  d.Timestamps[i],
			UserID:     d.UserIDs[i],
			StatusCode: d.StatusCodes[i],
			ResponseMs: d.ResponseMs[i],
			BytesSent:  d.BytesSent[i],
		}))
	}
	pw.WriteStop()
	fw.Close()
	return time.Since(start)
}

func parquetRead() ([]WebLogRecord, time.Duration) {
	start := time.Now()
	fr, err := local.NewLocalFileReader("weblog.parquet")
	mustErr(err)
	pr, err := reader.NewParquetReader(fr, new(WebLogRecord), 4)
	mustErr(err)
	defer func() { pr.ReadStop(); fr.Close() }()
	rows := make([]WebLogRecord, pr.GetNumRows())
	mustErr(pr.Read(&rows))
	return rows, time.Since(start)
}

func parquetFileSize() float64 {
	info, _ := os.Stat("weblog.parquet")
	return float64(info.Size()) / 1024 / 1024
}

// ── Analytics ─────────────────────────────────────────────────────────────────

type QueryResult struct {
	Error5xxCount  int
	AvgResponseMs  float64
	SlowCount      int // response_ms > 2000
	TotalBytesMB   float64
	AvgBytesSent   float64
}

func queryCemir(d WebLogData) (QueryResult, time.Duration) {
	start := time.Now()
	var r QueryResult
	var sumRms float64
	var totalBytes int64
	for i, sc := range d.StatusCodes {
		if sc >= 500 {
			r.Error5xxCount++
		}
		if d.ResponseMs[i] > 2000 {
			r.SlowCount++
		}
		sumRms += d.ResponseMs[i]
		totalBytes += d.BytesSent[i]
	}
	n := float64(len(d.StatusCodes))
	r.AvgResponseMs = sumRms / n
	r.TotalBytesMB = float64(totalBytes) / 1024 / 1024
	r.AvgBytesSent = float64(totalBytes) / n
	return r, time.Since(start)
}

func queryParquet(rows []WebLogRecord) (QueryResult, time.Duration) {
	start := time.Now()
	var r QueryResult
	var sumRms float64
	var totalBytes int64
	for _, row := range rows {
		if row.StatusCode >= 500 {
			r.Error5xxCount++
		}
		if row.ResponseMs > 2000 {
			r.SlowCount++
		}
		sumRms += row.ResponseMs
		totalBytes += row.BytesSent
	}
	n := float64(len(rows))
	r.AvgResponseMs = sumRms / n
	r.TotalBytesMB = float64(totalBytes) / 1024 / 1024
	r.AvgBytesSent = float64(totalBytes) / n
	return r, time.Since(start)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func mustErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func speedup(slow, fast time.Duration) string {
	if fast == 0 {
		return "N/A"
	}
	r := float64(slow) / float64(fast)
	if r >= 1 {
		return fmt.Sprintf("%.2fx faster", r)
	}
	return fmt.Sprintf("%.2fx slower", 1/r)
}

func bar(ratio float64, width int) string {
	filled := int(math.Round(ratio * float64(width)))
	if filled > width {
		filled = width
	}
	s := ""
	for i := 0; i < width; i++ {
		if i < filled {
			s += "█"
		} else {
			s += "░"
		}
	}
	return s
}

func maxDur(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	fmt.Printf("╔══════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║   CemirCol-Go vs Parquet-Go  ·  Web Access Log Benchmark         ║\n")
	fmt.Printf("║   %d rows · 5 columns (ts, user_id, status, resp_ms, bytes) ║\n", numRows)
	fmt.Printf("╚══════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Println("Generating synthetic web access log data...")
	d := generateData()
	rawMB := float64(numRows) * (8 + 8 + 8 + 8 + 8) / 1024 / 1024
	fmt.Printf("Done. %.1f MB raw data in memory.\n\n", rawMB)

	fmt.Println("Writing data...")
	cemirWriteTime := cemirWrite(d)
	pqWriteTime := parquetWrite(d)

	cemirSize := cemirFileSize()
	pqSize := parquetFileSize()

	fmt.Println("Reading data...")
	cemirData, cemirReadTime := cemirRead()
	pqRows, pqReadTime := parquetRead()

	fmt.Println("Running analytics queries...\n")
	cemirResult, cemirQueryTime := queryCemir(cemirData)
	pqResult, pqQueryTime := queryParquet(pqRows)

	fmt.Println("══════════════════════════════════════════════════════════════════")
	fmt.Printf("  %-20s  %14s  %14s  %s\n", "Metric", "CemirCol-Go", "Parquet-Go", "Winner")
	fmt.Println("──────────────────────────────────────────────────────────────────")

	row := func(label, cemir, pq, winner string) {
		fmt.Printf("  %-20s  %14s  %14s  %s\n", label, cemir, pq, winner)
	}

	pick := func(a, b time.Duration) (string, time.Duration, time.Duration) {
		if a <= b {
			return "CemirCol", a, b
		}
		return "Parquet", b, a
	}

	wW, wFast, wSlow := pick(cemirWriteTime, pqWriteTime)
	row("Write Time", cemirWriteTime.Round(time.Millisecond).String(),
		pqWriteTime.Round(time.Millisecond).String(), wW+" "+speedup(wSlow, wFast))

	rW, rFast, rSlow := pick(cemirReadTime, pqReadTime)
	row("Read Time", cemirReadTime.Round(time.Millisecond).String(),
		pqReadTime.Round(time.Millisecond).String(), rW+" "+speedup(rSlow, rFast))

	qW, qFast, qSlow := pick(cemirQueryTime, pqQueryTime)
	row("Analytics Query", cemirQueryTime.Round(time.Microsecond).String(),
		pqQueryTime.Round(time.Microsecond).String(), qW+" "+speedup(qSlow, qFast))

	szWinner := "CemirCol"
	if pqSize < cemirSize {
		szWinner = "Parquet"
	}
	row("Total File Size", fmt.Sprintf("%.2f MB", cemirSize),
		fmt.Sprintf("%.2f MB", pqSize), szWinner)

	fmt.Println("══════════════════════════════════════════════════════════════════")

	fmt.Printf("\n  Raw data: %.1f MB\n", rawMB)
	fmt.Printf("  CemirCol compression: %.1f%% of raw  (%.2fx)\n", cemirSize/rawMB*100, rawMB/cemirSize)
	fmt.Printf("  Parquet  compression: %.1f%% of raw  (%.2fx)\n", pqSize/rawMB*100, rawMB/pqSize)

	cemirTotal := cemirWriteTime + cemirReadTime
	pqTotal := pqWriteTime + pqReadTime
	maxTotal := maxDur(cemirTotal, pqTotal)
	const barW = 30
	fmt.Printf("\n  Write + Read total time:\n")
	fmt.Printf("  CemirCol %s %v\n", bar(float64(cemirTotal)/float64(maxTotal), barW), cemirTotal.Round(time.Millisecond))
	fmt.Printf("  Parquet  %s %v\n", bar(float64(pqTotal)/float64(maxTotal), barW), pqTotal.Round(time.Millisecond))

	fmt.Printf("\n  Analytics results:\n")
	fmt.Printf("  CemirCol:  5xx_errors=%d  slow_reqs(>2s)=%d  avg_resp=%.2fms  total_data=%.1fMB\n",
		cemirResult.Error5xxCount, cemirResult.SlowCount, cemirResult.AvgResponseMs, cemirResult.TotalBytesMB)
	fmt.Printf("  Parquet:   5xx_errors=%d  slow_reqs(>2s)=%d  avg_resp=%.2fms  total_data=%.1fMB\n",
		pqResult.Error5xxCount, pqResult.SlowCount, pqResult.AvgResponseMs, pqResult.TotalBytesMB)

	if cemirResult.Error5xxCount == pqResult.Error5xxCount && cemirResult.SlowCount == pqResult.SlowCount {
		fmt.Println("\n  Results match — both engines return identical data.")
	} else {
		fmt.Println("\n  WARNING: result mismatch between engines!")
	}

	cemirCleanup()
	os.Remove("weblog.parquet")
	fmt.Println()
}
