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
			if got := FindStructEmptyStringField(tt.s, tt.ignores); got != tt.want {
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
