package schedulers

import (
	"github.com/robfig/cron/v3"
)

// CronParser Cron表达式解析器
type CronParser struct {
	parser cron.Parser
}

// NewCronParser 创建Cron解析器
func NewCronParser() *CronParser {
	return &CronParser{
		parser: cron.NewParser(
			cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		),
	}
}

// Parse 解析Cron表达式
func (p *CronParser) Parse(spec string) (cron.Schedule, error) {
	return p.parser.Parse(spec)
}

// IsValid 检查Cron表达式是否有效
func (p *CronParser) IsValid(spec string) bool {
	_, err := p.Parse(spec)
	return err == nil
}
