// Path: ./models/ctype/list.go

package ctype

import (
	"database/sql/driver"
	"strings"
)

// List 是一个自定义类型，底层是 []string，用于数据库存储。
type List []string

// Scan 实现 sql.Scanner 接口
// 从数据库读取 JSON 字符串并解析成 List
func (l *List) Scan(value interface{}) error {
	// 将数据库读取的值断言为 []byte（也可以写为 []uint8）
	// 这是因为数据库 driver（如 MySQL）返回的值类型通常为 []byte
	val, ok := value.([]byte)
	if ok {
		// 如果断言成功，继续处理 val
		// 将 val 转换为字符串后判断是否为空字符串
		if string(val) == "" {
			// 如果是空字符串，就初始化 l 为一个空的字符串切片
			*l = []string{}
			return nil
		}
		// 否则，将字符串按逗号分割成字符串数组，并赋值给 l
		*l = strings.Split(string(val), ",")
	}
	// 如果类型断言失败（不是 []uint8），那么返回 nil（这里逻辑比较简化，没有处理错误）
	return nil
}

// Value 实现 driver.Valuer 接口
// 将 List 序列化为 JSON 字符串用于数据库存储
func (l List) Value() (driver.Value, error) {
	return strings.Join(l, ","), nil
}
