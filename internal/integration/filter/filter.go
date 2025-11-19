// Package filter 过滤模块
// 提供媒体过滤规则和解析功能
package filter

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/yfh-yun/moviepilot-go/pkg/logger"
	"github.com/yfh-yun/moviepilot-go/internal/models"

	"go.uber.org/zap"
)

// ModuleType 模块类型
type ModuleType string

const (
	ModuleTypeOther ModuleType = "Other"
)

// OtherModulesType 其他模块类型
type OtherModulesType string

const (
	OtherModulesTypeFilter OtherModulesType = "Filter"
)

// FilterModule 过滤模块
type FilterModule struct {
	logger     *zap.Logger
	ruleParser *RuleParser
	ruleSet    map[string]*FilterRule
	mutex      sync.RWMutex
}

// FilterRule 过滤规则
type FilterRule struct {
	Name     string            `json:"name"`
	Include  []string          `json:"include"`
	Exclude  []string          `json:"exclude"`
	Match    []string          `json:"match"`
	TMDB     map[string]string `json:"tmdb"`
	patterns map[string]*regexp.Regexp
}

// FilterResult 过滤结果
type FilterResult struct {
	Matched   bool     `json:"matched"`
	Rules     []string `json:"rules"`
	Reason    string   `json:"reason"`
	Score     int      `json:"score"`
	Priority  int      `json:"priority"`
}

// FilterRequest 过滤请求
type FilterRequest struct {
	TorrentInfo *models.TorrentInfo `json:"torrent_info"`
	MediaInfo   *models.MediaInfo    `json:"media_info"`
	MetaInfo    *models.MetaBase     `json:"meta_info"`
	Rules       []string             `json:"rules"`
}

// NewFilterModule 创建过滤模块
func NewFilterModule() *FilterModule {
	fm := &FilterModule{
		logger:     logger.Logger,
		ruleParser: NewRuleParser(),
		ruleSet:    make(map[string]*FilterRule),
		mutex:      sync.RWMutex{},
	}

	// 初始化内置规则集
	fm.initBuiltinRules()

	return fm
}

// InitModule 初始化模块
func (fm *FilterModule) InitModule(ctx context.Context) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	// 编译所有规则的正则表达式
	for name, rule := range fm.ruleSet {
		if err := fm.compileRule(rule); err != nil {
			return fmt.Errorf("failed to compile rule %s: %w", name, err)
		}
	}

	fm.logger.Info("Filter module initialized",
		zap.Int("rules_count", len(fm.ruleSet)))

	return nil
}

// GetName 获取模块名称
func (fm *FilterModule) GetName() string {
	return "过滤规则"
}

// GetType 获取模块类型
func (fm *FilterModule) GetType() ModuleType {
	return ModuleTypeOther
}

// GetSubType 获取模块子类型
func (fm *FilterModule) GetSubType() OtherModulesType {
	return OtherModulesTypeFilter
}

// GetPriority 获取模块优先级
func (fm *FilterModule) GetPriority() int {
	return 3
}

// Stop 停止模块
func (fm *FilterModule) Stop(ctx context.Context) error {
	fm.logger.Info("Filter module stopped")
	return nil
}

// Test 测试模块
func (fm *FilterModule) Test(ctx context.Context) error {
	// 测试规则解析器
	testExpression := "BLU & 4K & !REMUX"
	if _, err := fm.ruleParser.Parse(testExpression); err != nil {
		return fmt.Errorf("rule parser test failed: %w", err)
	}

	// 测试内置规则
	testTorrent := &models.TorrentInfo{
		Title: "Movie.2023.2160p.BluRay.HEVC.DTS-HD.MA.5.1",
	}

	result := fm.FilterTorrent(context.Background(), &FilterRequest{
		TorrentInfo: testTorrent,
		Rules:       []string{"4K", "BLU"},
	})

	if !result.Matched {
		return fmt.Errorf("builtin rule test failed")
	}

	return nil
}

// FilterTorrent 过滤种子
func (fm *FilterModule) FilterTorrent(ctx context.Context, req *FilterRequest) *FilterResult {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	result := &FilterResult{
		Matched: false,
		Rules:   make([]string, 0),
		Score:   0,
		Priority: 999,
	}

	// 如果没有指定规则，返回不匹配
	if len(req.Rules) == 0 {
		result.Reason = "no rules specified"
		return result
	}

	// 解析规则表达式
	expression := strings.Join(req.Rules, " & ")
	parsedExpr, err := fm.ruleParser.Parse(expression)
	if err != nil {
		result.Reason = fmt.Sprintf("failed to parse expression: %s", err.Error())
		return result
	}

	// 评估表达式
	matched, err := fm.evaluateExpression(ctx, parsedExpr, req)
	if err != nil {
		result.Reason = fmt.Sprintf("failed to evaluate expression: %s", err.Error())
		return result
	}

	result.Matched = matched
	result.Rules = req.Rules

	if matched {
		result.Reason = "all rules matched"
		result.Score = fm.calculateScore(req.Rules)
		result.Priority = fm.calculatePriority(req.Rules)
	} else {
		result.Reason = "some rules not matched"
	}

	return result
}

// FilterMedia 过滤媒体
func (fm *FilterModule) FilterMedia(ctx context.Context, mediaInfo *models.MediaInfo, rules []string) *FilterResult {
	req := &FilterRequest{
		MediaInfo: mediaInfo,
		Rules:     rules,
	}

	return fm.FilterTorrent(ctx, req)
}

// AddRule 添加规则
func (fm *FilterModule) AddRule(name string, rule *FilterRule) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if err := fm.compileRule(rule); err != nil {
		return fmt.Errorf("failed to compile rule: %w", err)
	}

	fm.ruleSet[name] = rule
	fm.logger.Info("Rule added", zap.String("name", name))

	return nil
}

// RemoveRule 移除规则
func (fm *FilterModule) RemoveRule(name string) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	delete(fm.ruleSet, name)
	fm.logger.Info("Rule removed", zap.String("name", name))
}

// GetRule 获取规则
func (fm *FilterModule) GetRule(name string) (*FilterRule, bool) {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	rule, exists := fm.ruleSet[name]
	return rule, exists
}

// ListRules 列出所有规则
func (fm *FilterModule) ListRules() map[string]*FilterRule {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	result := make(map[string]*FilterRule)
	for name, rule := range fm.ruleSet {
		// 复制规则，避免外部修改
		ruleCopy := *rule
		result[name] = &ruleCopy
	}

	return result
}

// compileRule 编译规则
func (fm *FilterModule) compileRule(rule *FilterRule) error {
	rule.patterns = make(map[string]*regexp.Regexp)

	// 编译include规则
	for i, pattern := range rule.Include {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("failed to compile include pattern %d: %w", i, err)
		}
		rule.patterns["include_"+fmt.Sprint(i)] = regex
	}

	// 编译exclude规则
	for i, pattern := range rule.Exclude {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("failed to compile exclude pattern %d: %w", i, err)
		}
		rule.patterns["exclude_"+fmt.Sprint(i)] = regex
	}

	return nil
}

// evaluateExpression 评估表达式
func (fm *FilterModule) evaluateExpression(ctx context.Context, expr *ParsedExpression, req *FilterRequest) (bool, error) {
	switch expr.Type {
	case ExprTypeAtom:
		return fm.evaluateAtom(ctx, expr.Value, req)
	case ExprTypeNot:
		child, ok := expr.Children[0].(*ParsedExpression)
		if !ok {
			return false, fmt.Errorf("invalid NOT expression")
		}
		result, err := fm.evaluateExpression(ctx, child, req)
		return !result, err
	case ExprTypeAnd:
		for _, child := range expr.Children {
			childExpr, ok := child.(*ParsedExpression)
			if !ok {
				return false, fmt.Errorf("invalid AND expression")
			}
			result, err := fm.evaluateExpression(ctx, childExpr, req)
			if err != nil {
				return false, err
			}
			if !result {
				return false, nil
			}
		}
		return true, nil
	case ExprTypeOr:
		for _, child := range expr.Children {
			childExpr, ok := child.(*ParsedExpression)
			if !ok {
				return false, fmt.Errorf("invalid OR expression")
			}
			result, err := fm.evaluateExpression(ctx, childExpr, req)
			if err != nil {
				return false, err
			}
			if result {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("unknown expression type: %s", expr.Type)
	}
}

// evaluateAtom 评估原子表达式
func (fm *FilterModule) evaluateAtom(ctx context.Context, atom string, req *FilterRequest) (bool, error) {
	// 检查是否为内置规则
	rule, exists := fm.GetRule(atom)
	if !exists {
		return false, fmt.Errorf("unknown rule: %s", atom)
	}

	// 检查include规则
	for _, pattern := range rule.patterns {
		if strings.HasPrefix(pattern.String(), "include_") {
			if req.TorrentInfo != nil && pattern.MatchString(req.TorrentInfo.Title) {
				return true, nil
			}
			if req.MediaInfo != nil && pattern.MatchString(req.MediaInfo.Title) {
				return true, nil
			}
		}
	}

	// 检查exclude规则
	for _, pattern := range rule.patterns {
		if strings.HasPrefix(pattern.String(), "exclude_") {
			if req.TorrentInfo != nil && pattern.MatchString(req.TorrentInfo.Title) {
				return false, nil
			}
			if req.MediaInfo != nil && pattern.MatchString(req.MediaInfo.Title) {
				return false, nil
			}
		}
	}

	// 检查TMDB规则
	if rule.TMDB != nil && req.MediaInfo != nil {
		for key, value := range rule.TMDB {
			switch key {
			case "original_language":
				if req.MediaInfo.OriginalLanguage == value {
					return true, nil
				}
			}
		}
	}

	// 检查match规则
	for _, matchType := range rule.Match {
		switch matchType {
		case "labels":
			if req.TorrentInfo != nil && len(req.TorrentInfo.Labels) > 0 {
				return true, nil
			}
		}
	}

	return false, nil
}

// calculateScore 计算分数
func (fm *FilterModule) calculateScore(rules []string) int {
	score := 0
	ruleScores := map[string]int{
		"BLU":    100,
		"4K":     90,
		"1080P":  80,
		"720P":   70,
		"H265":   60,
		"H264":   50,
		"DOLBY":  40,
		"ATMOS":  30,
		"HDR":    20,
		"CNSUB":  10,
		"GZ":     100,
	}

	for _, rule := range rules {
		if score, exists := ruleScores[strings.ToUpper(rule)]; exists {
			score += score
		}
	}

	return score
}

// calculatePriority 计算优先级
func (fm *FilterModule) calculatePriority(rules []string) int {
	priority := 999
	rulePriorities := map[string]int{
		"BLU":    1,
		"4K":     2,
		"1080P":  3,
		"720P":   4,
		"H265":   5,
		"H264":   6,
		"DOLBY":  7,
		"ATMOS":  8,
		"HDR":    9,
		"CNSUB":  10,
		"GZ":     1,
	}

	for _, rule := range rules {
		if priority, exists := rulePriorities[strings.ToUpper(rule)]; exists && priority < priority {
			priority = priority
		}
	}

	return priority
}

// initBuiltinRules 初始化内置规则
func (fm *FilterModule) initBuiltinRules() {
	// 蓝光原盘
	fm.ruleSet["BLU"] = &FilterRule{
		Name: "BLU",
		Include: []string{
			`(?i)(\bBlu-?Ray\b.*\b(?:VC-?1|AVC|MPEG-?2)\b|\b(?:UHD|4K|2160p)\b(?:.*Blu-?Ray)?.*\b(?:HEVC|H\.?265)\b|\bBlu-?Ray\b.*\b(?:UHD|4K|2160p)\b.*\b(?:HEVC|H\.?265)\b|\b(?:COMPLETE|FULL)\b.*\b(?:(?:UHD|4K|2160p)\b.*)?Blu-?Ray\b|\b(BD25|BD50|BD66|BD100|BDMV|MiniBD)\b)`,
		},
		Exclude: []string{
			`(?i)(\b[XH]\.?264\b|\b[XH]\.?265\b|\bWEB-?DL\b|\bWEB-?RIP\b|\bHDTV(?:RIP)?\b|\bREMUX\b|\bBDRip\b|\bBRRip\b|\bHDRip\b|\bENCODE\b|\b(?<!WEB-|HDTV)RIP\b)`,
		},
	}

	// 4K
	fm.ruleSet["4K"] = &FilterRule{
		Name:    "4K",
		Include: []string{`4k|2160p|x2160`},
		Exclude: []string{},
	}

	// 1080P
	fm.ruleSet["1080P"] = &FilterRule{
		Name:    "1080P",
		Include: []string{`1080[pi]|x1080`},
		Exclude: []string{},
	}

	// 720P
	fm.ruleSet["720P"] = &FilterRule{
		Name:    "720P",
		Include: []string{`720[pi]|x720`},
		Exclude: []string{},
	}

	// 中字
	fm.ruleSet["CNSUB"] = &FilterRule{
		Name: "CNSUB",
		Include: []string{
			`[中国國繁简](/|\s|\\|\|)?[繁简英粤]|[英简繁](/|\s|\\|\|)?[中繁简]|繁體|简体|[中国國][字配]|国语|國語|中文|中字|简日|繁日|简繁|繁体|([\s,.-\[])(CHT|CHS|cht|chs)(|[\s,.-\]])`,
		},
		Exclude: []string{},
		TMDB: map[string]string{
			"original_language": "zh,cn",
		},
	}

	// 官种
	fm.ruleSet["GZ"] = &FilterRule{
		Name:    "GZ",
		Include: []string{`官方`, `官种`, `官组`},
		Match:   []string{"labels"},
	}

	// 特效字幕
	fm.ruleSet["SPECSUB"] = &FilterRule{
		Name:    "SPECSUB",
		Include: []string{`特效`},
		Exclude: []string{},
	}

	// BluRay
	fm.ruleSet["BLURAY"] = &FilterRule{
		Name:    "BLURAY",
		Include: []string{`Blu-?Ray`},
		Exclude: []string{},
	}

	// UHD
	fm.ruleSet["UHD"] = &FilterRule{
		Name:    "UHD",
		Include: []string{`UHD|UltraHD`},
		Exclude: []string{},
	}

	// H265
	fm.ruleSet["H265"] = &FilterRule{
		Name:    "H265",
		Include: []string{`[Hx].?265|HEVC`},
		Exclude: []string{},
	}

	// H264
	fm.ruleSet["H264"] = &FilterRule{
		Name:    "H264",
		Include: []string{`[Hx].?264|AVC`},
		Exclude: []string{},
	}

	// 杜比视界
	fm.ruleSet["DOLBY"] = &FilterRule{
		Name:    "DOLBY",
		Include: []string{`Dolby[\s.]+Vision|DOVI|[\s.]+DV[\s.]+|杜比视界`},
		Exclude: []string{},
	}

	// 杜比全景声
	fm.ruleSet["ATMOS"] = &FilterRule{
		Name:    "ATMOS",
		Include: []string{`Dolby[\s.+]+Atmos|Atmos|杜比全景[声聲]`},
		Exclude: []string{},
	}

	// HDR
	fm.ruleSet["HDR"] = &FilterRule{
		Name:    "HDR",
		Include: []string{`[\s.]+HDR[\s.]+|HDR10|HDR10\+`},
		Exclude: []string{},
	}

	// SDR
	fm.ruleSet["SDR"] = &FilterRule{
		Name:    "SDR",
		Include: []string{`[\s.]+SDR[\s.]+`},
		Exclude: []string{},
	}
}