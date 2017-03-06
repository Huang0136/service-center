package utils

import (
	"fmt"
	"testing"
	"time"
)

// 生成uuid测试
func TestCreateUUID(t *testing.T) {
	count := 100000
	tBegin := time.Now()
	for i := 1; i <= count; i++ {
		fmt.Println("creat uuid:", i, ",", CreateUUID())
	}
	tEnd := time.Now()

	t.Logf("生成uuid:%d次，耗时:%d", count, tEnd.Sub(tBegin).Nanoseconds()/int64(count))
}
