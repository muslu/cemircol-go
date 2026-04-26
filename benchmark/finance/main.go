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

const numRows = 2_000_000

// OHLCVRecord is used for Parquet serialization.
type OHLCVRecord struct {
	Timestamp int64   `parquet:"name=timestamp, type=INT64"`
	Open      float64 `parquet:"name=open,      type=DOUBLE"`
	High      float64 `parquet:"name=high,      type=DOUBLE"`
	Low       float64 `parquet:"name=low,       type=DOUBLE"`
	Close     float64 `parquet:"name=close,     type=DOUBLE"`
	Volume    int64   `parquet:"name=volume,    type=INT64"`
}

type OHLCVData struct {
	Timestamps []int64
	Opens      []float64
	Highs      []float64
	Lows       []float64
	Closes     []float64
	Volumes    []int64
}

func generateData() OHLCVData {
	rng := rand.New(rand.NewSource(99))
	now := time.Now().Unix()
	d := OHLCVData{
		Timestamps: make([]int64, numRows),
		Opens:      make([]float64, numRows),
		Highs:      make([]float64, numRows),
		Lows:       make([]float64, numRows),
		Closes:     make([]float64, numRows),
		Volumes:    make([]int64, numRows),
	}
	price := 100.0
	for i := 0; i < numRows; i++ {
		d.Timestamps[i] = now + int64(i*60) // 1-minute candles
		open := price
		change := (rng.Float64() - 0.5) * 4
		close_ := math.Max(0.01, open+change)
		high := math.Max(open, close_) + rng.Float64()*2
		low := math.Min(open, close_) - rng.Float64()*2
		if low < 0.01 {
			low = 0.01
		}
		d.Opens[i] = open
		d.Highs[i] = high
		d.Lows[i] = low
		d.Closes[i] = close_
		d.Volumes[i] = int64(rng.Intn(1_000_000) + 10_000)
		price = close_
	}
	return d
}

// ── CemirCol ─────────────────────────────────────────────────────────────────

var cemirFiles = []string{
	"fin_ts.cemir", "fin_open.cemir", "fin_high.cemir",
	"fin_low.cemir", "fin_close.cemir", "fin_vol.cemir",
}

func cemirWrite(d OHLCVData) time.Duration {
	start := time.Now()
	mustErr(cemircol.WriteInt64("fin_ts.cemir", "timestamp", d.Timestamps))
	mustErr(cemircol.WriteFloat64("fin_open.cemir", "open", d.Opens))
	mustErr(cemircol.WriteFloat64("fin_high.cemir", "high", d.Highs))
	mustErr(cemircol.WriteFloat64("fin_low.cemir", "low", d.Lows))
	mustErr(cemircol.WriteFloat64("fin_close.cemir", "close", d.Closes))
	mustErr(cemircol.WriteInt64("fin_vol.cemir", "volume", d.Volumes))
	return time.Since(start)
}

func cemirRead() (OHLCVData, time.Duration) {
	start := time.Now()
	open := func(name, col string) interface{} { return name + col }
	_ = open

	rTs, _ := cemircol.NewReader("fin_ts.cemir")
	rOpen, _ := cemircol.NewReader("fin_open.cemir")
	rHigh, _ := cemircol.NewReader("fin_high.cemir")
	rLow, _ := cemircol.NewReader("fin_low.cemir")
	rClose, _ := cemircol.NewReader("fin_close.cemir")
	rVol, _ := cemircol.NewReader("fin_vol.cemir")
	defer func() {
		rTs.Close(); rOpen.Close(); rHigh.Close()
		rLow.Close(); rClose.Close(); rVol.Close()
	}()

	ts, err := rTs.QueryInt64("timestamp")
	mustErr(err)
	opens, err := rOpen.QueryFloat64("open")
	mustErr(err)
	highs, err := rHigh.QueryFloat64("high")
	mustErr(err)
	lows, err := rLow.QueryFloat64("low")
	mustErr(err)
	closes, err := rClose.QueryFloat64("close")
	mustErr(err)
	vols, err := rVol.QueryInt64("volume")
	mustErr(err)

	return OHLCVData{ts, opens, highs, lows, closes, vols}, time.Since(start)
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

func parquetWrite(d OHLCVData) time.Duration {
	start := time.Now()
	fw, err := local.NewLocalFileWriter("finance.parquet")
	mustErr(err)
	pw, err := writer.NewParquetWriter(fw, new(OHLCVRecord), 4)
	mustErr(err)
	for i := 0; i < numRows; i++ {
		mustErr(pw.Write(OHLCVRecord{
			Timestamp: d.Timestamps[i],
			Open:      d.Opens[i],
			High:      d.Highs[i],
			Low:       d.Lows[i],
			Close:     d.Closes[i],
			Volume:    d.Volumes[i],
		}))
	}
	pw.WriteStop()
	fw.Close()
	return time.Since(start)
}

func parquetRead() ([]OHLCVRecord, time.Duration) {
	start := time.Now()
	fr, err := local.NewLocalFileReader("finance.parquet")
	mustErr(err)
	pr, err := reader.NewParquetReader(fr, new(OHLCVRecord), 4)
	mustErr(err)
	defer func() { pr.ReadStop(); fr.Close() }()
	rows := make([]OHLCVRecord, pr.GetNumRows())
	mustErr(pr.Read(&rows))
	return rows, time.Since(start)
}

func parquetFileSize() float64 {
	info, _ := os.Stat("finance.parquet")
	return float64(info.Size()) / 1024 / 1024
}

// ── Analytics ─────────────────────────────────────────────────────────────────

type AnalyticsResult struct {
	BullishCount int
	BearishCount int
	MaxVolume    int64
	MaxVolIdx    int
	AvgRange     float64 // avg (high - low)
}

func queryCemir(d OHLCVData) (AnalyticsResult, time.Duration) {
	start := time.Now()
	var r AnalyticsResult
	var sumRange float64
	for i := range d.Opens {
		if d.Closes[i] > d.Opens[i] {
			r.BullishCount++
		} else {
			r.BearishCount++
		}
		if d.Volumes[i] > r.MaxVolume {
			r.MaxVolume = d.Volumes[i]
			r.MaxVolIdx = i
		}
		sumRange += d.Highs[i] - d.Lows[i]
	}
	r.AvgRange = sumRange / float64(len(d.Opens))
	return r, time.Since(start)
}

func queryParquet(rows []OHLCVRecord) (AnalyticsResult, time.Duration) {
	start := time.Now()
	var r AnalyticsResult
	var sumRange float64
	for i, row := range rows {
		if row.Close > row.Open {
			r.BullishCount++
		} else {
			r.BearishCount++
		}
		if row.Volume > r.MaxVolume {
			r.MaxVolume = row.Volume
			r.MaxVolIdx = i
		}
		sumRange += row.High - row.Low
	}
	r.AvgRange = sumRange / float64(len(rows))
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
	fmt.Printf("╔═══════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║   CemirCol-Go vs Parquet-Go  ·  Financial OHLCV Benchmark         ║\n")
	fmt.Printf("║   %d rows · 6 columns (ts, open, high, low, close, volume)  ║\n", numRows)
	fmt.Printf("╚═══════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Println("Generating synthetic OHLCV data...")
	d := generateData()
	rawMB := float64(numRows) * (8*5 + 8) / 1024 / 1024
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

	wWinner, wFast, wSlow := "CemirCol", cemirWriteTime, pqWriteTime
	if pqWriteTime < cemirWriteTime {
		wWinner, wFast, wSlow = "Parquet", pqWriteTime, cemirWriteTime
	}
	row("Write Time", cemirWriteTime.Round(time.Millisecond).String(),
		pqWriteTime.Round(time.Millisecond).String(), wWinner+" "+speedup(wSlow, wFast))

	rWinner, rFast, rSlow := "CemirCol", cemirReadTime, pqReadTime
	if pqReadTime < cemirReadTime {
		rWinner, rFast, rSlow = "Parquet", pqReadTime, cemirReadTime
	}
	row("Read Time", cemirReadTime.Round(time.Millisecond).String(),
		pqReadTime.Round(time.Millisecond).String(), rWinner+" "+speedup(rSlow, rFast))

	qWinner, qFast, qSlow := "CemirCol", cemirQueryTime, pqQueryTime
	if pqQueryTime < cemirQueryTime {
		qWinner, qFast, qSlow = "Parquet", pqQueryTime, cemirQueryTime
	}
	row("Analytics Query", cemirQueryTime.Round(time.Microsecond).String(),
		pqQueryTime.Round(time.Microsecond).String(), qWinner+" "+speedup(qSlow, qFast))

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

	fmt.Printf("\n  Analytics results (bullish/bearish/max-volume/avg-range):\n")
	fmt.Printf("  CemirCol:  bullish=%d  bearish=%d  max_vol=%d  avg_range=%.4f\n",
		cemirResult.BullishCount, cemirResult.BearishCount, cemirResult.MaxVolume, cemirResult.AvgRange)
	fmt.Printf("  Parquet:   bullish=%d  bearish=%d  max_vol=%d  avg_range=%.4f\n",
		pqResult.BullishCount, pqResult.BearishCount, pqResult.MaxVolume, pqResult.AvgRange)

	if cemirResult.BullishCount == pqResult.BullishCount && cemirResult.MaxVolume == pqResult.MaxVolume {
		fmt.Println("\n  Results match — both engines return identical data.")
	} else {
		fmt.Println("\n  WARNING: result mismatch between engines!")
	}

	cemirCleanup()
	os.Remove("finance.parquet")
	fmt.Println()
}
