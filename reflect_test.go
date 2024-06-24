package doraemon

import "testing"

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
