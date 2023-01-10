package sqlstorage

import (
	"crawler/engine"
	"crawler/sqldb"
	"crawler/storage"
	"encoding/json"
	"go.uber.org/zap"
)

// SqlStore 是对 Storage 接口的实现，SqlStore 实现了 option 模式，同时它的内部包含了操作数据库的 DBer 接口。
type SqlStorage struct {
	dataDocker  []*storage.DataCell // 分批输出结果缓存
	columnNames []sqldb.Field       // 标题字段
	db          sqldb.DBer          // 包含了操作数据库的 DBer 接口
	Table       map[string]struct{}
	options
}

func New(opts ...Option) (*SqlStorage, error) {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	s := &SqlStorage{}
	s.options = options
	s.Table = make(map[string]struct{})
	var err error
	s.db, err = sqldb.New(
		sqldb.WithLogger(s.logger),
		sqldb.WithConnUrl(s.sqlUrl),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// SqlStore 实现 DBer 接口中的 Save 方法，它主要实现了三个功能：
// - 循环遍历要存储的 DataCell，并判断当前 DataCell 对应的数据库表是否已经被创建。如果表格没有被创建，则调用 CreateTable 创建表格。
// - 如果当前的数据小于 s.BatchCount，则将数据放入到缓存中直接返回（使用缓冲区批量插入数据库可以提高程序的性能）。
// - 如果缓冲区已经满了，则调用 SqlStore.Flush() 方法批量插入数据。

func (s *SqlStorage) Save(dataCells ...*storage.DataCell) error {
	for _, cell := range dataCells { // 循环遍历要存储的 DataCell，并判断当前 DataCell 对应的数据库表是否已经被创建。
		name := cell.GetTableName()
		if _, ok := s.Table[name]; !ok {
			// 创建表
			columnNames := getFields(cell) // 在存储数据时，getFields 用于获取当前数据的表字段与字段类型，这是从采集规则节点的 ItemFields 数组中获得的。
			err := s.db.CreateTable(sqldb.TableData{
				TableName:   name,
				ColumnNames: columnNames,
				AutoKey:     true,
			}) // 如果表格没有被创建，则调用 CreateTable 创建表格。
			if err != nil {
				s.logger.Error("create table failed", zap.Error(err))
			}
			s.Table[name] = struct{}{}
		}
		// 如果缓冲区已经满了，则调用 SqlStore.Flush() 方法批量插入数据。
		if len(s.dataDocker) >= s.BatchCount {
			s.Flush()
		}
		// 如果当前的数据小于 s.BatchCount，则将数据放入到缓存中直接返回（使用缓冲区批量插入数据库可以提高程序的性能）
		s.dataDocker = append(s.dataDocker, cell)
	}
	return nil
}

// Flush 代码的核心是遍历缓冲区，解析每一个 DataCell 中的数据，将扩展后的字段值批量放入 args 参数中，并调用底层 DBer.Insert 方法批量插入数据
func (s *SqlStorage) Flush() error {
	if len(s.dataDocker) == 0 {
		return nil
	}
	args := make([]interface{}, 0)
	for _, datacell := range s.dataDocker {
		ruleName := datacell.Data["Rule"].(string)
		taskName := datacell.Data["Task"].(string)
		fields := engine.GetFields(taskName, ruleName)
		data := datacell.Data["Data"].(map[string]interface{})
		value := []string{}
		for _, field := range fields {
			v := data[field]
			switch v.(type) {
			case nil:
				value = append(value, "")
			case string:
				value = append(value, v.(string))
			default:
				j, err := json.Marshal(v)
				if err != nil {
					value = append(value, "")
				} else {
					value = append(value, string(j))
				}
			}
		}
		value = append(value, datacell.Data["Url"].(string), datacell.Data["Time"].(string))
		for _, v := range value {
			args = append(args, v)
		}
	}

	return s.db.Insert(sqldb.TableData{
		TableName:   s.dataDocker[0].GetTableName(),
		ColumnNames: getFields(s.dataDocker[0]),
		Args:        args,
		DataCount:   len(s.dataDocker),
	})
}

func getFields(cell *storage.DataCell) []sqldb.Field {
	taskName := cell.Data["Task"].(string)
	ruleName := cell.Data["Rule"].(string)
	fields := engine.GetFields(taskName, ruleName)

	var columnNames []sqldb.Field
	for _, field := range fields {
		columnNames = append(columnNames, sqldb.Field{
			Title: field,
			Type:  "MEDIUMTEXT",
		})
	}
	columnNames = append(columnNames,
		sqldb.Field{Title: "Url", Type: "VARCHAR(255)"},
		sqldb.Field{Title: "Time", Type: "VARCHAR(255)"},
	)
	return columnNames
}
