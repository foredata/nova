package strx

import "fmt"

// Sprintf 类似fmt.Sprintf，但如果没有参数，则不会执行format计算
func Sprintf(format string, args ...interface{}) string {
	if len(args) > 0 {
		return fmt.Sprintf(format, args...)
	}

	return format
}
