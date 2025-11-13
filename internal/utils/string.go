package utils

import (
	"fmt"
	"math"
	"time"
)

// StrFileSize 将字节大小转换为可读性更好的文件大小字符�?func StrFileSize(size float64) string {
	if size == 0 {
		return "0B"
	}
	
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	unitIndex := 0
	
	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}
	
	// 如果是整数，不显示小数点
	if math.Floor(size) == size {
		return fmt.Sprintf("%.0f%s", size, units[unitIndex])
	}
	
	return fmt.Sprintf("%.1f%s", size, units[unitIndex])
}

// FileSizeToString 将字节大小转换为可读性更好的文件大小字符�?func FileSizeToString(size float64) string {
	return StrFileSize(size)
}

// SecondsToString 将秒数转换为可读性更好的时间字符�?func SecondsToString(seconds int64) string {
	if seconds <= 0 {
		return "0�?
	}
	
	// 转换为时�?	duration := time.Duration(seconds) * time.Second
	
	// 格式化时�?	if duration.Hours() >= 24 {
		return fmt.Sprintf("%d�?, int(duration.Hours()/24))
	} else if duration.Hours() >= 1 {
		return fmt.Sprintf("%d小时", int(duration.Hours()))
	} else if duration.Minutes() >= 1 {
		return fmt.Sprintf("%d分钟", int(duration.Minutes()))
	} else {
		return fmt.Sprintf("%d�?, int(duration.Seconds()))
	}
}
