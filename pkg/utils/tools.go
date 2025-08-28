package utils

import (
	"fmt"
	"strconv"
)

// FormatCpuCapacity 格式化CPU容量为易读格式
func FormatCpuCapacity(cpuHz int64) string {
	cpuGHz := float64(cpuHz) / 1000000000
	return strconv.FormatFloat(cpuGHz, 'f', 2, 64) + " GHz"
}

// FormatMemorySize 格式化内存大小为易读格式
func FormatMemorySize(sizeInBytes int64) string {
	if sizeInBytes < 1024 {
		return fmt.Sprintf("%d B", sizeInBytes)
	} else if sizeInBytes < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(sizeInBytes)/1024)
	} else if sizeInBytes < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(sizeInBytes)/(1024*1024))
	} else if sizeInBytes < 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2f GB", float64(sizeInBytes)/(1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2f TB", float64(sizeInBytes)/(1024*1024*1024*1024))
	}
}

// FormatDiskSize 格式化磁盘大小为易读格式
func FormatDiskSize(sizeInBytes int64) string {
	return FormatMemorySize(sizeInBytes) // 可以复用内存格式化函数
}
