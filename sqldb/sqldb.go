package sqldb

import (
	"database/sql"
	"errors"
	"go.uber.org/zap"
	"strings"
)

type DBer interface {
	CreateTable(t TableData) error
	Insert(t TableData) error
}

type Field struct {
	Title string
	Type  string
}

type TableData struct {
	TableName   string        // 表名
	ColumnNames []Field       // 标题字段，包含了字段名和字段的属性
	Args        []interface{} // 要插入的数据
	DataCount   int           // 插入数据的个数
	AutoKey     bool          // 是否为表创建自增主键
}

// Sqldb 使用 option 模式生成了 SqlDB 结构体，实现了 DBer 接口。
type Sqldb struct {
	options
	db *sql.DB
}

func New(opts ...Option) (*Sqldb, error) {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	d := &Sqldb{}
	d.options = options
	if err := d.OpenDB(); err != nil {
		return nil, err
	}
	return d, nil
}

// OpenDB 方法用于与数据库建立连接，需要从外部传入远程 MySQL 数据库的连接地址。
func (d *Sqldb) OpenDB() error {
	db, err := sql.Open("mysql", d.sqlUrl)
	if err != nil {
		return nil
	}
	db.SetMaxOpenConns(2048)
	db.SetMaxIdleConns(2048)
	if err = db.Ping(); err != nil {
		return nil
	}
	d.db = db
	return nil
}

func (d *Sqldb) CreateTable(t TableData) error {
	if len(t.ColumnNames) == 0 {
		return errors.New("column cannot be empty")
	}
	sql := `CREATE TABLE IF NOT EXISTS ` + t.TableName + ` (`
	if t.AutoKey {
		sql += `id INT(12) NOT NULL PRIMARY KEY AUTO_INCREMENT`
	}
	for _, t := range t.ColumnNames {
		sql += t.Title + ` ` + t.Type + `,`
	}
	sql = sql[:len(sql)-1] + `) ENGINE=MyISAM DEFAULT CHARSET=utf8;`
	d.logger.Debug("create table", zap.String("sql", sql))

	_, err := d.db.Exec(sql)
	return err
}

func (d *Sqldb) Insert(t TableData) error {
	if len(t.ColumnNames) == 0 {
		return errors.New("empty column")
	}
	sql := `INSERT INTO` + t.TableName + `(`
	for _, v := range t.ColumnNames {
		sql += v.Title + `,`
	}
	sql = sql[:len(sql)-1] + `) VALUES `
	blank := ",(" + strings.Repeat(`,?`, len(t.ColumnNames))[1:] + ")"
	sql += strings.Repeat(blank, t.DataCount)[1:] + `;`
	d.logger.Debug("insert table", zap.String("sql", sql))
	_, err := d.db.Exec(sql, t.Args...)
	return err
}
