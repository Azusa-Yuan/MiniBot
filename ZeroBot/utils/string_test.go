package utils

import (
	"strings"
	"testing"
)

var test = strings.Repeat("基准测试  测测性能", 40)

func BenchmarkString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		data := []byte(test)
		test = string(data)
	}
}

func BenchmarkStringUtils(b *testing.B) {
	for i := 0; i < b.N; i++ {
		data := StringToBytes(test)
		test = BytesToString(data)
	}
}
