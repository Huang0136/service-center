package utils

import (
	"testing"
	"fmt"
)

func TestCreateUUID(t *testing.T) {
	uuid := CreateUUID()
	if uuid == "" {
		t.Error("error uuid")
	}
}

func BenchmarkCreateUUID(b *testing.B) {
	for i := 1; i < b.N; i++ {
		str := CreateUUID()
		fmt.Println(str)
	}
}
