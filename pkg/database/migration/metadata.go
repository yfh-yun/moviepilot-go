// Package migration 数据库迁移包
package migration

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/yfh-yun/moviepilot-go/internal/database"
)

// MetadataRegistry 元数据注册表
type MetadataRegistry struct {
	mu       sync.RWMutex
	tables   map[string]*TableMetadata
	views    map[string]*ViewMetadata
	indexes  map[string][]*IndexMetadata
}

// TableMetadata 表元数据
type TableMetadata struct {
	Name        string            `json:"name"`
	Columns     []*ColumnMetadata `json:"columns"`
	Constraints []*ConstraintMetadata `json:"constraints"`
	Indexes     []*IndexMetadata  `json:"indexes"`
	Options     map[string]interface{} `json:"options"`
}

// ColumnMetadata 列元数据
type ColumnMetadata struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Nullable     bool                   `json:"nullable"`
	PrimaryKey   bool                   `json:"primary_key"`
	AutoGenerate bool                   `json:"auto_generate"`
	Unique       bool                   `json:"unique"`
	DefaultValue string                 `json:"default_value"`
	MaxLength    int                    `json:"max_length"`
	Precision    int                    `json:"precision"`
	Scale        int                    `json:"scale"`
	Options      map[string]interface{} `json:"options"`
}

// ConstraintMetadata 约束元数据
type ConstraintMetadata struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"` // PRIMARY, FOREIGN, UNIQUE, CHECK
	Table      string   `json:"table"`
	Columns    []string `json:"columns"`
	RefTable   string   `json:"ref_table"`
	RefColumns []string `json:"ref_columns"`
	OnDelete   string   `json:"on_delete"`
	OnUpdate   string   `json:"on_update"`
}

// IndexMetadata 索引元数据
type IndexMetadata struct {
	Name    string   `json:"name"`
	Table   string   `json:"table"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Type    string   `json:"type"`
	Where   string   `json:"where"`
}

// ViewMetadata 视图元数据
type ViewMetadata struct {
	Name       string `json:"name"`
	SQL        string `json:"sql"`
	Definition string `json:"definition"`
	Options    map[string]interface{} `json:"options"`
}

// NewMetadataRegistry 创建元数据注册表
func NewMetadataRegistry() *MetadataRegistry {
	return &MetadataRegistry{
		tables:  make(map[string]*TableMetadata),
		views:   make(map[string]*ViewMetadata),
		indexes: make(map[string][]*IndexMetadata),
	}
}

// RegisterTable 注册表
func (r *MetadataRegistry) RegisterTable(name string, table *TableMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if table == nil {
		return fmt.Errorf("表元数据不能为空")
	}

	table.Name = name
	r.tables[name] = table

	// 注册索引
	if len(table.Indexes) > 0 {
		r.indexes[name] = table.Indexes
	}

	return nil
}

// RegisterView 注册视图
func (r *MetadataRegistry) RegisterView(name string, view *ViewMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if view == nil {
		return fmt.Errorf("视图元数据不能为空")
	}

	view.Name = name
	r.views[name] = view

	return nil
}

// GetTable 获取表元数据
func (r *MetadataRegistry) GetTable(name string) (*TableMetadata, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	table, exists := r.tables[name]
	return table, exists
}

// GetView 获取视图元数据
func (r *MetadataRegistry) GetView(name string) (*ViewMetadata, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	view, exists := r.views[name]
	return view, exists
}

// GetAllTables 获取所有表
func (r *MetadataRegistry) GetAllTables() map[string]*TableMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*TableMetadata)
	for name, table := range r.tables {
		// 深拷贝
		result[name] = r.copyTable(table)
	}

	return result
}

// GetAllViews 获取所有视图
func (r *MetadataRegistry) GetAllViews() map[string]*ViewMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*ViewMetadata)
	for name, view := range r.views {
		// 深拷贝
		result[name] = r.copyView(view)
	}

	return result
}

// GetTableNames 获取所有表名
func (r *MetadataRegistry) GetTableNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name := range r.tables {
		names = append(names, name)
	}

	return names
}

// GetViewNames 获取所有视图名
func (r *MetadataRegistry) GetViewNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name := range r.views {
		names = append(names, name)
	}

	return names
}

// ReflectFromStruct 从结构体反射元数据
func (r *MetadataRegistry) ReflectFromStruct(model interface{}, tableName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("模型必须是结构体类型")
	}

	table := &TableMetadata{
		Name:        tableName,
		Columns:     make([]*ColumnMetadata, 0),
		Constraints: make([]*ConstraintMetadata, 0),
		Indexes:     make([]*IndexMetadata, 0),
		Options:     make(map[string]interface{}),
	}

	// 遍历结构体字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		column := r.reflectColumn(field)
		if column != nil {
			table.Columns = append(table.Columns, column)

			// 检查是否为主键
			if column.PrimaryKey {
				pkConstraint := &ConstraintMetadata{
					Name:    fmt.Sprintf("pk_%s", tableName),
					Type:    "PRIMARY",
					Table:   tableName,
					Columns: []string{column.Name},
				}
				table.Constraints = append(table.Constraints, pkConstraint)
			}

			// 检查是否为唯一索引
			if column.Unique {
				uniqueIndex := &IndexMetadata{
					Name:    fmt.Sprintf("idx_%s_%s_unique", tableName, column.Name),
					Table:   tableName,
					Columns: []string{column.Name},
					Unique:  true,
				}
				table.Indexes = append(table.Indexes, uniqueIndex)
			}
		}
	}

	r.tables[tableName] = table
	return nil
}

// reflectColumn 反射列信息
func (r *MetadataRegistry) reflectColumn(field reflect.StructField) *ColumnMetadata {
	// 获取标签
	tag := field.Tag.Get("db")
	if tag == "-" || tag == "" {
		return nil
	}

	// 解析标签
	column := &ColumnMetadata{
		Name:     r.parseColumnName(tag),
		Type:     r.parseColumnType(field),
		Nullable: r.parseNullable(field),
		Options:  make(map[string]interface{}),
	}

	// 解析其他标签属性
	if pk := field.Tag.Get("primary_key"); pk == "true" {
		column.PrimaryKey = true
	}

	if auto := field.Tag.Get("auto_increment"); auto == "true" {
		column.AutoGenerate = true
	}

	if unique := field.Tag.Get("unique"); unique == "true" {
		column.Unique = true
	}

	if def := field.Tag.Get("default"); def != "" {
		column.DefaultValue = def
	}

	if size := field.Tag.Get("size"); size != "" {
		if maxLen, err := parseSize(size); err == nil {
			column.MaxLength = maxLen
		}
	}

	return column
}

// parseColumnName 解析列名
func (r *MetadataRegistry) parseColumnName(tag string) string {
	if idx := strings.Index(tag, ","); idx >= 0 {
		return tag[:idx]
	}
	return tag
}

// parseColumnType 解析列类型
func (r *MetadataRegistry) parseColumnType(field reflect.StructField) string {
	t := field.Type
	
	// 处理指针类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		maxLen := 255 // 默认长度
		if tag := field.Tag.Get("size"); tag != "" {
			if len, err := parseSize(tag); err == nil && len > 0 {
				maxLen = len
			}
		}
		return fmt.Sprintf("VARCHAR(%d)", maxLen)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if t.Name() == "Time" || t.PkgPath() == "time" {
			return "TIMESTAMP"
		}
		return "INTEGER"

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "INTEGER"

	case reflect.Float32, reflect.Float64:
		return "REAL"

	case reflect.Bool:
		return "BOOLEAN"

	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return "BLOB"
		}
		return "TEXT"

	case reflect.Struct:
		if t.Name() == "Time" || t.PkgPath() == "time" {
			return "TIMESTAMP"
		}
		return "TEXT"

	default:
		return "TEXT"
	}
}

// parseNullable 解析是否可为空
func (r *MetadataRegistry) parseNullable(field reflect.StructField) bool {
	// 指针类型通常可为空
	if field.Type.Kind() == reflect.Ptr {
		return true
	}

	// 检查not null标签
	if tag := field.Tag.Get("not_null"); tag == "true" {
		return false
	}

	// 主键通常不可为空
	if tag := field.Tag.Get("primary_key"); tag == "true" {
		return false
	}

	return true // 默认可为空
}

// parseSize 解析大小
func parseSize(size string) (int, error) {
	// 移除非数字字符
	var numStr string
	for _, ch := range size {
		if ch >= '0' && ch <= '9' {
			numStr += string(ch)
		}
	}

	if numStr == "" {
		return 0, fmt.Errorf("无效的大小格式: %s", size)
	}

	var result int
	_, err := fmt.Sscanf(numStr, "%d", &result)
	return result, err
}

// copyTable 深拷贝表元数据
func (r *MetadataRegistry) copyTable(table *TableMetadata) *TableMetadata {
	if table == nil {
		return nil
	}

	copy := &TableMetadata{
		Name:        table.Name,
		Columns:     make([]*ColumnMetadata, len(table.Columns)),
		Constraints: make([]*ConstraintMetadata, len(table.Constraints)),
		Indexes:     make([]*IndexMetadata, len(table.Indexes)),
		Options:     make(map[string]interface{}),
	}

	// 拷贝列
	for i, col := range table.Columns {
		copy.Columns[i] = r.copyColumn(col)
	}

	// 拷贝约束
	for i, constraint := range table.Constraints {
		copy.Constraints[i] = r.copyConstraint(constraint)
	}

	// 拷贝索引
	for i, index := range table.Indexes {
		copy.Indexes[i] = r.copyIndex(index)
	}

	// 拷贝选项
	for k, v := range table.Options {
		copy.Options[k] = v
	}

	return copy
}

// copyColumn 深拷贝列元数据
func (r *MetadataRegistry) copyColumn(col *ColumnMetadata) *ColumnMetadata {
	if col == nil {
		return nil
	}

	copy := &ColumnMetadata{
		Name:         col.Name,
		Type:         col.Type,
		Nullable:     col.Nullable,
		PrimaryKey:   col.PrimaryKey,
		AutoGenerate: col.AutoGenerate,
		Unique:       col.Unique,
		DefaultValue: col.DefaultValue,
		MaxLength:    col.MaxLength,
		Precision:    col.Precision,
		Scale:        col.Scale,
		Options:      make(map[string]interface{}),
	}

	for k, v := range col.Options {
		copy.Options[k] = v
	}

	return copy
}

// copyConstraint 深拷贝约束元数据
func (r *MetadataRegistry) copyConstraint(constraint *ConstraintMetadata) *ConstraintMetadata {
	if constraint == nil {
		return nil
	}

	return &ConstraintMetadata{
		Name:       constraint.Name,
		Type:       constraint.Type,
		Table:      constraint.Table,
		Columns:    append([]string{}, constraint.Columns...),
		RefTable:   constraint.RefTable,
		RefColumns: append([]string{}, constraint.RefColumns...),
		OnDelete:   constraint.OnDelete,
		OnUpdate:   constraint.OnUpdate,
	}
}

// copyIndex 深拷贝索引元数据
func (r *MetadataRegistry) copyIndex(index *IndexMetadata) *IndexMetadata {
	if index == nil {
		return nil
	}

	return &IndexMetadata{
		Name:    index.Name,
		Table:   index.Table,
		Columns: append([]string{}, index.Columns...),
		Unique:  index.Unique,
		Type:    index.Type,
		Where:   index.Where,
	}
}

// copyView 深拷贝视图元数据
func (r *MetadataRegistry) copyView(view *ViewMetadata) *ViewMetadata {
	if view == nil {
		return nil
	}

	return &ViewMetadata{
		Name:       view.Name,
		SQL:        view.SQL,
		Definition: view.Definition,
		Options:    view.Options,
	}
}

// GenerateCreateSQL 生成创建表的SQL
func (r *MetadataRegistry) GenerateCreateSQL(tableName string, databaseType string) (string, error) {
	table, exists := r.GetTable(tableName)
	if !exists {
		return "", fmt.Errorf("表 %s 不存在", tableName)
	}

	return r.generateTableSQL(table, databaseType), nil
}

// generateTableSQL 生成表SQL
func (r *MetadataRegistry) generateTableSQL(table *TableMetadata, databaseType string) string {
	var sql strings.Builder

	sql.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", table.Name))

	// 生成列定义
	var columns []string
	for _, column := range table.Columns {
		colDef := r.generateColumnDefinition(column, databaseType)
		columns = append(columns, colDef)
	}

	sql.WriteString("  " + strings.Join(columns, ",\n  "))

	// 生成约束定义
	if len(table.Constraints) > 0 {
		sql.WriteString(",\n")
		for _, constraint := range table.Constraints {
			if constraint.Type == "PRIMARY" {
				constraintDef := r.generateConstraintDefinition(constraint, databaseType)
				sql.WriteString("  " + constraintDef + "\n")
			}
		}
	}

	sql.WriteString(")")

	// 添加表选项
	if databaseType == "mysql" {
		// MySQL特定的表选项
		if charset, exists := table.Options["charset"]; exists {
			sql.WriteString(fmt.Sprintf(" CHARACTER SET %s", charset))
		}
		if collate, exists := table.Options["collate"]; exists {
			sql.WriteString(fmt.Sprintf(" COLLATE %s", collate))
		}
		if engine, exists := table.Options["engine"]; exists {
			sql.WriteString(fmt.Sprintf(" ENGINE=%s", engine))
		}
	}

	sql.WriteString(";")

	return sql.String()
}

// generateColumnDefinition 生成列定义
func (r *MetadataRegistry) generateColumnDefinition(column *ColumnMetadata, databaseType string) string {
	def := fmt.Sprintf("%s %s", column.Name, column.Type)

	if !column.Nullable {
		def += " NOT NULL"
	}

	if column.DefaultValue != "" {
		if databaseType == "postgresql" {
			def += fmt.Sprintf(" DEFAULT %s", column.DefaultValue)
		} else {
			def += fmt.Sprintf(" DEFAULT %s", column.DefaultValue)
		}
	}

	if column.Unique {
		def += " UNIQUE"
	}

	if column.PrimaryKey && column.AutoGenerate {
		if databaseType == "postgresql" {
			def += " GENERATED BY DEFAULT AS IDENTITY"
		} else {
			def += " AUTOINCREMENT"
		}
	}

	return def
}

// generateConstraintDefinition 生成约束定义
func (r *MetadataRegistry) generateConstraintDefinition(constraint *ConstraintMetadata, databaseType string) string {
	if constraint.Type == "PRIMARY" {
		return fmt.Sprintf("CONSTRAINT %s PRIMARY KEY (%s)", 
			constraint.Name, strings.Join(constraint.Columns, ", "))
	}

	return ""
}

// Validate 验证元数据
func (r *MetadataRegistry) Validate() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 验证表名
	for tableName, table := range r.tables {
		if tableName == "" {
			return fmt.Errorf("表名不能为空")
		}

		if table == nil {
			return fmt.Errorf("表 %s 的元数据为空", tableName)
		}

		// 验证列
		if len(table.Columns) == 0 {
			return fmt.Errorf("表 %s 必须至少有一个列", tableName)
		}

		for _, column := range table.Columns {
			if column.Name == "" {
				return fmt.Errorf("表 %s 中存在列名为空的列", tableName)
			}

			if column.Type == "" {
				return fmt.Errorf("表 %s 的列 %s 类型不能为空", tableName, column.Name)
			}
		}

		// 验证主键
		hasPrimaryKey := false
		for _, column := range table.Columns {
			if column.PrimaryKey {
				hasPrimaryKey = true
				break
			}
		}

		if !hasPrimaryKey {
			// 检查约束中是否有主键
			for _, constraint := range table.Constraints {
				if constraint.Type == "PRIMARY" {
					hasPrimaryKey = true
					break
				}
			}
		}

		// SQLite可以没有主键，但建议有
	}

	return nil
}