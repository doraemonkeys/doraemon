package doraemon

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"math/big"
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

// IsPowerOfTwoBig checks if a given positive *big.Int k is a power of two.
// A number is a power of two if it's greater than 0 and has only one bit set
// in its binary representation. This means k > 0 and (k & (k-1)) == 0.
func IsPowerOfTwoBig(k *big.Int) bool {
	// k must be > 0
	if k.Sign() != 1 { // Sign() returns -1 for <0, 0 for 0, 1 for >0
		return false
	}

	// A power of two 'k' has only one bit set. 'k-1' has all bits below that set to 1.
	// So 'k & (k-1)' will be zero if k is a power of two.
	// Example: k=4 (100), k-1=3 (011). 100 & 011 = 0.
	// Example: k=1 (001), k-1=0 (000). 001 & 000 = 0.

	// Calculate k-1
	one := big.NewInt(1)
	kMinus1 := new(big.Int).Sub(k, one)

	// Check k & (k-1) == 0
	resultAnd := new(big.Int).And(k, kMinus1)
	zero := big.NewInt(0)

	return resultAnd.Cmp(zero) == 0
}

// DecomposeAsPowerOfTwoMinusOneShifted attempts to find non-negative integers n and m
// such that num = (2^n - 1) << m.
// It returns (ok bool, n int, m int).
// n and m are standard integers as they represent bit counts or powers,
// which are unlikely to exceed standard integer limits even for very large nums.
func DecomposeAsPowerOfTwoMinusOneShifted(num *big.Int) (ok bool, n int, m int) {
	zero := big.NewInt(0)
	one := big.NewInt(1)

	// Handle num == 0 case: (2^0 - 1) << 0 = 0 << 0 = 0. So n=0, m=0.
	if num.Cmp(zero) == 0 {
		return true, 0, 0
	}

	// (2^n - 1) << m is always non-negative.
	if num.Sign() < 0 { // num < 0
		return false, 0, 0
	}

	// For num > 0, (2^n - 1) must be positive, so n >= 1.
	// (2^n - 1) is always odd for n >= 1.
	// We need to find 'm' such that num / (2^m) is of the form (2^n - 1).
	// num / (2^m) is the odd part of num.
	// 'm' is the number of trailing zeros in num's binary representation.

	mVal := int(num.TrailingZeroBits()) // TrailingZeroBits returns uint, cast to int

	// baseVal = num >> mVal
	baseVal := new(big.Int).Rsh(num, uint(mVal))

	// Now, baseVal must be of the form 2^n - 1.
	// So, baseVal + 1 must be 2^n (a power of two).
	// Since num > 0, baseVal must be > 0. So baseVal + 1 >= 2.
	potentialPowerOfTwo := new(big.Int).Add(baseVal, one)

	if IsPowerOfTwoBig(potentialPowerOfTwo) {
		// If potentialPowerOfTwo = 2^n, then n = log2(potentialPowerOfTwo).
		// For positive integers, n = (potentialPowerOfTwo.BitLen() - 1).
		// BitLen() returns 0 if potentialPowerOfTwo is 0, but it's >= 2 here.
		nVal := potentialPowerOfTwo.BitLen() - 1
		return true, nVal, mVal
	}

	return false, 0, 0
}

// DecomposeAsPowerOfTwoPlusOneShifted attempts to find non-negative integers n and m
// such that num = (2^n + 1) << m.
func DecomposeAsPowerOfTwoPlusOneShifted(num *big.Int) (ok bool, n int, m int) {
	zero := big.NewInt(0)
	one := big.NewInt(1)
	two := big.NewInt(2)

	// (2^n + 1) is always positive (>= 2 since n >= 0).
	// Shifting a positive number results in a positive number.
	// Thus, num must be positive.
	if num.Cmp(zero) <= 0 { // num <= 0
		return false, 0, 0
	}

	currentVal := new(big.Int).Set(num) // Make a copy to modify
	mVal := 0

	for {
		// The smallest value for (2^n + 1) is 2 (when n=0).
		// If currentVal drops below this, no solution is possible with further shifts.
		if currentVal.Cmp(two) < 0 {
			return false, 0, 0
		}

		// We need to check if currentVal can be expressed as 2^n + 1.
		// If currentVal = 2^n + 1, then currentVal - 1 = 2^n.
		// So, currentVal - 1 must be a power of two.
		potentialPowerOfTwo := new(big.Int).Sub(currentVal, one)

		if IsPowerOfTwoBig(potentialPowerOfTwo) {
			// If potentialPowerOfTwo is 2^n, then n = log2(potentialPowerOfTwo).
			// n = potentialPowerOfTwo.BitLen() - 1.
			// If potentialPowerOfTwo is 1 (e.g. currentVal was 2), BitLen() is 1, so n = 0. This is correct (2^0+1 = 2).
			// isPowerOfTwoBig ensures potentialPowerOfTwo is > 0.
			nVal := potentialPowerOfTwo.BitLen() - 1
			return true, nVal, mVal
		}

		// If currentVal is not a solution yet for the current 'mVal':
		// If currentVal is odd, we cannot shift it right further by dividing by 2
		// to find a smaller base for (2^n + 1). So, no solution possible by increasing mVal.
		// currentVal % 2 != 0  is  currentVal.Bit(0) == 1
		if currentVal.Bit(0) == 1 { // Check if the LSB is 1 (odd)
			return false, 0, 0
		}

		// If currentVal is even and not a solution yet,
		// shift it right by 1 (effectively incrementing mVal) and try again.
		currentVal.Rsh(currentVal, 1)
		mVal++
	}
}

// FormatAsPowerOfTwoMinusOneShiftedBig formats the result from DecomposeAsPowerOfTwoMinusOneShifted.
func FormatAsPowerOfTwoMinusOneShiftedBig(num *big.Int) (bool, string) {
	ok, n, m := DecomposeAsPowerOfTwoMinusOneShifted(num)
	if ok {
		// return fmt.Sprintf("(2^%d - 1) << %d", n, m)
		if n == 0 {
			return true, "0"
		}
		left := fmt.Sprintf("(1<<%d - 1)", n)
		if n == 1 {
			left = "1"
			if m == 1 {
				return true, "2"
			}
		}
		if m == 0 {
			return true, strings.Trim(left, "()")
		}
		return true, fmt.Sprintf("%s << %d", left, m)
	}
	return false, ""
}

// FormatAsPowerOfTwoPlusOneShiftedBig formats the result from DecomposeAsPowerOfTwoPlusOneShifted.
func FormatAsPowerOfTwoPlusOneShiftedBig(num *big.Int) (ok bool, result string) {
	ok, n, m := DecomposeAsPowerOfTwoPlusOneShifted(num)
	if ok {
		// return fmt.Sprintf("(2^%d + 1) << %d", n, m)
		left := fmt.Sprintf("(1<<%d + 1)", n)
		if m == 0 {
			return true, strings.Trim(left, "()")
		}
		if n == 0 {
			return true, fmt.Sprintf("1 << %d", m+1)
		}
		return true, fmt.Sprintf("%s << %d", left, m)
	}
	return false, ""
}
