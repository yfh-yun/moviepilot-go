package download

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"moviepilot-go/pkg/logger"
)

// LimiterService 速度限制服务
type LimiterService interface {
	// SetGlobalLimit 设置全局速度限制
	SetGlobalLimit(ctx context.Context, downloadLimit, uploadLimit int64) error

	// SetTaskLimit 设置任务速度限制
	SetTaskLimit(ctx context.Context, taskID string, downloadLimit, uploadLimit int64) error

	// GetGlobalLimit 获取全局速度限制
	GetGlobalLimit(ctx context.Context) (*SpeedLimit, error)

	// GetTaskLimit 获取任务速度限制
	GetTaskLimit(ctx context.Context, taskID string) (*SpeedLimit, error)

	// SetSchedule 设置定时限速
	SetSchedule(ctx context.Context, schedule *LimitSchedule) error

	// GetSchedules 获取所有定时限速
	GetSchedules(ctx context.Context) ([]*LimitSchedule, error)

	// ApplyCurrentSchedule 应用当前时间段的限速
	ApplyCurrentSchedule(ctx context.Context) error
}

// limiterService 限速服务实现
type limiterService struct {
	globalLimit *SpeedLimit
	taskLimits  map[string]*SpeedLimit
	schedules   []*LimitSchedule
	mutex       sync.RWMutex
	logger      *zap.Logger
}

// NewLimiterService 创建限速服务
func NewLimiterService() LimiterService {
	return &limiterService{
		globalLimit: &SpeedLimit{
			DownloadLimit: 0, // 0 表示不限速
			UploadLimit:   0,
		},
		taskLimits: make(map[string]*SpeedLimit),
		schedules:  make([]*LimitSchedule, 0),
		logger:     logger.GetLogger(),
	}
}

// SpeedLimit 速度限制
type SpeedLimit struct {
	DownloadLimit int64 `json:"download_limit"` // 字节/秒，0 表示不限速
	UploadLimit   int64 `json:"upload_limit"`   // 字节/秒，0 表示不限速
}

// LimitSchedule 定时限速
type LimitSchedule struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Enabled       bool   `json:"enabled"`
	StartTime     string `json:"start_time"` // HH:MM 格式
	EndTime       string `json:"end_time"`   // HH:MM 格式
	WeekDays      []int  `json:"week_days"`  // 0-6，0表示周日
	DownloadLimit int64  `json:"download_limit"`
	UploadLimit   int64  `json:"upload_limit"`
}

// SetGlobalLimit 设置全局速度限制
func (s *limiterService) SetGlobalLimit(ctx context.Context, downloadLimit, uploadLimit int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Info("设置全局速度限制",
		zap.Int64("download_limit", downloadLimit),
		zap.Int64("upload_limit", uploadLimit),
	)

	s.globalLimit = &SpeedLimit{
		DownloadLimit: downloadLimit,
		UploadLimit:   uploadLimit,
	}

	return nil
}

// SetTaskLimit 设置任务速度限制
func (s *limiterService) SetTaskLimit(ctx context.Context, taskID string, downloadLimit, uploadLimit int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Info("设置任务速度限制",
		zap.String("task_id", taskID),
		zap.Int64("download_limit", downloadLimit),
		zap.Int64("upload_limit", uploadLimit),
	)

	s.taskLimits[taskID] = &SpeedLimit{
		DownloadLimit: downloadLimit,
		UploadLimit:   uploadLimit,
	}

	return nil
}

// GetGlobalLimit 获取全局速度限制
func (s *limiterService) GetGlobalLimit(ctx context.Context) (*SpeedLimit, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.globalLimit, nil
}

// GetTaskLimit 获取任务速度限制
func (s *limiterService) GetTaskLimit(ctx context.Context, taskID string) (*SpeedLimit, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	limit, ok := s.taskLimits[taskID]
	if !ok {
		return nil, nil
	}

	return limit, nil
}

// SetSchedule 设置定时限速
func (s *limiterService) SetSchedule(ctx context.Context, schedule *LimitSchedule) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Info("设置定时限速",
		zap.String("name", schedule.Name),
		zap.String("start_time", schedule.StartTime),
		zap.String("end_time", schedule.EndTime),
	)

	// 检查是否已存在
	found := false
	for i, sch := range s.schedules {
		if sch.ID == schedule.ID {
			s.schedules[i] = schedule
			found = true
			break
		}
	}

	if !found {
		s.schedules = append(s.schedules, schedule)
	}

	return nil
}

// GetSchedules 获取所有定时限速
func (s *limiterService) GetSchedules(ctx context.Context) ([]*LimitSchedule, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.schedules, nil
}

// ApplyCurrentSchedule 应用当前时间段的限速
func (s *limiterService) ApplyCurrentSchedule(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	currentTime := now.Format("15:04")
	currentWeekDay := int(now.Weekday())

	// 查找匹配的定时限速
	for _, schedule := range s.schedules {
		if !schedule.Enabled {
			continue
		}

		// 检查星期
		weekDayMatch := false
		if len(schedule.WeekDays) == 0 {
			weekDayMatch = true // 未指定星期，表示每天
		} else {
			for _, day := range schedule.WeekDays {
				if day == currentWeekDay {
					weekDayMatch = true
					break
				}
			}
		}

		if !weekDayMatch {
			continue
		}

		// 检查时间范围
		if currentTime >= schedule.StartTime && currentTime <= schedule.EndTime {
			s.logger.Info("应用定时限速",
				zap.String("schedule", schedule.Name),
				zap.Int64("download_limit", schedule.DownloadLimit),
				zap.Int64("upload_limit", schedule.UploadLimit),
			)

			s.globalLimit = &SpeedLimit{
				DownloadLimit: schedule.DownloadLimit,
				UploadLimit:   schedule.UploadLimit,
			}

			return nil
		}
	}

	return nil
}

// FormatSpeed 格式化速度
func FormatSpeed(bytesPerSecond int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	if bytesPerSecond >= GB {
		return fmt.Sprintf("%.2f GB/s", float64(bytesPerSecond)/float64(GB))
	} else if bytesPerSecond >= MB {
		return fmt.Sprintf("%.2f MB/s", float64(bytesPerSecond)/float64(MB))
	} else if bytesPerSecond >= KB {
		return fmt.Sprintf("%.2f KB/s", float64(bytesPerSecond)/float64(KB))
	}
	return fmt.Sprintf("%d B/s", bytesPerSecond)
}

// ParseSpeed 解析速度字符串
func ParseSpeed(speedStr string) (int64, error) {
	if len(speedStr) == 0 {
		return 0, nil
	}

	// 简单实现：提取数字部分和单位
	var value float64
	var unit string

	// 使用fmt.Sscanf解析
	n, err := fmt.Sscanf(speedStr, "%f%s", &value, &unit)
	if err != nil || n < 2 {
		return 0, fmt.Errorf("invalid speed format: %s", speedStr)
	}

	unit = strings.ToLower(strings.TrimSpace(unit))

	switch unit {
	case "b", "b/s":
		return int64(value), nil
	case "kb", "kb/s":
		return int64(value * 1024), nil
	case "mb", "mb/s":
		return int64(value * 1024 * 1024), nil
	case "gb", "gb/s":
		return int64(value * 1024 * 1024 * 1024), nil
	case "tb", "tb/s":
		return int64(value * 1024 * 1024 * 1024 * 1024), nil
	default:
		return 0, fmt.Errorf("unsupported speed unit: %s", unit)
	}
}
