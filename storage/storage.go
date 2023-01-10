package storage

// 爬取到足够的信息之后，为了将数据存储起来，首先我们需要完成对数据的抽象。
// 在这里我将每一条要存储的数据都抽象为了 DataCell 结构。我们可以把 DataCell 想象为 MySQL 中的一行数据。

// DataCell 中的 Key 为“Task”的数据存储了当前的任务名，Key 为“Rule”的数据存储了当前的规则名，Key 为“Url”的数据存储了当前的网址，Key 为“Time”的数据存储了当前的时间。
// 而最重要的 Key 为“Data”的数据存储了当前核心的数据，即当前书籍的详细信息。
type DataCell struct {
	Data map[string]interface{}
}

func (d *DataCell) GetTableName() string {
	return d.Data["Task"].(string)
}

func (d *DataCell) GetTaskName() string {
	return d.Data["Task"].(string)
}

// Storage 创建了一个数据存储的接口，Storage 中包含了 Save 方法，任何实现了 Save 方法的后端引擎都可以存储数据。
type Storage interface {
	Save(data ...*DataCell) error
}
