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

// FindEmptyStringField recursively searches a struct for the first field that is an
// empty string ("") or a pointer to an empty string. If an empty string field is
// found, the function returns the name of that field.
//
// The function performs a depth-first search, exploring nested structs and
// pointers to structs. It only inspects exported fields.
//
// Parameters:
//   - obj: The struct (or a pointer to a struct) to inspect. If obj is not a struct
//     or a pointer to one, or if it's nil, the function returns an empty string.
//   - ignoredFields: A set of field names to ignore during the search. A field name
//     in this map will be skipped. A nil map is treated as having no fields to ignore.
//
// Returns:
//   - The name of the first empty string field found.
//   - An empty string ("") if no empty string fields are found, or if the input is invalid.
func FindEmptyStringField(obj any, ignoredFields map[string]bool) string {
	if obj == nil {
		return ""
	}
	// Start the recursive search with the reflected value of the input object.
	return findEmptyStringField(reflect.ValueOf(obj), ignoredFields)
}

// findEmptyStringField is the recursive helper that implements the search logic.
func findEmptyStringField(v reflect.Value, ignoredFields map[string]bool) string {
	// If the value is a pointer, dereference it. We continue until we find
	// a non-pointer value or a nil pointer.
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return ""
		}
		// Recurse on the element the pointer points to.
		return findEmptyStringField(v.Elem(), ignoredFields)
	}

	// This function is designed to work only on structs.
	if v.Kind() != reflect.Struct {
		return ""
	}

	t := v.Type()
	// Iterate over all fields of the struct.
	for i := range v.NumField() {
		fieldVal := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields.
		// if !fieldType.IsExported() {
		// 	continue
		// }

		// Check if the current field name is in the ignore list.
		if ignoredFields != nil && ignoredFields[fieldType.Name] {
			continue
		}

		// Inspect the field based on its kind.
		switch fieldVal.Kind() {
		case reflect.String:
			// Found a string field. Check if it's empty.
			if fieldVal.Len() == 0 {
				return fieldType.Name
			}

		case reflect.Pointer:
			// For pointer fields, we only care about pointers to strings or structs.
			// A nil pointer is considered valid (not empty), so we skip it.
			if fieldVal.IsNil() {
				continue
			}

			elem := fieldVal.Elem()
			switch elem.Kind() {
			case reflect.String:
				// Found a pointer to a string. Check if the string is empty.
				if elem.Len() == 0 {
					return fieldType.Name
				}
			case reflect.Struct:
				// Found a pointer to a struct. Recurse into the nested struct.
				if foundField := findEmptyStringField(elem, ignoredFields); foundField != "" {
					return foundField
				}
			}

		case reflect.Struct:
			// Found a nested struct. Recurse into it to check its fields.
			if foundField := findEmptyStringField(fieldVal, ignoredFields); foundField != "" {
				// If an empty string is found in the nested struct, return its name.
				return foundField
			}
		}
	}

	// No empty string fields were found in this struct.
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
	case reflect.Pointer:
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
