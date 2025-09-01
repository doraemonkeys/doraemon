package doraemon

import (
	"reflect"
	"testing"
)

func TestFindStructEmptyStringField(t *testing.T) {
	tests := []struct {
		name    string
		s       any
		ignores map[string]bool
		want    string
	}{
		{"1", struct {
			A string
			B string
		}{A: "a", B: ""}, nil, "B"},
		{"2", struct {
			A string
			B string
		}{A: "a", B: "b"}, nil, ""},
		{"3", &struct {
			A string
			B string
		}{A: "a", B: "b"}, nil, ""},
		{"4", &struct {
			A string
			B string
		}{A: "a", B: ""}, nil, "B"},
		{"5", struct {
			A string
			B string
			C *string
		}{A: "a", B: "b", C: nil}, nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindEmptyStringField(tt.s, tt.ignores); got != tt.want {
				t.Errorf("name: %v FindStructEmptyStringField() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

type SimpleStruct struct {
	SimpleStructA int
	SimpleStructB string
}

type MyStruct struct {
	A    int
	AA   *int
	B    string
	BB   *string
	C    float64
	CC   *float64
	D    bool
	E    []string
	EE   *[]string
	F    map[string]string
	FF   *map[string]string
	FFF  map[string]*string
	FFFF map[string]*SimpleStruct
	G    SimpleStruct
	GG   *SimpleStruct
	c    int
	cc   *int
}

func emptyInstance() *MyStruct {
	return &MyStruct{
		A:    0,
		AA:   new(int),
		B:    "",
		BB:   new(string),
		C:    0,
		CC:   new(float64),
		D:    false,
		E:    []string{""},
		EE:   &[]string{""},
		F:    map[string]string{defaultMapKey: ""},
		FF:   &map[string]string{defaultMapKey: ""},
		FFF:  map[string]*string{defaultMapKey: new(string)},
		FFFF: map[string]*SimpleStruct{defaultMapKey: {0, ""}},
		G:    SimpleStruct{0, ""},
		GG:   &SimpleStruct{0, ""},
		c:    0,
		cc:   nil,
	}
}

func TestCreateEmptyInstance(t *testing.T) {
	type args struct {
		rType reflect.Type
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "TestCreateEmptyInstance_Struct",
			args: args{
				rType: reflect.TypeOf(MyStruct{}),
			},
			want: *emptyInstance(),
		},
		{
			name: "TestCreateEmptyInstance_Slice",
			args: args{
				rType: reflect.TypeOf([]MyStruct{}),
			},
			want: []MyStruct{*emptyInstance()},
		},
		{
			name: "TestCreateEmptyInstance_Map",
			args: args{
				rType: reflect.TypeOf(map[string]MyStruct{}),
			},
			want: map[string]MyStruct{defaultMapKey: *emptyInstance()},
		},
		{
			name: "TestCreateEmptyInstance_Pointer",
			args: args{
				rType: reflect.TypeOf(&MyStruct{}),
			},
			want: emptyInstance(),
		},
		{
			name: "TestCreateEmptyInstance_SliceInt",
			args: args{
				rType: reflect.TypeOf([]int{}),
			},
			want: []int{0},
		},
		{
			name: "TestCreateEmptyInstance_Map[string]int",
			args: args{
				rType: reflect.TypeOf(map[string]int{}),
			},
			want: map[string]int{defaultMapKey: 0},
		},
		{
			name: "TestCreateEmptyInstance_Map[string]intPointer",
			args: args{
				rType: reflect.TypeOf(&map[string]int{}),
			},
			want: &map[string]int{defaultMapKey: 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeepCreateEmptyInstance(tt.args.rType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateEmptyInstance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateEmptyInstanceMapPtrKey(t *testing.T) {
	got := DeepCreateEmptyInstance(reflect.TypeOf(&map[*int]string{}))
	gotType := reflect.TypeOf(got)
	if gotType.Kind() != reflect.Ptr {
		t.Errorf("CreateEmptyInstance() = %v, want %v", gotType.Kind(), reflect.Ptr)
	}
	if gotType.Elem().Kind() != reflect.Map {
		t.Errorf("CreateEmptyInstance() = %v, want %v", gotType.Elem().Kind(), reflect.Map)
	}
	if gotType.Elem().Key().Kind() != reflect.Ptr {
		t.Errorf("CreateEmptyInstance() = %v, want %v", gotType.Elem().Key().Kind(), reflect.Ptr)
	}
	if gotType.Elem().Elem().Kind() != reflect.String {
		t.Errorf("CreateEmptyInstance() = %v, want %v", gotType.Elem().Elem().Kind(), reflect.String)
	}
}

// SimpleStruct 用于基本测试
type SimpleStruct2 struct {
	Name    string
	Address string
	ZipCode string
}

// PointerFieldStruct 用于测试 *string 字段
type PointerFieldStruct struct {
	ID          string
	Description *string
	Notes       *string
}

// NestedStruct 用于测试嵌套结构体
type NestedStruct struct {
	Street  string
	City    string
	Country string
}

// ComplexStruct 用于测试嵌套和指针嵌套
type ComplexStruct struct {
	RefID     string
	Primary   NestedStruct
	Secondary *NestedStruct
	Comment   *string
}

// --- 单元测试 ---

func TestFindStructEmptyStringField2(t *testing.T) {
	// 辅助变量，用于测试 *string
	emptyStr := ""
	nonEmptyStr := "some value"

	// 定义测试用例
	testCases := []struct {
		name    string          // 测试用例的描述
		s       any             // 输入的结构体
		ignores map[string]bool // 要忽略的字段
		want    string          // 期望返回的字段名
	}{
		// --- 基本情况 ---
		{
			name: "Input is nil",
			s:    nil,
			want: "",
		},
		{
			name: "Input is not a struct",
			s:    12345,
			want: "",
		},
		{
			name: "All fields are filled",
			s:    SimpleStruct2{Name: "Alice", Address: "123 Main St", ZipCode: "90210"},
			want: "",
		},
		{
			name: "First string field is empty",
			s:    SimpleStruct2{Name: "", Address: "123 Main St", ZipCode: "90210"},
			want: "Name",
		},
		{
			name: "Last string field is empty",
			s:    SimpleStruct2{Name: "Alice", Address: "123 Main St", ZipCode: ""},
			want: "ZipCode",
		},
		{
			name: "Multiple empty fields, should find first",
			s:    SimpleStruct2{Name: "Alice", Address: "", ZipCode: ""},
			want: "Address",
		},

		// --- 指针处理 ---
		{
			name: "Input is a pointer to a struct with an empty field",
			s:    &SimpleStruct2{Name: "Bob", Address: "", ZipCode: "10001"},
			want: "Address",
		},
		{
			name: "Input is a nil pointer to a struct",
			s:    (*SimpleStruct2)(nil),
			want: "",
		},
		{
			name: "Pointer string field is nil",
			s:    PointerFieldStruct{ID: "p1", Description: nil, Notes: &nonEmptyStr},
			want: "", // nil pointer is not an empty string
		},
		{
			name: "Pointer string field points to an empty string",
			s:    PointerFieldStruct{ID: "p1", Description: &emptyStr, Notes: &nonEmptyStr},
			want: "Description",
		},
		{
			name: "Pointer string field points to a non-empty string",
			s:    PointerFieldStruct{ID: "p1", Description: &nonEmptyStr, Notes: &nonEmptyStr},
			want: "",
		},

		// --- 嵌套结构体 ---
		{
			name: "Nested struct value has an empty field",
			s:    ComplexStruct{RefID: "c1", Primary: NestedStruct{Street: "", City: "Testville"}},
			want: "Street",
		},
		{
			name: "Nested struct pointer has an empty field",
			s: &ComplexStruct{
				RefID:     "c2",
				Primary:   NestedStruct{Street: "456 Oak Ave", City: "Testville", Country: "Testcountry"},
				Secondary: &NestedStruct{Street: "789 Pine Ln", City: "", Country: "Testcountry"},
			},
			want: "City",
		},
		{
			name: "Nested struct pointer is nil",
			s: &ComplexStruct{
				RefID:     "c3",
				Primary:   NestedStruct{Street: "456 Oak Ave", City: "Testville", Country: "Testcountry"},
				Secondary: nil,
			},
			want: "",
		},

		// --- ignores 映射 ---
		{
			name:    "Ignore the only empty field",
			s:       SimpleStruct2{Name: "", Address: "123 Main St", ZipCode: "90210"},
			ignores: map[string]bool{"Name": true},
			want:    "",
		},
		{
			name:    "Ignore the first empty field, find the second",
			s:       SimpleStruct2{Name: "", Address: "", ZipCode: "90210"},
			ignores: map[string]bool{"Name": true},
			want:    "Address",
		},
		{
			name:    "Ignore a field in a nested struct",
			s:       ComplexStruct{RefID: "c4", Primary: NestedStruct{Street: "", City: "Testville", Country: "Testcountry"}},
			ignores: map[string]bool{"Street": true},
			want:    "",
		},
		{
			name: "Ignore an empty field in a nested struct, find another",
			s: &ComplexStruct{
				RefID:     "c5",
				Primary:   NestedStruct{Street: "", City: ""},
				Secondary: nil,
			},
			ignores: map[string]bool{"Street": true},
			want:    "City",
		},
		{
			name:    "Ignores map is nil, normal behavior",
			s:       SimpleStruct2{Name: ""},
			ignores: nil,
			want:    "Name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := FindEmptyStringField(tc.s, tc.ignores)
			if got != tc.want {
				t.Errorf("FindStructEmptyStringField() = %q, want %q", got, tc.want)
			}
		})
	}
}

// SimpleStruct is a basic struct for testing.
type SimpleStruct3 struct {
	Name string
	Age  int
}

// PointerStruct tests fields that are pointers.
type PointerStruct struct {
	Name        *string
	Description *string
}

// NestedStruct tests nested structs by value.
type NestedStruct3 struct {
	ID    string
	Inner SimpleStruct3
}

// PointerNestedStruct tests nested structs by pointer.
type PointerNestedStruct struct {
	ID    string
	Inner *SimpleStruct3
}

// PrivateFieldStruct tests behavior with unexported fields.
type PrivateFieldStruct struct {
	Exported   string
	unexported string // This field is private
}

// MultiLevelNestedStruct tests multiple levels of nesting.
type MultiLevelNestedStruct struct {
	Level1Field string
	NextLevel   PointerNestedStruct
}

// --- Test Cases ---

func TestFindEmptyStringField(t *testing.T) {
	// Helper variables for pointer tests
	emptyStr := ""
	nonEmptyStr := "hello"

	testCases := []struct {
		name          string
		inputObj      any
		ignoredFields map[string]bool
		expected      string
	}{
		// --- Basic Cases ---
		{
			name:     "Simple struct with empty string field",
			inputObj: SimpleStruct3{Name: "", Age: 30},
			expected: "Name",
		},
		{
			name:     "Simple struct with no empty fields",
			inputObj: SimpleStruct3{Name: "Bob", Age: 30},
			expected: "",
		},
		{
			name:     "Struct with multiple fields, first is empty",
			inputObj: struct{ ID, Name string }{ID: "", Name: "Test"},
			expected: "ID",
		},
		{
			name:     "Struct with multiple fields, second is empty",
			inputObj: struct{ ID, Name string }{ID: "123", Name: ""},
			expected: "Name",
		},

		// --- Pointer Field Cases ---
		{
			name:     "Pointer to an empty string",
			inputObj: PointerStruct{Name: &emptyStr},
			expected: "Name",
		},
		{
			name:     "Pointer to a non-empty string",
			inputObj: PointerStruct{Name: &nonEmptyStr},
			expected: "",
		},
		{
			name:     "Nil pointer to string field",
			inputObj: PointerStruct{Name: nil},
			expected: "",
		},

		// --- Nested Struct Cases ---
		{
			name:     "Nested struct with empty field",
			inputObj: NestedStruct3{ID: "outer", Inner: SimpleStruct3{Name: ""}},
			expected: "Name",
		},
		{
			name:     "Pointer to nested struct with empty field",
			inputObj: PointerNestedStruct{ID: "outer", Inner: &SimpleStruct3{Name: ""}},
			expected: "Name",
		},
		{
			name:     "Nil pointer to nested struct",
			inputObj: PointerNestedStruct{ID: "outer", Inner: nil},
			expected: "",
		},
		{
			name: "Multi-level nested struct with empty field",
			inputObj: MultiLevelNestedStruct{
				Level1Field: "L1",
				NextLevel: PointerNestedStruct{
					ID:    "L2",
					Inner: &SimpleStruct3{Name: ""},
				},
			},
			expected: "Name",
		},

		// --- Ignored Fields Cases ---
		{
			name:          "Empty field is in ignored list",
			inputObj:      SimpleStruct3{Name: ""},
			ignoredFields: map[string]bool{"Name": true},
			expected:      "",
		},
		{
			name: "First empty field is ignored, finds second",
			inputObj: struct {
				A string
				B string
			}{A: "", B: ""},
			ignoredFields: map[string]bool{"A": true},
			expected:      "B",
		},
		{
			name:          "Ignored fields map is nil",
			inputObj:      SimpleStruct3{Name: ""},
			ignoredFields: nil,
			expected:      "Name",
		},

		// --- Private Field Case ---
		{
			name: "Private field is empty (should be found as IsExported is commented out)",
			inputObj: PrivateFieldStruct{
				Exported:   "not empty",
				unexported: "",
			},
			expected: "unexported",
		},
		{
			name: "Exported field is empty, private is not",
			inputObj: PrivateFieldStruct{
				Exported:   "",
				unexported: "not empty",
			},
			expected: "Exported",
		},

		// --- Edge Cases & Invalid Input ---
		{
			name:     "Input is a pointer to a struct",
			inputObj: &SimpleStruct3{Name: ""},
			expected: "Name",
		},
		{
			name:     "Input is a nil pointer to a struct",
			inputObj: (*SimpleStruct)(nil),
			expected: "",
		},
		{
			name:     "Input is nil",
			inputObj: nil,
			expected: "",
		},
		{
			name:     "Input is not a struct (int)",
			inputObj: 42,
			expected: "",
		},
		{
			name:     "Input is not a struct (string)",
			inputObj: "i am a string",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := FindEmptyStringField(tc.inputObj, tc.ignoredFields)
			if actual != tc.expected {
				t.Errorf("FindEmptyStringField() = %q, want %q", actual, tc.expected)
			}
		})
	}
}
