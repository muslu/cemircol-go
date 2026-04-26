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

const numRows = 1_000_000

// SensorRecord is used for Parquet serialization.
type SensorRecord struct {
	Timestamp  int64   `parquet:"name=timestamp,  type=INT64"`
	DeviceID   int64   `parquet:"name=device_id,  type=INT64"`
	Temp       float64 `parquet:"name=temp,       type=DOUBLE"`
	Humidity   float64 `parquet:"name=humidity,   type=DOUBLE"`
}

// SensorData holds all columns in columnar layout for CemirCol.
type SensorData struct {
	Timestamps []int64
	DeviceIDs  []int64
	Temps      []float64
	Humidities []float64
}

func generateData() SensorData {
	rng := rand.New(rand.NewSource(42))
	now := time.Now().Unix()
	d := SensorData{
		Timestamps: make([]int64, numRows),
		DeviceIDs:  make([]int64, numRows),
		Temps:      make([]float64, numRows),
		Humidities: make([]float64, numRows),
	}
	for i := 0; i < numRows; i++ {
		d.Timestamps[i] = now + int64(i*10)
		d.DeviceIDs[i] = int64(rng.Intn(500) + 1)
		d.Temps[i] = 15.0 + rng.Float64()*35.0     // 15–50 °C
		d.Humidities[i] = 20.0 + rng.Float64()*70.0 // 20–90 %
	}
	return d
}

// ── CemirCol ─────────────────────────────────────────────────────────────────

func cemirWrite(d SensorData) time.Duration {
	start := time.Now()
	must(cemircol.WriteInt64("sensor_ts.cemir", "timestamp", d.Timestamps))
	must(cemircol.WriteInt64("sensor_dev.cemir", "device_id", d.DeviceIDs))
	must(cemircol.WriteFloat64("sensor_temp.cemir", "temp", d.Temps))
	must(cemircol.WriteFloat64("sensor_hum.cemir", "humidity", d.Humidities))
	return time.Since(start)
}

func cemirRead() ([]int64, []int64, []float64, []float64, time.Duration) {
	start := time.Now()

	rTs, err := cemircol.NewReader("sensor_ts.cemir")
	mustErr(err)
	defer rTs.Close()
	rDev, err := cemircol.NewReader("sensor_dev.cemir")
	mustErr(err)
	defer rDev.Close()
	rTemp, err := cemircol.NewReader("sensor_temp.cemir")
	mustErr(err)
	defer rTemp.Close()
	rHum, err := cemircol.NewReader("sensor_hum.cemir")
	mustErr(err)
	defer rHum.Close()

	ts, err := rTs.QueryInt64("timestamp")
	mustErr(err)
	devs, err := rDev.QueryInt64("device_id")
	mustErr(err)
	temps, err := rTemp.QueryFloat64("temp")
	mustErr(err)
	hums, err := rHum.QueryFloat64("humidity")
	mustErr(err)

	return ts, devs, temps, hums, time.Since(start)
}

func cemirFileSize() float64 {
	total := int64(0)
	for _, name := range []string{"sensor_ts.cemir", "sensor_dev.cemir", "sensor_temp.cemir", "sensor_hum.cemir"} {
		info, _ := os.Stat(name)
		total += info.Size()
	}
	return float64(total) / 1024 / 1024
}

func cemirCleanup() {
	for _, name := range []string{"sensor_ts.cemir", "sensor_dev.cemir", "sensor_temp.cemir", "sensor_hum.cemir"} {
		os.Remove(name)
	}
}

// ── Parquet ───────────────────────────────────────────────────────────────────

func parquetWrite(d SensorData) time.Duration {
	start := time.Now()

	fw, err := local.NewLocalFileWriter("sensor.parquet")
	mustErr(err)
	pw, err := writer.NewParquetWriter(fw, new(SensorRecord), 4)
	mustErr(err)

	for i := 0; i < numRows; i++ {
		mustErr(pw.Write(SensorRecord{
			Timestamp: d.Timestamps[i],
			DeviceID:  d.DeviceIDs[i],
			Temp:      d.Temps[i],
			Humidity:  d.Humidities[i],
		}))
	}
	pw.WriteStop()
	fw.Close()

	return time.Since(start)
}

func parquetRead() ([]SensorRecord, time.Duration) {
	start := time.Now()

	fr, err := local.NewLocalFileReader("sensor.parquet")
	mustErr(err)
	pr, err := reader.NewParquetReader(fr, new(SensorRecord), 4)
	mustErr(err)
	defer func() { pr.ReadStop(); fr.Close() }()

	rows := make([]SensorRecord, pr.GetNumRows())
	mustErr(pr.Read(&rows))

	return rows, time.Since(start)
}

func parquetFileSize() float64 {
	info, _ := os.Stat("sensor.parquet")
	return float64(info.Size()) / 1024 / 1024
}

// ── Analytics ─────────────────────────────────────────────────────────────────

type QueryResult struct {
	HotCount    int
	AvgTemp     float64
	AvgHumidity float64
}

func queryCemir(temps, hums []float64, threshold float64) (QueryResult, time.Duration) {
	start := time.Now()
	var sumT, sumH float64
	count := 0
	for i, t := range temps {
		if t > threshold {
			sumT += t
			sumH += hums[i]
			count++
		}
	}
	elapsed := time.Since(start)
	if count == 0 {
		return QueryResult{}, elapsed
	}
	return QueryResult{count, sumT / float64(count), sumH / float64(count)}, elapsed
}

func queryParquet(rows []SensorRecord, threshold float64) (QueryResult, time.Duration) {
	start := time.Now()
	var sumT, sumH float64
	count := 0
	for _, r := range rows {
		if r.Temp > threshold {
			sumT += r.Temp
			sumH += r.Humidity
			count++
		}
	}
	elapsed := time.Since(start)
	if count == 0 {
		return QueryResult{}, elapsed
	}
	return QueryResult{count, sumT / float64(count), sumH / float64(count)}, elapsed
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func mustErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func speedup(base, cmp time.Duration) string {
	r := float64(base) / float64(cmp)
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
	b := ""
	for i := 0; i < width; i++ {
		if i < filled {
			b += "█"
		} else {
			b += "░"
		}
	}
	return b
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	const tempThreshold = 38.0

	fmt.Printf("╔══════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║      CemirCol-Go vs Parquet-Go  ·  IoT Sensor Data Benchmark      ║\n")
	fmt.Printf("║      %d rows · 4 columns (timestamp, device_id, temp, humidity)  ║\n", numRows)
	fmt.Printf("╚══════════════════════════════════════════════════════════════════╝\n\n")

	fmt.Println("Generating synthetic sensor data...")
	d := generateData()
	fmt.Printf("Done. %.1f MB of raw float/int data in memory.\n\n", float64(numRows*32)/1024/1024)

	// ── Write ──
	fmt.Println("Writing data...")
	cemirWriteTime := cemirWrite(d)
	pqWriteTime := parquetWrite(d)

	cemirSize := cemirFileSize()
	pqSize := parquetFileSize()

	// ── Read ──
	fmt.Println("Reading data...")
	_, _, cemirTemps, cemirHums, cemirReadTime := cemirRead()
	pqRows, pqReadTime := parquetRead()

	// ── Query ──
	fmt.Printf("Running filter query (temp > %.1f °C)...\n\n", tempThreshold)
	cemirQR, cemirQueryTime := queryCemir(cemirTemps, cemirHums, tempThreshold)
	pqQR, pqQueryTime := queryParquet(pqRows, tempThreshold, )

	// ── Results ──
	fmt.Println("══════════════════════════════════════════════════════════════════")
	fmt.Printf("  %-20s  %14s  %14s  %s\n", "Metric", "CemirCol-Go", "Parquet-Go", "Winner")
	fmt.Println("──────────────────────────────────────────────────────────────────")

	printRow := func(label, cemir, pq, winner string) {
		fmt.Printf("  %-20s  %14s  %14s  %s\n", label, cemir, pq, winner)
	}

	// Write time
	wWinner := "CemirCol"
	if pqWriteTime < cemirWriteTime {
		wWinner = "Parquet"
	}
	printRow("Write Time",
		cemirWriteTime.Round(time.Millisecond).String(),
		pqWriteTime.Round(time.Millisecond).String(),
		wWinner+" "+speedup(max(cemirWriteTime, pqWriteTime), min(cemirWriteTime, pqWriteTime)))

	// Read time
	rWinner := "CemirCol"
	if pqReadTime < cemirReadTime {
		rWinner = "Parquet"
	}
	printRow("Read Time",
		cemirReadTime.Round(time.Millisecond).String(),
		pqReadTime.Round(time.Millisecond).String(),
		rWinner+" "+speedup(max(cemirReadTime, pqReadTime), min(cemirReadTime, pqReadTime)))

	// Query time
	qWinner := "CemirCol"
	if pqQueryTime < cemirQueryTime {
		qWinner = "Parquet"
	}
	printRow("Filter Query",
		cemirQueryTime.Round(time.Microsecond).String(),
		pqQueryTime.Round(time.Microsecond).String(),
		qWinner+" "+speedup(max(cemirQueryTime, pqQueryTime), min(cemirQueryTime, pqQueryTime)))

	// File size
	szWinner := "CemirCol"
	if pqSize < cemirSize {
		szWinner = "Parquet"
	}
	printRow("Total File Size",
		fmt.Sprintf("%.2f MB", cemirSize),
		fmt.Sprintf("%.2f MB", pqSize),
		szWinner)

	fmt.Println("══════════════════════════════════════════════════════════════════")

	// ── Compression ratio ──
	rawMB := float64(numRows) * (8 + 8 + 8 + 8) / 1024 / 1024 // 4 columns × 8 bytes
	fmt.Printf("\n  Raw data: %.1f MB\n", rawMB)
	fmt.Printf("  CemirCol compression: %.1f%% of raw  (%.2fx)\n",
		cemirSize/rawMB*100, rawMB/cemirSize)
	fmt.Printf("  Parquet  compression: %.1f%% of raw  (%.2fx)\n",
		pqSize/rawMB*100, rawMB/pqSize)

	// ── Visual bar chart ──
	maxTime := max(cemirWriteTime+cemirReadTime, pqWriteTime+pqReadTime)
	cemirTotal := cemirWriteTime + cemirReadTime
	pqTotal := pqWriteTime + pqReadTime
	const barWidth = 30

	fmt.Printf("\n  Write + Read total time (bar chart):\n")
	fmt.Printf("  CemirCol %s %v\n",
		bar(float64(cemirTotal)/float64(maxTime), barWidth), cemirTotal.Round(time.Millisecond))
	fmt.Printf("  Parquet  %s %v\n",
		bar(float64(pqTotal)/float64(maxTime), barWidth), pqTotal.Round(time.Millisecond))

	// ── Query results validation ──
	fmt.Printf("\n  Filter query (temp > %.1f °C) results:\n", tempThreshold)
	fmt.Printf("  %-12s  matches=%d  avg_temp=%.2f°C  avg_hum=%.2f%%\n",
		"CemirCol:", cemirQR.HotCount, cemirQR.AvgTemp, cemirQR.AvgHumidity)
	fmt.Printf("  %-12s  matches=%d  avg_temp=%.2f°C  avg_hum=%.2f%%\n",
		"Parquet:", pqQR.HotCount, pqQR.AvgTemp, pqQR.AvgHumidity)

	if cemirQR.HotCount == pqQR.HotCount {
		fmt.Println("\n  Results match — both engines return identical data.")
	} else {
		fmt.Println("\n  WARNING: result mismatch between engines!")
	}

	// ── Cleanup ──
	cemirCleanup()
	os.Remove("sensor.parquet")
	fmt.Println()
}

func max(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
