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
	for i := range v.NumField() {
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
func createStructEmptyInstance(structType reflect.Type, created map[reflect.Type]bool) interface{} {
	created[structType] = true
	value := reflect.New(structType).Elem()
	if value.Kind() != reflect.Struct {
		panic("createStructEmptyInstance: value is not a struct")
	}
	for i := range value.NumField() {
		field := value.Field(i)
		if !field.CanSet() {
			continue
		}
		if field.Kind() == reflect.Struct {
			if created[field.Type()] {
				continue
			}
			field.Set(reflect.ValueOf(createStructEmptyInstance(field.Type(), created)))
		} else {
			field.Set(reflect.ValueOf(createEmptyInstance(field.Type(), created)))
		}
	}
	return value.Interface()
}

const defaultMapKey = "key"

// DeepCreateEmptyInstance create a empty instance of the type,
// recursively create the instance of the struct field,include the struct field's pointer
func DeepCreateEmptyInstance(rType reflect.Type) interface{} {
	structTypeMap := make(map[reflect.Type]bool)
	value := reflect.New(rType).Elem()
	if value.Kind() == reflect.Struct {
		return createStructEmptyInstance(rType, structTypeMap)
	}
	if !value.CanSet() {
		return value.Interface()
	}
	return createEmptyInstance(rType, structTypeMap)
}

func createEmptyInstance(rType reflect.Type, created map[reflect.Type]bool) interface{} {
	value := reflect.New(rType).Elem()
	if value.Kind() == reflect.Struct {
		if created[rType] {
			return value.Interface()
		}
		return createStructEmptyInstance(rType, created)
	}
	if !value.CanSet() {
		return value.Interface()
	}
	switch value.Kind() {
	case reflect.Slice:
		slice := reflect.MakeSlice(rType, 1, 1)
		slice.Index(0).Set(reflect.ValueOf(createEmptyInstance(rType.Elem(), created)))
		return slice.Interface()
	case reflect.Map:
		m := reflect.MakeMap(rType)
		if rType.Key().Kind() == reflect.String {
			m.SetMapIndex(reflect.ValueOf(defaultMapKey), reflect.ValueOf(createEmptyInstance(rType.Elem(), created)))
		} else {
			m.SetMapIndex(reflect.ValueOf(createEmptyInstance(rType.Key(), created)), reflect.ValueOf(createEmptyInstance(rType.Elem(), created)))
		}
		return m.Interface()
	case reflect.Ptr:
		if created[rType] {
			return value.Interface()
		}
		instance := createEmptyInstance(rType.Elem(), created)
		ptrInstance := reflect.New(rType.Elem())
		ptrInstance.Elem().Set(reflect.ValueOf(instance))
		return ptrInstance.Interface()
	case reflect.Struct:
		if created[rType] {
			return value.Interface()
		}
		return createStructEmptyInstance(rType, created)
	default:
		return value.Interface()
	}
}
