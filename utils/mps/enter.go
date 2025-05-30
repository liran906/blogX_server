// Path: ./blogX_server/utils/mps/enter.go

package mps

import "reflect"

// StructToMap 将结构体转换为 map[string]any
// - data: 要转换的结构体（必须是 struct 实例）
// - tag: 需要读取的标签名（如 json、db、form 等）
// 返回值: 以 tag 为 key，字段值为 value 的 map
func StructToMap(data any, tag string) map[string]any {
	// 创建一个空的 map，用于存储结果
	m := make(map[string]any)
	// 获取 data 的反射值（用于访问字段值）
	v := reflect.ValueOf(data)
	// 获取 data 的反射类型（用于访问字段类型信息）
	t := reflect.TypeOf(data)

	// 遍历结构体的每一个字段
	for i := 0; i < v.NumField(); i++ {
		// 获取第 i 个字段的值
		field := v.Field(i)
		// 获取第 i 个字段的类型信息
		fieldType := t.Field(i)
		// 获取该字段的标签值（如 json:"name" 中的 "name"）
		tagValue := fieldType.Tag.Get(tag)

		// 如果标签值为空字符串或为 "-"（表示忽略该字段），则跳过
		if tagValue == "" || tagValue == "-" {
			continue
		}

		// 如果字段是指针类型
		if field.Kind() == reflect.Ptr {
			// 先判断是否为 nil，如果是 nil，跳过该字段
			if field.IsNil() {
				continue
			}
			// 获取指针指向的实际值
			field = field.Elem()
		}
		// 将标签值作为 key，字段值作为 value 存入 map
		m[tagValue] = field.Interface()
	}
	// 返回最终生成的 map
	return m
}
