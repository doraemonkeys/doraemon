package study

import "testing"

func TestCalculateSelector(t *testing.T) {
	tests := []struct {
		name      string
		signature string
		want      string
	}{
		{
			name:      "test1",
			signature: "transfer(address,uint256)",
			want:      "0xa9059cbb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateSelector(tt.signature); got != tt.want {
				t.Errorf("CalculateSelector() = %v, want %v", got, tt.want)
			}
		})
	}
}
