package logrotate

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkRotatorWrite(b *testing.B) {
	r := &Rotator{
		Filename:   filepath.Join("bench.log"),
		MaxSize:    5,
		MaxBackups: 7,
	}
	defer func() {
		r.Close()
		os.Remove("bench.log")
	}()

	data := []byte("2026-05-23 20:00:00.000 info  this is a benchmark log line for testing rotator performance\n")

	b.ResetTimer()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		r.Write(data)
	}
}

func BenchmarkRotatorWriteWithRotation(b *testing.B) {
	r := &Rotator{
		Filename:   filepath.Join("bench.log"),
		MaxSize:    1, // 1MB to trigger rotation frequently
		MaxBackups: 3,
	}
	defer func() {
		r.Close()
		os.Remove("bench.log")
		matches, _ := filepath.Glob(filepath.Join("bench_*"))
		for _, f := range matches {
			os.Remove(f)
		}
	}()

	data := []byte("2026-05-23 20:00:00.000 info  this is a benchmark log line for testing rotator performance\n")

	b.ResetTimer()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		r.Write(data)
	}
}
