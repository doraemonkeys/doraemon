package doraemon

import (
	"fmt"
	"math"
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
