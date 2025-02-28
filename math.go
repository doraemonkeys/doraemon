package doraemon

import (
	"slices"
	"sync"
	"sync/atomic"

	"math/rand/v2"
)

// Prize represents a prize with a weight
type Prize[T comparable] struct {
	Item   T
	Weight int64
}

// WeightedRandom holds the prizes and their cumulative distribution
type WeightedRandom[T comparable] struct {
	prizesMu sync.RWMutex
	prizes   []Prize[T]
	// totalWeight is the total weight of the prizes
	totalWeight atomic.Int64
}

func NewWeightedRandom[T comparable]() *WeightedRandom[T] {
	return &WeightedRandom[T]{}
}

// AddPrize adds a new prize with a specified weight to the collection.
func (wr *WeightedRandom[T]) AddPrize(item T, weight int) {
	if weight <= 0 {
		panic("weight must be greater than 0")
	}
	wr.prizesMu.Lock()
	defer wr.prizesMu.Unlock()
	wr.prizes = append(wr.prizes, Prize[T]{Item: item, Weight: int64(weight)})
	wr.totalWeight.Add(int64(weight))
}

// RemovePrize removes a prize from the collection.
func (wr *WeightedRandom[T]) RemovePrize(item T) {
	wr.prizesMu.Lock()
	defer wr.prizesMu.Unlock()
	for i, prize := range wr.prizes {
		if prize.Item == item {
			wr.prizes = slices.Delete(wr.prizes, i, i+1)
			wr.totalWeight.Add(int64(-prize.Weight))
			break
		}
	}
}

// GetRandomPrize returns a random prize based on weights.
func (wr *WeightedRandom[T]) GetRandomPrize() Prize[T] {
	p, _ := wr.GetRandomPrizeWithTotalWeight()
	return p
}

// GetRandomPrizeWithTotalWeight returns a random prize and the total weight.
func (wr *WeightedRandom[T]) GetRandomPrizeWithTotalWeight() (Prize[T], int64) {
	wr.prizesMu.RLock()
	defer wr.prizesMu.RUnlock()
	var totalWeight int64 = wr.totalWeight.Load()
	return wr.getRandomPrizeFromSlice(wr.prizes, totalWeight), totalWeight
}

// GetTotalWeight returns the total weight of all prizes.
func (wr *WeightedRandom[T]) GetTotalWeight() int64 {
	wr.prizesMu.RLock()
	defer wr.prizesMu.RUnlock()
	var totalWeight int64
	for _, prize := range wr.prizes {
		totalWeight += prize.Weight
	}
	return totalWeight
}

// GetAllPrizes returns a copy of all prizes.
func (wr *WeightedRandom[T]) GetAllPrizes() []Prize[T] {
	var prizes []Prize[T] = make([]Prize[T], len(wr.prizes))
	wr.prizesMu.RLock()
	defer wr.prizesMu.RUnlock()
	copy(prizes, wr.prizes)
	return prizes
}

// GetRandomPrizeFromSlice selects a random prize from a slice based on weights.
// It panics if no prize is found.
func (wr *WeightedRandom[T]) GetRandomPrizeFromSlice(prizes []Prize[T]) Prize[T] {
	return wr.getRandomPrizeFromSlice(prizes)
}

func (wr *WeightedRandom[T]) getRandomPrizeFromSlice(prizes []Prize[T], totalWeight0 ...int64) Prize[T] {
	var totalWeight int64
	if len(totalWeight0) > 0 && totalWeight0[0] > 0 {
		totalWeight = totalWeight0[0]
	} else {
		for _, prize := range prizes {
			totalWeight += prize.Weight
		}
	}
	random := rand.Int64N(totalWeight)
	for _, prize := range prizes {
		if random < prize.Weight {
			return prize
		}
		random -= prize.Weight
	}
	panic("no prize found")
}

// for test
func (wr *WeightedRandom[T]) shuffle() {
	rand.Shuffle(len(wr.prizes), func(i, j int) {
		wr.prizes[i], wr.prizes[j] = wr.prizes[j], wr.prizes[i]
	})
	// if len(wr.prizes) > 1 && rand.Int32N(10) > 8 {
	// 	wr.prizes[0], wr.prizes[len(wr.prizes)-1] = wr.prizes[len(wr.prizes)-1], wr.prizes[0]
	// }
}
