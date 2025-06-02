// Path: ./models/ctype/list.go

package ctype

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// List 是一个自定义类型，底层是 []string，用于数据库存储。
type List []string

// Scan 实现 sql.Scanner 接口
// 从数据库读取 JSON 字符串并解析成 List
func (l *List) Scan(value interface{}) error {
	// 数据库存储的是 JSON 字符串，通常是 []byte 类型
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan List: expected []byte, got %T", value)
	}

	// 如果为空字符串或 null，赋值为空 List
	if len(bytes) == 0 || string(bytes) == "null" {
		*l = []string{}
		return nil
	}

	// 解析 JSON 字符串为 List
	return json.Unmarshal(bytes, l)
}

// Value 实现 driver.Valuer 接口
// 将 List 序列化为 JSON 字符串用于数据库存储
func (l List) Value() (driver.Value, error) {
	return json.Marshal(l)
}
