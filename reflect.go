package doraemon

import (
	"reflect"
)

func IsNil(x any) bool {
	if x == nil {
		return true
	}
	return reflect.ValueOf(x).IsNil()
}

// 查找结构体中的空字符串字段。如果找到，则返回字段名；否则返回空字符串。
// string指针可以为空，但是string不可以为空。
func FindStructEmptyStringField(s any, ignores map[string]bool) string {
	if s == nil {
		return ""
	}
	return findStructEmptyStringField(reflect.ValueOf(s), ignores)
}

func findStructEmptyStringField(v reflect.Value, ignores map[string]bool) string {
	if v.Kind() != reflect.Pointer && v.Kind() != reflect.Struct {
		return ""
	}
	kind := v.Kind()
	if kind == reflect.Pointer {
		if v.IsNil() {
			return ""
		}
		return findStructEmptyStringField(v.Elem(), ignores)
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if ignores != nil && ignores[t.Field(i).Name] {
			continue
		}
		field := v.Field(i)
		// 如果字段是私有的
		// if !field.CanSet() {
		// 	continue
		// }
		switch field.Kind() {
		case reflect.String:
			if field.Len() == 0 {
				return t.Field(i).Name
			}
		case reflect.Ptr:
			if field.IsNil() {
				continue
			}
			field = field.Elem()
			if field.Kind() == reflect.String && field.Len() == 0 {
				return t.Field(i).Name
			}
			if field.Kind() == reflect.Struct {
				return findStructEmptyStringField(field, ignores)
			}
			continue
		case reflect.Struct:
			return findStructEmptyStringField(field, ignores)
		}
	}
	return ""
}

// 创建一个空的结构体实例，只能传入结构体类型
func createStructEmptyInstance(structType reflect.Type) interface{} {
	value := reflect.New(structType).Elem()
	if value.Kind() != reflect.Struct {
		panic("createStructEmptyInstance: value is not a struct")
	}
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		if !field.CanSet() {
			continue
		}
		if field.Kind() == reflect.Struct {
			field.Set(reflect.ValueOf(createStructEmptyInstance(field.Type())))
		} else {
			field.Set(reflect.ValueOf(CreateEmptyInstance(field.Type())))
		}
	}
	return value.Interface()
}

const defaultMapKey = "key"

func CreateEmptyInstance(rType reflect.Type) interface{} {
	value := reflect.New(rType).Elem()
	if value.Kind() == reflect.Struct {
		return createStructEmptyInstance(rType)
	}
	if !value.CanSet() {
		return value.Interface()
	}
	switch value.Kind() {
	case reflect.Slice:
		slice := reflect.MakeSlice(rType, 1, 1)
		slice.Index(0).Set(reflect.ValueOf(CreateEmptyInstance(rType.Elem())))
		return slice.Interface()
	case reflect.Map:
		m := reflect.MakeMap(rType)
		if rType.Key().Kind() == reflect.String {
			m.SetMapIndex(reflect.ValueOf(defaultMapKey), reflect.ValueOf(CreateEmptyInstance(rType.Elem())))
		} else {
			m.SetMapIndex(reflect.ValueOf(CreateEmptyInstance(rType.Key())), reflect.ValueOf(CreateEmptyInstance(rType.Elem())))
		}
		return m.Interface()
	case reflect.Ptr:
		instance := CreateEmptyInstance(rType.Elem())
		ptrInstance := reflect.New(rType.Elem())
		ptrInstance.Elem().Set(reflect.ValueOf(instance))
		return ptrInstance.Interface()
	case reflect.Struct:
		return createStructEmptyInstance(rType)
	default:
		return value.Interface()
	}
}
