package utils

import (
	"github.com/satori/go.uuid"
)

// 生成UUID
func CreateUUID() string {
	return uuid.NewV1().String()
}
