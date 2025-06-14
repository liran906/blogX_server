// Path: ./utils/mps/enter.go

package mps

import (
	"encoding/json"
	"reflect"
)

// StructToMap 将结构体转换为 map[string]any
// - data: 要转换的结构体（必须是 struct 实例）
// - tag: 要读取的结构体标签（如 json、form、db 等）
// 返回值: 一个 map，其中 key 为标签值，value 为字段值
func StructToMap(data any, tag string) map[string]any {
	// 创建空 map 用于存储字段名与值
	m := make(map[string]any)

	// 获取 data 的反射值（Value）和类型（Type）
	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	// 结构体必须是指向 struct 的指针或 struct 本身
	// 若是指针，取其元素
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	// 遍历结构体的所有字段
	for i := 0; i < v.NumField(); i++ {
		// 获取字段值和字段类型信息
		field := v.Field(i)
		fieldType := t.Field(i)

		// 获取指定 tag 对应的字段名，如 json:"name" 中的 "name"
		tagValue := fieldType.Tag.Get(tag)

		// 跳过没有 tag 或标记为 "-"（不序列化）的字段
		if tagValue == "" || tagValue == "-" {
			continue
		}

		// 如果字段是指针类型，需要进一步处理
		if field.Kind() == reflect.Ptr {
			// 若指针为 nil，直接跳过
			if field.IsNil() {
				continue
			}

			// 获取指针指向的实际值
			val := field.Elem().Interface()

			// 判断这个值是否是切片类型（例如 *[]int）
			if field.Elem().Kind() == reflect.Slice {
				// 将切片序列化成 JSON 字符串，方便存储和展示
				// 否则存入数据库会报错
				// 而且，直接存切片在 map 里，会变成 [1 2 3] 格式，不标准
				byteData, _ := json.Marshal(val)
				m[tagValue] = string(byteData)
			} else {
				// 如果不是切片，就直接把实际值赋给 map
				m[tagValue] = val
			}
			continue // 指针处理完毕，跳过后续处理
		}

		// 非指针字段直接存入 map
		m[tagValue] = field.Interface()
	}

	// 返回结构体转换后的 map
	return m
}
