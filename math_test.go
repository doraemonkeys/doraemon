package doraemon

import (
	"fmt"
	"math"
	"math/big"
	"testing"
)

// Test adding prizes and retrieving total weight
func TestAddPrizeAndGetTotalWeight(t *testing.T) {
	wr := NewWeightedRandom[string]()

	wr.AddPrize("Prize1", 10)
	wr.AddPrize("Prize2", 20)

	expectedTotalWeight := int64(30)
	if wr.GetTotalWeight() != expectedTotalWeight {
		t.Errorf("expected total weight %d, got %d", expectedTotalWeight, wr.GetTotalWeight())
	}
}

// Test retrieving all prizes
func TestGetAllPrizes(t *testing.T) {
	wr := NewWeightedRandom[string]()

	wr.AddPrize("Prize1", 10)
	wr.AddPrize("Prize2", 20)

	prizes := wr.GetAllPrizes()
	if len(prizes) != 2 {
		t.Errorf("expected 2 prizes, got %d", len(prizes))
	}

	if prizes[0].Item != "Prize1" || prizes[1].Item != "Prize2" {
		t.Errorf("prizes do not match expected values")
	}
}

// Test getting a random prize
func TestGetRandomPrize(t *testing.T) {
	wr := NewWeightedRandom[string]()

	wr.AddPrize("Prize1", 10)
	wr.AddPrize("Prize2", 20)

	// This test is probabilistic and may not be reliable in all cases.
	// It's better to test the distribution over many iterations.
	prize := wr.GetRandomPrize()
	if prize.Item != "Prize1" && prize.Item != "Prize2" {
		t.Errorf("unexpected prize: %s", prize.Item)
	}
}

// Test GetRandomPrizeFromSlice with no prizes
func TestGetRandomPrizeFromSliceNoPrizes(t *testing.T) {
	wr := NewWeightedRandom[string]()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic, but did not occur")
		}
	}()

	wr.GetRandomPrizeFromSlice(nil)
}

// Test GetRandomPrizeFromSlice with no prizes
func TestGetRandomPrizeFromSliceNoPrizes2(t *testing.T) {
	wr := NewWeightedRandom[string]()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic, but did not occur")
		}
	}()

	wr.GetRandomPrizeFromSlice([]Prize[string]{})
}
func TestGetRandomPrizeFromSliceNoPrizes3(t *testing.T) {
	wr := NewWeightedRandom[string]()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic, but did not occur")
		}
	}()

	wr.GetRandomPrizeFromSlice([]Prize[string]{{Item: "Prize1", Weight: 0}})
}

// Test concurrent access
func TestConcurrentAccess(t *testing.T) {
	wr := NewWeightedRandom[string]()

	// Add initial prize
	wr.AddPrize("Prize1", 10)

	done := make(chan bool)

	// Start concurrent readers
	for i := 0; i < 100; i++ {
		go func() {
			_ = wr.GetRandomPrize()
			done <- true
		}()
	}

	// Start concurrent writers
	for i := 0; i < 100; i++ {
		go func(i int) {
			wr.AddPrize(fmt.Sprintf("Prize%d", i), 10)
			done <- true
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < 200; i++ {
		<-done
	}

	// Check that total weight is as expected
	expectedWeight := int64(1010)
	if wr.GetTotalWeight() != expectedWeight {
		t.Errorf("expected total weight %d, got %d", expectedWeight, wr.GetTotalWeight())
	}
}

func TestMultipleDraws(t *testing.T) {
	wr := NewWeightedRandom[string]()

	// 添加奖品和权重
	wr.AddPrize("A", 10)
	wr.AddPrize("B", 30)
	wr.AddPrize("C", 60)
	wr.AddPrize("D", 1)
	wr.AddPrize("E", 9)

	totalWeight := wr.GetTotalWeight()
	drawCount := 10000000 // 抽奖次数

	// 记录每个奖品被抽中的次数
	results := make(map[string]int)

	// 进行抽奖
	for i := 0; i < drawCount; i++ {
		prize := wr.GetRandomPrize()
		results[prize.Item]++
	}

	// 检查结果
	for _, prize := range wr.GetAllPrizes() {
		expectedProb := float64(prize.Weight) / float64(totalWeight)
		actualProb := float64(results[prize.Item]) / float64(drawCount)

		// 计算误差百分比
		errorPercentage := math.Abs(expectedProb-actualProb) / expectedProb * 100

		fmt.Printf("Prize %s: Expected %.4f, Actual %.4f, Error %.2f%%\n",
			prize.Item, expectedProb, actualProb, errorPercentage)

		// 允许 1% 的误差
		if errorPercentage > 1 {
			t.Errorf("Prize %s: probability mismatch. Expected %.4f, got %.4f",
				prize.Item, expectedProb, actualProb)
		}
	}
}

func TestMultipleDraws2(t *testing.T) {
	wr := NewWeightedRandom[string]()

	// 添加奖品和权重
	wr.AddPrize("A", 10)
	wr.AddPrize("B", 30)
	wr.AddPrize("C", 60)
	wr.AddPrize("D", 1)
	wr.AddPrize("E", 9)

	totalWeight := wr.GetTotalWeight()
	drawCount := 10000000 // 抽奖次数

	// 记录每个奖品被抽中的次数
	results := make(map[string]int)

	// 进行抽奖
	for i := 0; i < drawCount; i++ {
		prize := wr.GetRandomPrize()
		results[prize.Item]++
		wr.shuffle()
	}

	// 检查结果
	for _, prize := range wr.GetAllPrizes() {
		expectedProb := float64(prize.Weight) / float64(totalWeight)
		actualProb := float64(results[prize.Item]) / float64(drawCount)

		// 计算误差百分比
		errorPercentage := math.Abs(expectedProb-actualProb) / expectedProb * 100

		fmt.Printf("Prize %s: Expected %.4f, Actual %.4f, Error %.2f%%\n",
			prize.Item, expectedProb, actualProb, errorPercentage)

		// 允许 1% 的误差
		if errorPercentage > 1 {
			t.Errorf("Prize %s: probability mismatch. Expected %.4f, got %.4f",
				prize.Item, expectedProb, actualProb)
		}
	}
}

func TestIsPowerOfTwoBig(t *testing.T) {
	bigOne := big.NewInt(1)
	bigEighty := new(big.Int).Lsh(bigOne, 80) // 2^80

	tests := []struct {
		name string
		k    *big.Int
		want bool
	}{
		{"zero", big.NewInt(0), false},
		{"negative one", big.NewInt(-1), false},
		{"negative power of two", big.NewInt(-4), false},
		{"one", big.NewInt(1), true},
		{"two", big.NewInt(2), true},
		{"three", big.NewInt(3), false},
		{"four", big.NewInt(4), true},
		{"five", big.NewInt(5), false},
		{"six", big.NewInt(6), false},
		{"eight", big.NewInt(8), true},
		{"1023 (2^10-1)", big.NewInt(1023), false},
		{"1024 (2^10)", big.NewInt(1024), true},
		{"large power of two (2^80)", bigEighty, true},
		{"large non-power of two (2^80 + 1)", new(big.Int).Add(bigEighty, bigOne), false},
		{"large non-power of two (2^80 - 1)", new(big.Int).Sub(bigEighty, bigOne), false},
		{"large non-power of two (2^80 + 2^79)", new(big.Int).Add(bigEighty, new(big.Int).Rsh(bigEighty, 1)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPowerOfTwoBig(tt.k); got != tt.want {
				t.Errorf("IsPowerOfTwoBig(%v) = %v, want %v", tt.k, got, tt.want)
			}
		})
	}
}

func TestDecomposeAsPowerOfTwoMinusOneShifted(t *testing.T) {
	// Helper to create 2^n - 1
	pow2nMinus1 := func(n uint) *big.Int {
		if n == 0 {
			return big.NewInt(0) // 2^0 - 1 = 0
		}
		one := big.NewInt(1)
		pow2n := new(big.Int).Lsh(one, n)
		return new(big.Int).Sub(pow2n, one)
	}

	tests := []struct {
		name   string
		num    *big.Int
		wantOk bool
		wantN  int
		wantM  int
	}{
		// Basic cases (m=0)
		{"num=0 (2^0-1)", big.NewInt(0), true, 0, 0},         // (2^0-1)<<0
		{"num=1 (2^1-1)", big.NewInt(1), true, 1, 0},         // (2^1-1)<<0
		{"num=3 (2^2-1)", big.NewInt(3), true, 2, 0},         // (2^2-1)<<0
		{"num=7 (2^3-1)", big.NewInt(7), true, 3, 0},         // (2^3-1)<<0
		{"num=15 (2^4-1)", big.NewInt(15), true, 4, 0},       // (2^4-1)<<0
		{"num=1023 (2^10-1)", big.NewInt(1023), true, 10, 0}, // (2^10-1)<<0
		{"large 2^60-1", pow2nMinus1(60), true, 60, 0},       // (2^60-1)<<0

		// Shifted cases (m>0)
		{"num=2 ((2^1-1)<<1)", big.NewInt(2), true, 1, 1},   // (1)<<1
		{"num=6 ((2^2-1)<<1)", big.NewInt(6), true, 2, 1},   // (3)<<1
		{"num=12 ((2^2-1)<<2)", big.NewInt(12), true, 2, 2}, // (3)<<2
		{"num=14 ((2^3-1)<<1)", big.NewInt(14), true, 3, 1}, // (7)<<1
		{"num=56 ((2^3-1)<<3)", big.NewInt(56), true, 3, 3}, // (7)<<3
		{"large (2^60-1)<<5", new(big.Int).Lsh(pow2nMinus1(60), 5), true, 60, 5},

		// Powers of two are (2^1-1) << m
		{"num=4 (power of two)", big.NewInt(4), true, 1, 2},        // (2^1-1)<<2 = 1<<2
		{"num=8 (power of two)", big.NewInt(8), true, 1, 3},        // (2^1-1)<<3 = 1<<3
		{"num=1024 (power of two)", big.NewInt(1024), true, 1, 10}, // (2^1-1)<<10 = 1<<10

		// Non-decomposable cases
		{"num=5", big.NewInt(5), false, 0, 0},
		{"num=9", big.NewInt(9), false, 0, 0},   // 2^n+1 form
		{"num=10", big.NewInt(10), false, 0, 0}, // 5 << 1, 5 is not 2^n-1
		{"num=11", big.NewInt(11), false, 0, 0},
		{"num=13", big.NewInt(13), false, 0, 0},
		{"num=17", big.NewInt(17), false, 0, 0}, // 2^n+1 form
		{"negative num=-1", big.NewInt(-1), false, 0, 0},
		{"negative num=-6", big.NewInt(-6), false, 0, 0},
		{"(2^60+1)", new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 60), big.NewInt(1)), false, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOk, gotN, gotM := DecomposeAsPowerOfTwoMinusOneShifted(tt.num)
			if gotOk != tt.wantOk || gotN != tt.wantN || gotM != tt.wantM {
				t.Errorf("DecomposeAsPowerOfTwoMinusOneShifted(%v) = (%v, %v, %v), want (%v, %v, %v)",
					tt.num, gotOk, gotN, gotM, tt.wantOk, tt.wantN, tt.wantM)
			}
		})
	}
}

func TestDecomposeAsPowerOfTwoPlusOneShifted(t *testing.T) {
	// Helper to create 2^n + 1
	pow2nPlus1 := func(n uint) *big.Int {
		one := big.NewInt(1)
		pow2n := new(big.Int).Lsh(one, n)
		return new(big.Int).Add(pow2n, one)
	}

	tests := []struct {
		name   string
		num    *big.Int
		wantOk bool
		wantN  int
		wantM  int
	}{
		// Basic cases (m=0)
		{"num=2 (2^0+1)", big.NewInt(2), true, 0, 0},         // (2^0+1)<<0
		{"num=3 (2^1+1)", big.NewInt(3), true, 1, 0},         // (2^1+1)<<0
		{"num=5 (2^2+1)", big.NewInt(5), true, 2, 0},         // (2^2+1)<<0
		{"num=9 (2^3+1)", big.NewInt(9), true, 3, 0},         // (2^3+1)<<0
		{"num=17 (2^4+1)", big.NewInt(17), true, 4, 0},       // (2^4+1)<<0
		{"num=1025 (2^10+1)", big.NewInt(1025), true, 10, 0}, // (2^10+1)<<0
		{"large 2^60+1", pow2nPlus1(60), true, 60, 0},        // (2^60+1)<<0

		// Shifted cases (m>0)
		{"num=4 ((2^0+1)<<1)", big.NewInt(4), true, 0, 1},   // (2)<<1
		{"num=6 ((2^1+1)<<1)", big.NewInt(6), true, 1, 1},   // (3)<<1
		{"num=10 ((2^2+1)<<1)", big.NewInt(10), true, 2, 1}, // (5)<<1
		{"num=20 ((2^2+1)<<2)", big.NewInt(20), true, 2, 2}, // (5)<<2
		{"num=18 ((2^3+1)<<1)", big.NewInt(18), true, 3, 1}, // (9)<<1
		{"num=72 ((2^3+1)<<3)", big.NewInt(72), true, 3, 3}, // (9)<<3
		{"large (2^60+1)<<5", new(big.Int).Lsh(pow2nPlus1(60), 5), true, 60, 5},

		// Non-decomposable cases
		{"num=0", big.NewInt(0), false, 0, 0},
		{"num=1", big.NewInt(1), false, 0, 0},
		{"num=7", big.NewInt(7), false, 0, 0}, // 2^n-1 form
		{"num=11", big.NewInt(11), false, 0, 0},
		{"num=13", big.NewInt(13), false, 0, 0},
		{"num=15", big.NewInt(15), false, 0, 0}, // 2^n-1 form
		{"negative num=-1", big.NewInt(-1), false, 0, 0},
		{"negative num=-10", big.NewInt(-10), false, 0, 0},
		{"(2^60-1)", new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 60), big.NewInt(1)), false, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOk, gotN, gotM := DecomposeAsPowerOfTwoPlusOneShifted(tt.num)
			if gotOk != tt.wantOk || gotN != tt.wantN || gotM != tt.wantM {
				t.Errorf("DecomposeAsPowerOfTwoPlusOneShifted(%v) = (%v, %v, %v), want (%v, %v, %v)",
					tt.num, gotOk, gotN, gotM, tt.wantOk, tt.wantN, tt.wantM)
			}
		})
	}
}

// --- Tests ---

func TestDecomposeAsPowerOfTwoMinusOneShifted2(t *testing.T) {
	tests := []struct {
		name   string
		numStr string
		wantOk bool
		wantN  int
		wantM  int
	}{
		// Edge Cases
		{"zero", "0", true, 0, 0},
		{"negative", "-10", false, 0, 0},
		// (2^n - 1) forms
		{"1 (2^1-1)", "1", true, 1, 0},         // (2^1-1) << 0 = 1
		{"3 (2^2-1)", "3", true, 2, 0},         // (2^2-1) << 0 = 3
		{"7 (2^3-1)", "7", true, 3, 0},         // (2^3-1) << 0 = 7
		{"15 (2^4-1)", "15", true, 4, 0},       // (2^4-1) << 0 = 15
		{"1023 (2^10-1)", "1023", true, 10, 0}, // (2^10-1) << 0 = 1023
		// Shifted (2^n - 1) forms
		{"2 (1<<1)", "2", true, 1, 1},                                        // (2^1-1) << 1 = 1 << 1 = 2
		{"6 (3<<1)", "6", true, 2, 1},                                        // (2^2-1) << 1 = 3 << 1 = 6
		{"12 (3<<2)", "12", true, 2, 2},                                      // (2^2-1) << 2 = 3 << 2 = 12
		{"14 (7<<1)", "14", true, 3, 1},                                      // (2^3-1) << 1 = 7 << 1 = 14
		{"56 (7<<3)", "56", true, 3, 3},                                      // (2^3-1) << 3 = 7 << 3 = 56
		{"1111111111111111111 (2^60-1)", "1152921504606846975", true, 60, 0}, // 2^60-1
		{"(2^3-1)<<70", new(big.Int).Lsh(big.NewInt(7), 70).String(), true, 3, 70},
		{"(2^1-1)<<65 (power of 2)", new(big.Int).Lsh(big.NewInt(1), 65).String(), true, 1, 65}, // 2^65 = (2^1-1) << 65

		// Let's correct the above "not form 4"
		{"form 4 is (2^1-1)<<2", "4", true, 1, 2},
		{"not form 5", "5", false, 0, 0},
		{"not form 9", "9", false, 0, 0},
		{"not form 10 (5<<1)", "10", false, 0, 0},
		{"not form 17", "17", false, 0, 0},
		{"large not form (2^60)", new(big.Int).Lsh(big.NewInt(1), 60).String(), true, 1, 60}, // 2^60 = (2^1-1) << 60
		{"large not form (2^60+1)", new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 60), big.NewInt(1)).String(), false, 0, 0},
		{"large not form (3 * 2^50)", new(big.Int).Lsh(big.NewInt(3), 50).String(), true, 2, 50}, // (2^2-1) << 50
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num := new(big.Int)
			_, okParse := num.SetString(tt.numStr, 10)
			if !okParse {
				t.Fatalf("Failed to parse numStr: %s", tt.numStr)
			}

			gotOk, gotN, gotM := DecomposeAsPowerOfTwoMinusOneShifted(num)

			if gotOk != tt.wantOk {
				t.Errorf("DecomposeAsPowerOfTwoMinusOneShifted() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotOk { // Only check n and m if ok is true and expected to be true
				if gotN != tt.wantN {
					t.Errorf("DecomposeAsPowerOfTwoMinusOneShifted() gotN = %v, want %v", gotN, tt.wantN)
				}
				if gotM != tt.wantM {
					t.Errorf("DecomposeAsPowerOfTwoMinusOneShifted() gotM = %v, want %v", gotM, tt.wantM)
				}
			}
		})
	}
}

func TestDecomposeAsPowerOfTwoPlusOneShifted2(t *testing.T) {
	tests := []struct {
		name   string
		numStr string
		wantOk bool
		wantN  int
		wantM  int
	}{
		// Edge Cases
		{"zero", "0", false, 0, 0},
		{"negative", "-10", false, 0, 0},
		{"one (too small)", "1", false, 0, 0},
		// (2^n + 1) forms
		{"2 (2^0+1)", "2", true, 0, 0},         // (2^0+1) << 0 = 2
		{"3 (2^1+1)", "3", true, 1, 0},         // (2^1+1) << 0 = 3
		{"5 (2^2+1)", "5", true, 2, 0},         // (2^2+1) << 0 = 5
		{"9 (2^3+1)", "9", true, 3, 0},         // (2^3+1) << 0 = 9
		{"17 (2^4+1)", "17", true, 4, 0},       // (2^4+1) << 0 = 17
		{"1025 (2^10+1)", "1025", true, 10, 0}, // (2^10+1) << 0 = 1025
		// Shifted (2^n + 1) forms
		{"4 (2<<1)", "4", true, 0, 1},   // (2^0+1) << 1 = 2 << 1 = 4
		{"6 (3<<1)", "6", true, 1, 1},   // (2^1+1) << 1 = 3 << 1 = 6
		{"10 (5<<1)", "10", true, 2, 1}, // (2^2+1) << 1 = 5 << 1 = 10
		{"12 (3<<2)", "12", true, 1, 2}, // (2^1+1) << 2 = 3 << 2 = 12
		{"18 (9<<1)", "18", true, 3, 1}, // (2^3+1) << 1 = 9 << 1 = 18
		{"40 (5<<3)", "40", true, 2, 3}, // (2^2+1) << 3 = 5 << 3 = 40
		{"(2^60+1)", new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 60), big.NewInt(1)).String(), true, 60, 0},
		{"(2^3+1)<<70", new(big.Int).Lsh(big.NewInt(9), 70).String(), true, 3, 70},
		// Not of the form
		{"not form 7", "7", false, 0, 0},
		{"not form 11", "11", false, 0, 0},
		{"not form 13", "13", false, 0, 0},
		{"not form 14 (7<<1)", "14", false, 0, 0},
		{"not form (2^60-1)", new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 60), big.NewInt(1)).String(), false, 0, 0},
		{"not form (3 * 2^50)", new(big.Int).Lsh(big.NewInt(3), 50).String(), true, 1, 50}, // (2^1+1) << 50
		{"power of 2 (2^5)", "32", true, 0, 4}, // 32 = 2 << 4 = (2^0+1) << 4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num := new(big.Int)
			_, okParse := num.SetString(tt.numStr, 10)
			if !okParse {
				t.Fatalf("Failed to parse numStr: %s", tt.numStr)
			}

			gotOk, gotN, gotM := DecomposeAsPowerOfTwoPlusOneShifted(num)

			if gotOk != tt.wantOk {
				t.Errorf("DecomposeAsPowerOfTwoPlusOneShifted() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotOk { // Only check n and m if ok is true and expected to be true
				if gotN != tt.wantN {
					t.Errorf("DecomposeAsPowerOfTwoPlusOneShifted() gotN = %v, want %v", gotN, tt.wantN)
				}
				if gotM != tt.wantM {
					t.Errorf("DecomposeAsPowerOfTwoPlusOneShifted() gotM = %v, want %v", gotM, tt.wantM)
				}
			}
		})
	}
}

// Test for IsPowerOfTwoBig (helper, but good to have some tests)
func TestIsPowerOfTwoBig2(t *testing.T) {
	tests := []struct {
		name   string
		numStr string
		want   bool
	}{
		{"0", "0", false},
		{"1", "1", true}, // 2^0
		{"2", "2", true}, // 2^1
		{"3", "3", false},
		{"4", "4", true}, // 2^2
		{"6", "6", false},
		{"8", "8", true},       // 2^3
		{"1024", "1024", true}, // 2^10
		{"1023", "1023", false},
		{"negative", "-4", false},
		{"large power of two (2^100)", new(big.Int).Lsh(big.NewInt(1), 100).String(), true},
		{"large non-power of two (2^100+1)", new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 100), big.NewInt(1)).String(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var num *big.Int
			if tt.name == "nil" {
				num = nil
			} else {
				num = new(big.Int)
				_, okParse := num.SetString(tt.numStr, 10)
				if !okParse {
					t.Fatalf("Failed to parse numStr for IsPowerOfTwoBig: %s", tt.numStr)
				}
			}

			if got := IsPowerOfTwoBig(num); got != tt.want {
				t.Errorf("IsPowerOfTwoBig(%s) = %v, want %v", tt.numStr, got, tt.want)
			}
		})
	}
}

func TestDecomposeAsPowerOfTwoMinusOneShifted3(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		expected bool
		n        int
		m        int
	}{
		{"zero", big.NewInt(0), true, 0, 0},
		{"negative", big.NewInt(-1), false, 0, 0},
		{"one", big.NewInt(1), true, 1, 0},                  // 2^1-1 = 1
		{"three", big.NewInt(3), true, 2, 0},                // 2^2-1 = 3
		{"six", big.NewInt(6), true, 2, 1},                  // (2^2-1)<<1 = 3<<1 = 6
		{"seven", big.NewInt(7), true, 3, 0},                // 2^3-1 = 7
		{"fourteen", big.NewInt(14), true, 3, 1},            // (2^3-1)<<1 = 7<<1 = 14
		{"fifteen", big.NewInt(15), true, 4, 0},             // 2^4-1 = 15
		{"thirty", big.NewInt(30), true, 4, 1},              // (2^4-1)<<1 = 15<<1 = 30
		{"sixty", big.NewInt(60), true, 4, 2},               // (2^4-1)<<2 = 15<<2 = 60
		{"large_num", big.NewInt(0xFFFF0000), true, 16, 16}, // (2^16-1)<<16
		{"not_decomposable", big.NewInt(5), false, 0, 0},
		{"not_decomposable_2", big.NewInt(9), false, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, n, m := DecomposeAsPowerOfTwoMinusOneShifted(tt.input)
			if ok != tt.expected || n != tt.n || m != tt.m {
				t.Errorf("DecomposeAsPowerOfTwoMinusOneShifted(%v) = (%v, %v, %v), want (%v, %v, %v)",
					tt.input, ok, n, m, tt.expected, tt.n, tt.m)
			}
		})
	}
}

func TestDecomposeAsPowerOfTwoPlusOneShifted3(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		expected bool
		n        int
		m        int
	}{
		{"zero", big.NewInt(0), false, 0, 0},
		{"negative", big.NewInt(-1), false, 0, 0},
		{"two", big.NewInt(2), true, 0, 0},                   // 2^0+1 = 2
		{"three", big.NewInt(3), true, 1, 0},                 // 2^1+1 = 3
		{"five", big.NewInt(5), true, 2, 0},                  // 2^2+1 = 5
		{"six", big.NewInt(6), true, 1, 1},                   // (2^1+1)<<1 = 3<<1 = 6
		{"nine", big.NewInt(9), true, 3, 0},                  // 2^3+1 = 9
		{"ten", big.NewInt(10), true, 2, 1},                  // (2^2+1)<<1 = 5<<1 = 10
		{"seventeen", big.NewInt(17), true, 4, 0},            // 2^4+1 = 17
		{"thirty_four", big.NewInt(34), true, 4, 1},          // (2^4+1)<<1 = 17<<1 = 34
		{"large_num", big.NewInt(0x100010000), true, 16, 16}, // (2^16+1)<<16
		{"not_decomposable", big.NewInt(7), false, 0, 0},
		{"not_decomposable_2", big.NewInt(13), false, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, n, m := DecomposeAsPowerOfTwoPlusOneShifted(tt.input)
			if ok != tt.expected || n != tt.n || m != tt.m {
				t.Errorf("DecomposeAsPowerOfTwoPlusOneShifted(%v) = (%v, %v, %v), want (%v, %v, %v)",
					tt.input, ok, n, m, tt.expected, tt.n, tt.m)
			}
		})
	}
}

func TestEdgeCases3(t *testing.T) {
	// Test very large numbers
	largeNum := new(big.Int).Lsh(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)), 64)
	ok, n, m := DecomposeAsPowerOfTwoMinusOneShifted(largeNum)
	if !ok || n != 128 || m != 64 {
		t.Errorf("Failed to decompose large number (2^128-1)<<64")
	}

	largeNum2 := new(big.Int).Lsh(new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)), 32)
	ok, n, m = DecomposeAsPowerOfTwoPlusOneShifted(largeNum2)
	if !ok || n != 128 || m != 32 {
		t.Errorf("Failed to decompose large number (2^128+1)<<32")
	}

	// Test numbers that are just below/above decomposable forms
	almostPower := new(big.Int).Sub(new(big.Int).Lsh(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(1)), 8), big.NewInt(1))
	ok, _, _ = DecomposeAsPowerOfTwoMinusOneShifted(almostPower)
	if ok {
		t.Errorf("Should not decompose number that is (2^n-1)<<m - 1")
	}

	almostPower2 := new(big.Int).Add(new(big.Int).Lsh(new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(1)), 8), big.NewInt(1))
	ok, _, _ = DecomposeAsPowerOfTwoPlusOneShifted(almostPower2)
	if ok {
		t.Errorf("Should not decompose number that is (2^n+1)<<m + 1")
	}
}

func TestFormatAsPowerOfTwoMinusOneShiftedBig(t *testing.T) {

	bigs := []string{
		"170141183460469231731687303715884105728",
		"47890485652059026823698344598447161988085597568237568",
		"115792089237316195423570985008687907853269984665640564039457584007913129639935",
		"6277101735386680763835789423207666416102355444464034512896",
		"24519928653854221733733552434404946937899825954937634944",
		"24519928653854221733733552434404946937899825954937634688",
		"65536",
		"12",
	}

	bigInts := make([]*big.Int, len(bigs))
	for i, s := range bigs {
		bigInts[i] = new(big.Int)
		_, ok := bigInts[i].SetString(s, 10)
		if !ok {
			t.Fatalf("Failed to parse bigInt: %s", s)
		}
	}

	tests := []struct {
		name  string
		num   *big.Int
		want  bool
		want1 string
	}{
		{"0", big.NewInt(0), true, "0"},
		{"1", big.NewInt(1), true, "1"},
		{"2", big.NewInt(2), true, "2"},
		{"3", big.NewInt(3), true, "1<<2 - 1"},
		{"4", big.NewInt(4), true, "1 << 2"},
		{"5", big.NewInt(5), false, ""},

		{"170141183460469231731687303715884105728", bigInts[0], true, "1 << 127"},
		{"47890485652059026823698344598447161988085597568237568", bigInts[1], true, "1 << 175"},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935", bigInts[2], true, "1<<256 - 1"},
		{"6277101735386680763835789423207666416102355444464034512896", bigInts[3], true, "1 << 192"},
		{"24519928653854221733733552434404946937899825954937634944", bigInts[4], false, ""},
		{"24519928653854221733733552434404946937899825954937634688", bigInts[5], true, "(1<<177 - 1) << 7"},
		{"65536", bigInts[6], true, "1 << 16"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := FormatAsPowerOfTwoMinusOneShiftedBig(tt.num)
			if got != tt.want {
				t.Errorf("FormatAsPowerOfTwoMinusOneShiftedBig() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("FormatAsPowerOfTwoMinusOneShiftedBig() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestFormatAsPowerOfTwoPlusOneShiftedBig(t *testing.T) {
	bigs := []string{
		"170141183460469231731687303715884105728",
		"47890485652059026823698344598447161988085597568237568",
		"115792089237316195423570985008687907853269984665640564039457584007913129639935",
		"6277101735386680763835789423207666416102355444464034512896",
		"24519928653854221733733552434404946937899825954937634944",
		"24519928653854221733733552434404946937899825954937634688",
		"65537",
		"12",
	}

	bigInts := make([]*big.Int, len(bigs))
	for i, s := range bigs {
		bigInts[i] = new(big.Int)
		_, ok := bigInts[i].SetString(s, 10)
		if !ok {
			t.Fatalf("Failed to parse bigInt: %s", s)
		}
	}

	tests := []struct {
		name  string
		num   *big.Int
		want  bool
		want1 string
	}{

		{"170141183460469231731687303715884105728", bigInts[0], true, "2 << 126"},
		{"47890485652059026823698344598447161988085597568237568", bigInts[1], true, "2 << 174"},
		{"115792089237316195423570985008687907853269984665640564039457584007913129639935", bigInts[2], false, ""},
		{"6277101735386680763835789423207666416102355444464034512896", bigInts[3], true, "2 << 191"},
		{"24519928653854221733733552434404946937899825954937634944", bigInts[4], true, "(1<<177 + 1) << 7"},
		{"24519928653854221733733552434404946937899825954937634688", bigInts[5], false, ""},
		{"65537", bigInts[6], true, "1<<16 + 1"},
		{"12", bigInts[7], true, "(1<<1 + 1) << 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := FormatAsPowerOfTwoPlusOneShiftedBig(tt.num)
			if got != tt.want {
				t.Errorf("FormatAsPowerOfTwoPlusOneShiftedBig(%s) got = %v, want %v", tt.name, got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("FormatAsPowerOfTwoPlusOneShiftedBig(%s) got1 = %v, want %v", tt.name, got1, tt.want1)
			}
		})
	}

}
