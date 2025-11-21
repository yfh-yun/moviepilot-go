// Package parsers 解析器工厂
package parsers

import (
	"fmt"
	"strings"

	"moviepilot-go/internal/integration/indexer"
)

// ParserFactory 解析器工厂
type ParserFactory struct {
	parsers map[string]indexer.SiteParser
}

// NewParserFactory 创建解析器工厂
func NewParserFactory() *ParserFactory {
	factory := &ParserFactory{
		parsers: make(map[string]indexer.SiteParser),
	}

	// 注册所有解析器
	factory.registerParsers()

	return factory
}

// registerParsers 注册所有解析器
func (f *ParserFactory) registerParsers() {
	// 注册已知解析器
	f.RegisterParser("nexusphp", NewNexusPHPParser())
	f.RegisterParser("gazelle", NewGazelleParser())
	f.RegisterParser("discuz", NewDiscuzParser())
	f.RegisterParser("unit3d", NewUnit3DParser())
	f.RegisterParser("tnode", NewTNodeParser())
	f.RegisterParser("filelist", NewFileListParser())
	f.RegisterParser("smallhorse", NewSmallHorseParser())
	f.RegisterParser("yema", NewYemaParser())
	f.RegisterParser("hddolby", NewHDDolbyParser())

	// 注册Nexus系列解析器
	f.RegisterParser("nexus_audiences", NewNexusPHPParser()) // 基于NexusPHP
	f.RegisterParser("nexus_hhanclub", NewNexusPHPParser())     // 基于NexusPHP
	f.RegisterParser("nexus_project", NewNexusPHPParser())      // 基于NexusPHP
	f.RegisterParser("nexus_rabbit", NewNexusPHPParser())       // 基于NexusPHP

	// 注册其他解析器
	f.RegisterParser("ipt_project", NewNexusPHPParser()) // 基于NexusPHP
	f.RegisterParser("mtorrent", NewNexusPHPParser())   // 基于NexusPHP
	f.RegisterParser("torrent_leech", NewNexusPHPParser()) // 基于NexusPHP
}

// RegisterParser 注册解析器
func (f *ParserFactory) RegisterParser(schema string, parser indexer.SiteParser) {
	f.parsers[strings.ToLower(schema)] = parser
}

// GetParser 获取解析器
func (f *ParserFactory) GetParser(schema string) (indexer.SiteParser, error) {
	schema = strings.ToLower(schema)
	parser, exists := f.parsers[schema]
	if !exists {
		return nil, fmt.Errorf("unsupported site schema: %s", schema)
	}
	return parser, nil
}

// GetSupportedSchemas 获取支持的站点模式
func (f *ParserFactory) GetSupportedSchemas() []string {
	var schemas []string
	for schema := range f.parsers {
		schemas = append(schemas, schema)
	}
	return schemas
}

// HasParser 检查是否支持指定模式
func (f *ParserFactory) HasParser(schema string) bool {
	schema = strings.ToLower(schema)
	_, exists := f.parsers[schema]
	return exists
}

// GetAllParsers 获取所有解析器
func (f *ParserFactory) GetAllParsers() map[string]indexer.SiteParser {
	result := make(map[string]indexer.SiteParser)
	for schema, parser := range f.parsers {
		result[schema] = parser
	}
	return result
}

// 全局解析器工厂实例
var DefaultParserFactory = NewParserFactory()

// GetParser 获取解析器（全局函数）
func GetParser(schema string) (indexer.SiteParser, error) {
	return DefaultParserFactory.GetParser(schema)
}

// HasParser 检查是否支持指定模式（全局函数）
func HasParser(schema string) bool {
	return DefaultParserFactory.HasParser(schema)
}

// GetSupportedSchemas 获取支持的站点模式（全局函数）
func GetSupportedSchemas() []string {
	return DefaultParserFactory.GetSupportedSchemas()
}