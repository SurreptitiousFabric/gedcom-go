package decoder

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

const (
	bench10MBEnv        = "GEDCOM_BENCH_10MB"
	bench100MBEnv       = "GEDCOM_BENCH_100MB"
	benchAssertEnv      = "GEDCOM_BENCH_ASSERT"
	bench10MBMinSize    = 10 * 1024 * 1024
	bench100MBMinSize   = 100 * 1024 * 1024
	bench10MBMaxElapsed = 2 * time.Second
	bench100MBMaxHeap   = 350 * 1024 * 1024
)

// BenchmarkDecodeMinimal benchmarks parsing a minimal GEDCOM file (~170 bytes)
func BenchmarkDecodeMinimal(b *testing.B) {
	f, err := os.Open("../testdata/gedcom-5.5/minimal.ged")
	if err != nil {
		b.Skip("Test file not found:", err)
	}
	defer f.Close()

	// Read file into memory once
	data, err := os.ReadFile("../testdata/gedcom-5.5/minimal.ged")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := Decode(newBytesReader(data))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDecodeSmall benchmarks parsing a small GEDCOM file (~15KB, GEDCOM 7.0 maximal)
func BenchmarkDecodeSmall(b *testing.B) {
	data, err := os.ReadFile("../testdata/gedcom-7.0/maximal70.ged")
	if err != nil {
		b.Skip("Test file not found:", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := Decode(newBytesReader(data))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDecodeMedium benchmarks parsing a medium GEDCOM file (~458KB, British Royal Family)
func BenchmarkDecodeMedium(b *testing.B) {
	data, err := os.ReadFile("../testdata/gedcom-5.5/royal92.ged")
	if err != nil {
		b.Skip("Test file not found:", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := Decode(newBytesReader(data))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDecodeLarge benchmarks parsing a large GEDCOM file (~1.1MB, US Presidents)
func BenchmarkDecodeLarge(b *testing.B) {
	data, err := os.ReadFile("../testdata/gedcom-5.5/pres2020.ged")
	if err != nil {
		b.Skip("Test file not found:", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := Decode(newBytesReader(data))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDecode10MB benchmarks parsing a GEDCOM file ~10MB (set GEDCOM_BENCH_10MB to override).
func BenchmarkDecode10MB(b *testing.B) {
	data := readBenchmarkGED(b, bench10MBEnv, bench10MBMinSize)

	if os.Getenv(benchAssertEnv) != "" {
		start := time.Now()
		if _, err := Decode(newBytesReader(data)); err != nil {
			b.Fatal(err)
		}
		if elapsed := time.Since(start); elapsed > bench10MBMaxElapsed {
			b.Fatalf("decode exceeded %s: %s", bench10MBMaxElapsed, elapsed)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))

	for i := 0; i < b.N; i++ {
		_, err := Decode(newBytesReader(data))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDecode100MBMemory benchmarks memory usage for ~100MB file parsing (set GEDCOM_BENCH_100MB to override).
func BenchmarkDecode100MBMemory(b *testing.B) {
	data := readBenchmarkGED(b, bench100MBEnv, bench100MBMinSize)

	heapBytes, err := decodeHeapBytes(data)
	if err != nil {
		b.Fatal(err)
	}
	// heapBytes excludes the input buffer; it reflects decoder allocations only.
	heapMB := float64(heapBytes) / (1024 * 1024)

	if os.Getenv(benchAssertEnv) != "" && heapBytes > bench100MBMaxHeap {
		b.Fatalf("decode heap usage exceeded %d bytes: %d bytes", bench100MBMaxHeap, heapBytes)
	}

	b.ResetTimer()
	b.ReportMetric(heapMB, "heap_mb")
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))

	for i := 0; i < b.N; i++ {
		_, err := Decode(newBytesReader(data))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper to create a fresh bytes.Reader for each iteration
func newBytesReader(data []byte) io.Reader {
	return bytes.NewReader(data)
}

func readBenchmarkGED(b *testing.B, envVar string, minSize int64) []byte {
	b.Helper()

	path, err := benchmarkFilePath(envVar, minSize)
	if err != nil {
		b.Skip(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatal(err)
	}
	if int64(len(data)) < minSize {
		b.Skipf("benchmark file %q is %d bytes; need at least %d bytes", path, len(data), minSize)
	}

	return data
}

func benchmarkFilePath(envVar string, minSize int64) (string, error) {
	if path := os.Getenv(envVar); path != "" {
		info, err := os.Stat(path)
		if err != nil {
			return "", err
		}
		if info.Size() < minSize {
			return "", fmt.Errorf("%s is %d bytes; need at least %d bytes", path, info.Size(), minSize)
		}
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	downloads := filepath.Join(home, "Downloads")
	entries, err := os.ReadDir(downloads)
	if err != nil {
		return "", err
	}

	var candidates []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".ged") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.Size() < minSize {
			continue
		}
		candidates = append(candidates, filepath.Join(downloads, name))
	}

	sort.Strings(candidates)
	switch len(candidates) {
	case 0:
		return "", fmt.Errorf("no .ged files >= %d bytes in %s (set %s)", minSize, downloads, envVar)
	case 1:
		return candidates[0], nil
	default:
		return "", fmt.Errorf("multiple .ged files >= %d bytes in %s (set %s to choose)", minSize, downloads, envVar)
	}
}

func decodeHeapBytes(data []byte) (int64, error) {
	runtime.GC()

	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	doc, err := Decode(newBytesReader(data))
	if err != nil {
		return 0, err
	}

	runtime.ReadMemStats(&after)
	runtime.KeepAlive(doc)

	delta := int64(after.HeapAlloc) - int64(before.HeapAlloc)
	if delta < 0 {
		delta = 0
	}

	return delta, nil
}
