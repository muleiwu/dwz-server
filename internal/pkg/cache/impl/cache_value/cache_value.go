package cache_value

import (
	"fmt"
	"reflect"

	"github.com/vmihailenco/msgpack/v5"
)

// CacheValue 封装缓存值，保留类型信息
type CacheValue struct {
	Type  string      `msgpack:"t"`
	Value interface{} `msgpack:"v"`
}

// Encode 序列化缓存值
func Encode(value interface{}) ([]byte, error) {
	cacheValue := CacheValue{
		Type:  getType(value),
		Value: value,
	}

	return msgpack.Marshal(cacheValue)
}

// Decode 带反序列化
func Decode(data []byte, obj any) error {

	var cacheValue CacheValue
	if err := msgpack.Unmarshal(data, &cacheValue); err != nil {
		return err
	}

	// 使用反射赋值
	return assignValue(obj, cacheValue.Value)
}

// 辅助函数
func getType(v interface{}) string {
	return reflect.TypeOf(v).String()
}

// assignValue 使用反射将值赋给目标对象
func assignValue(obj any, value interface{}) error {
	if obj == nil {
		return fmt.Errorf("obj cannot be nil")
	}

	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Ptr {
		return fmt.Errorf("obj must be a pointer")
	}

	objElem := objValue.Elem()
	if !objElem.CanSet() {
		return fmt.Errorf("obj cannot be set")
	}

	valueReflect := reflect.ValueOf(value)
	if !valueReflect.IsValid() {
		return fmt.Errorf("value is not valid")
	}

	// 确保类型匹配
	if objElem.Type() != valueReflect.Type() {
		return fmt.Errorf("type mismatch: expected %s, got %s", objElem.Type(), valueReflect.Type())
	}

	objElem.Set(valueReflect)
	return nil
}
